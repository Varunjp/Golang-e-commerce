package admin

import (
	db "first-project/DB"
	"first-project/models"
	"first-project/utils"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WalletTransactions(c *gin.Context){

	pageStr := c.DefaultQuery("page","1")
	limitStr := c.DefaultQuery("limit","10")

	page, err := strconv.Atoi(pageStr)

	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1{
		limit = 10
	}

	offset := (page - 1) * limit
	var total int64

	
	var WalletTransactions []models.WalletTransaction
	type response struct{
		ID 			uint
		UserName 	string
		Type 		string
		Amount 		float64
		Description string 
		CreatedAt 	time.Time 
	}

	dbOrder := db.Db.Model(&models.WalletTransaction{})

	dbOrder.Count(&total)

	if err := db.Db.Order("id DESC").Limit(limit).Offset(offset).Find(&WalletTransactions).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Could not transaction details, please try again later"})
		return 
	}

	ResponseTransactions := make([]response,len(WalletTransactions))

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	pageRange := utils.GetPaginationPages(page,totalPages)


	for i, transaction := range WalletTransactions {
		var user models.User
		db.Db.Where("id = ?",transaction.UserID).First(&user)
		ResponseTransactions[i] = response{
			ID: transaction.ID,
			UserName: user.Username,
			Type: transaction.Type,
			Amount: math.Abs(transaction.Amount),
			Description: transaction.Description,
			CreatedAt: transaction.CreatedAt,
		}
	}

	c.HTML(http.StatusOK,"wallet.html",gin.H{
		"transactions":ResponseTransactions,
		"page":page,
		"totalPages":totalPages,
		"PageRange":pageRange,
		"limit":limit,
		})

}

func WalletRefunds(c *gin.Context){

	var walletTransactions []models.WalletTransaction
	
	type response struct{
		ID 			uint
		UserName 	string
		OrderID 	uint
		Amount 		float64
		Description string 
		CreatedAt 	time.Time 
	}

	

	if err := db.Db.Where("refund_status = ?",true).Find(&walletTransactions).Order("id DESC").Error; err != nil{
		if err != gorm.ErrRecordNotFound {
			c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to fetch wallet transcations, please try again later"})
			return 
		}
	}


	if len(walletTransactions) > 0 {

		ResponseTransactions := make([]response,len(walletTransactions))

		for i, transaction := range walletTransactions {
			var user models.User
			db.Db.Where("id = ?",transaction.UserID).First(&user)
			ResponseTransactions[i] = response{
				ID: transaction.ID,
				UserName: user.Username,
				OrderID: transaction.OrderID,
				Amount: math.Abs(transaction.Amount),
				Description: transaction.Description,
				CreatedAt: transaction.CreatedAt,
			}
		}

		c.HTML(http.StatusOK,"refund_request.html",gin.H{"refundRequests":ResponseTransactions})

	}else{
		c.HTML(http.StatusNotFound,"refund_request.html",gin.H{"message":"No refund transactions"})
	}

}

func WalletRefundApproval (c *gin.Context){
	
	transactionId := c.PostForm("request_id")
	reason := c.PostForm("note")
	var transaction models.WalletTransaction

	if err := db.Db.Where("id = ?",transactionId).First(&transaction).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to get details for specific transcation, please try again later"})
		return 
	}

	if err := db.Db.Model(&models.Wallet{}).Where("user_id = ?",transaction.UserID).Update("balance",gorm.Expr("balance + ?",math.Abs(transaction.Amount))).Error; err != nil {
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to update user wallet, please try again later"})
		return
	}

	transaction.RefundStatus = false
	transaction.Description = reason
	transaction.Type = "Credit"

	if err := db.Db.Save(&transaction).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to save transaction, please try again later"})
		return
	}

	var order models.Order
	var orderItem models.OrderItem

	if transaction.OrderID != 0 {
		
		if err := db.Db.Where("id = ?",transaction.OrderID).First(&order).Error; err != nil{
			c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to retrive order details"})
			return 
		}

		if transaction.OrderItemID != 0{

			if err := db.Db.Unscoped().Where("id = ?",transaction.OrderItemID).First(&orderItem).Error; err != nil{
				c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to get order item details"})
				return
			}

			var product models.Product_Variant
			db.Db.Where("id = ?",orderItem.ProductID).First(&product)
			itemTotal := orderItem.Price * float64(orderItem.Quantity) + product.Tax * float64(orderItem.Quantity)
			updatedTotal := order.SubTotal - itemTotal

			var usedCoupon models.UsedCoupon
			var coupon models.Coupons

			if err := db.Db.Where("order_id = ? AND user_id = ?",order.ID,order.UserID).First(&usedCoupon).Error; err != nil{
				log.Println(err)
			}

			if err := db.Db.Where("id = ?",usedCoupon.CouponID).First(&coupon).Error; err != nil{
				log.Println(err)
			}


			newDiscount := order.DiscountTotal - transaction.Amount

			if coupon.ID != 0 {
				if coupon.MaxAmount > updatedTotal{
					if newDiscount < 0 {
						order.DiscountTotal = 0.0
					}else{
						order.DiscountTotal = newDiscount
					}
				}
			}else{
				if newDiscount <= 0 {
					order.DiscountTotal = 0.0
				}else{
					order.DiscountTotal = newDiscount
				}
			}

			orderItem.PaymentStatus = "Refunded"
			db.Db.Save(&orderItem)

		}

		db.Db.Save(&order)
	}
	

	c.Redirect(http.StatusSeeOther,"/admin/wallet-transactions")
}


func WalletRefundDecline (c *gin.Context){
	
	transactionId := c.PostForm("request_id")
	reason := c.PostForm("note")
	var transaction models.WalletTransaction

	if err := db.Db.Where("id = ?",transactionId).First(&transaction).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to get details for specific transcation, please try again later"})
		return 
	}

	transaction.RefundStatus = false
	transaction.Description = reason
	transaction.Type = "Refund declined"

	if err := db.Db.Save(&transaction).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to save transaction, please try again later"})
		return
	}

	if transaction.OrderID != 0 {
		var orderitem models.OrderItem
		
		if transaction.OrderItemID != 0 {
			if err := db.Db.Unscoped().Where("id = ?",transaction.OrderItemID).First(&orderitem).Error; err != nil{
				c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to load order item details"})
				return 
			}
		}
		

		orderitem.PaymentStatus = "Refund rejected"
		db.Db.Save(&orderitem)
	}

	c.Redirect(http.StatusSeeOther,"/admin/wallet-transactions")
}