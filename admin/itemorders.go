package admin

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"

	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminItemCancel(c *gin.Context) {

	var orderItem models.OrderItem
	var order models.Order
	itemId := c.Param("id")

	if err := db.Db.Where("id = ?",itemId).First(&orderItem).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_orderDetails.html",gin.H{"error":"Failed to retrieve order item"})
		return 
	}

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderItem.OrderID).First(&order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_orderDetails.html",gin.H{"error":"Failed to retrieve order"})
		return 
	}

	checkRemaing := 0

	for _,item := range order.OrderItems{
		if item.Status == "Pending" || item.Status == "Processing" || item.Status == "Returned"{
			checkRemaing ++
		}
	}

	if checkRemaing == 0{
		order.Status = "Delivered"
		order.PaymentStatus = "Successful"
		db.Db.Save(&order)
	}

	if order.PaymentMethod == "cod" && order.Status != "Delivered" {
		newTotal := order.TotalAmount + (orderItem.Price * float64(orderItem.Quantity))
		order.TotalAmount = newTotal
	}
	
	if order.Status == "Delivered"{
		orderItem.Status = "Delivered non returnable"
		order.Status = "Delivered non returnable"
	}else{
		orderItem.Status = "Cancel rejected"
	}
	

	orderId := orderItem.OrderID
	orderIdStr := strconv.Itoa(int(orderId))
	db.Db.Save(&orderItem)
	db.Db.Save(&order)
	// db.Db.Delete(&orderItem)

	c.Redirect(http.StatusSeeOther,"/admin/order/"+orderIdStr)
}

func AdminSideItemCancel(c *gin.Context){
	var order models.Order
	var orderItem models.OrderItem

	itemId := c.Param("id")

	if err := db.Db.Where("id = ?",itemId).First(&orderItem).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_orderDetails.html",gin.H{"error":"Could not load order items,please try again later"})
		return 
	}

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderItem.OrderID).First(&order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_orderDetails.html",gin.H{"error":"Failed to load order details"})
		return 
	}

	if err := db.Db.Model(&models.Product_Variant{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity)).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_orderDetails.html",gin.H{"error":"Failed to update order"})
		return 
	}

	checkRemaing := 0

	for _,item := range order.OrderItems{
		if item.Status != "Cancelled"{
			checkRemaing++
		}
	}

	var product models.Product_Variant
	db.Db.Where("id = ?",orderItem.ProductID).First(&product)

	retrunAmount := orderItem.Price * float64(orderItem.Quantity) + product.Tax * float64(orderItem.Quantity)
	
	newTotal := order.SubTotal - retrunAmount
	ogbalance := order.TotalAmount
	

	if newTotal < 0 {
		newTotal = 0
	}

	if retrunAmount > order.TotalAmount {
		retrunAmount = order.TotalAmount - newTotal
	}

	valueCheck,usedCouponId,errVal := helper.GetOrderValue(order.ID,order.UserID,newTotal)
	var walletTransaction models.WalletTransaction
	db.Db.Where("order_id = ? AND user_id = ? AND type = ?",order.ID,order.UserID,"Debit").First(&walletTransaction)

	if valueCheck && errVal == nil{

		if newTotal == 0{
			order.SubTotal = 0
			order.TotalAmount = 0
		}else{
			order.SubTotal = order.SubTotal - retrunAmount
			order.TotalAmount = order.TotalAmount - retrunAmount
		}

	}else if !valueCheck && errVal == nil{

		if walletTransaction.ID != 0 {
			if newTotal == 0{
				order.SubTotal = 0
				order.TotalAmount = 0
				order.DiscountTotal = 0
			}else{
				order.SubTotal = newTotal
				order.TotalAmount = newTotal
				order.DiscountTotal = 0
			}
			
		}else{

			if newTotal == 0 {
				order.TotalAmount = 0
				order.SubTotal = 0
			}else{
				order.TotalAmount = newTotal
				order.SubTotal = newTotal
			}
			
			order.DiscountTotal = 0.0
		}
		
		if usedCouponId != 0 {
			db.Db.Delete(&models.UsedCoupon{},usedCouponId)
		}

	}else if errVal != nil{
		
		log.Println(errVal)
	}

	if order.PaymentMethod != "cod" && order.TotalAmount !=0{
		err := helper.CreditWallet(order.UserID,retrunAmount,"admin forwarded")
		if err != nil{
			log.Println(err)
		}
	}else if order.PaymentMethod != "cod" && order.TotalAmount == 0{
		err := helper.CreditWallet(order.UserID,ogbalance,"admin forwarded")
		if err != nil{
			log.Println(err)
		}
	}
	
	if checkRemaing == 0 || len(order.OrderItems) == 1 || newTotal == 0{

		if order.PaymentMethod != "cod"{
			order.PaymentStatus = "Refunded"
		}else{
			order.PaymentStatus = "Not vaild"
		}
		order.Status = "Cancelled"
		order.Reason = orderItem.Reason
	}


	orderItem.Status = "Cancelled"
	orderId := orderItem.OrderID
	orderIdStr := strconv.Itoa(int(orderId))
	db.Db.Save(&order)
	db.Db.Save(&orderItem)
	db.Db.Delete(&orderItem)

	
	c.Redirect(http.StatusSeeOther,"/admin/order/"+orderIdStr)

}