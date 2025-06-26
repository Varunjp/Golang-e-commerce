package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CancelOrderItem(c *gin.Context){
	OrderId := c.PostForm("order_id")
	ItemId := c.PostForm("item_id")
	reason := c.PostForm("reason")

	var order models.Order
	var orderItem models.OrderItem
	referer := c.Request.Referer()
	session := sessions.Default(c)

	if err := db.Db.Preload("OrderItems").Where("id = ?",OrderId).First(&order).Error; err != nil{
		session.Set("flash","Could not find order details, please try again later")
		session.Save()
		if referer != ""{
			c.Redirect(http.StatusSeeOther,referer)
		}else{
			c.Redirect(http.StatusSeeOther,"/user/orders")
		}
		return 
	}

	if err := db.Db.Where("id = ?",ItemId).First(&orderItem).Error; err != nil{
		session.Set("flash","Could not find order details, please try again later")
		session.Save()
		if referer != ""{
			c.Redirect(http.StatusSeeOther,referer)
		}else{
			c.Redirect(http.StatusSeeOther,"/user/orders")
		}
		return 
	}


	if order.PaymentMethod != "cod"{
		err := helper.ItemCancelOnline(OrderId,ItemId,reason)

		if err != nil{
			session.Set("flash","Could not cancel order item, please try again later")
			session.Save()
			if referer != ""{
				c.Redirect(http.StatusSeeOther,referer)
			}else{
				c.Redirect(http.StatusSeeOther,"/user/orders")
			}
			return 
		}
	}else{

		err := helper.ItemCancelCod(OrderId,ItemId,reason)
		if err != nil{
			session.Set("flash","Could not cancel order item, please try again later")
			session.Save()
			if referer != ""{
				c.Redirect(http.StatusSeeOther,referer)
			}else{
				c.Redirect(http.StatusSeeOther,"/user/orders")
			}
			return 
		}
	}

	err := helper.UserOrderCancelItem(ItemId)

	if err != nil{
		session.Set("flash","Could not update payment calculation, please try again later")
		session.Save()
		if referer != ""{
			c.Redirect(http.StatusSeeOther,referer)
		}else{
			c.Redirect(http.StatusSeeOther,"/user/orders")
		}
		return
	}

	cancelCount := 0
	db.Db.Preload("OrderItems").Where("id = ?",OrderId).First(&order)

	for _,item := range order.OrderItems{
		if item.Status == "Pending" || item.Status == "Processing" || item.Status == "Delivered" || item.Status == "Delivered non returnable"{
			cancelCount ++
		}
	}

	if cancelCount == 0 && order.PaymentMethod == "cod"{
		order.Status = "Cancelled"
		order.PaymentStatus = "Not Valid"
	}else if cancelCount == 0 && order.PaymentMethod != "cod"{
		order.Status = "Cancelled"
		order.PaymentStatus = "Refunded"
	}

	orderItem.Status = "Cancelled"
	db.Db.Save(&order)
	db.Db.Save(&orderItem)

	c.Redirect(http.StatusSeeOther,"/user/order/"+OrderId)
}

func CancelOrder(c *gin.Context){
	orderId,_ := strconv.Atoi(c.PostForm("order_id"))
	OrderIDStr := c.PostForm("order_id")
	reason := c.PostForm("reason")
	var order models.Order

	if reason == "" {
		c.HTML(http.StatusBadRequest,"myOrders.html",gin.H{"error":"Please provide a reason","user":"done"})
		return 
	}

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderId).First(&order).Error; err != nil{
		c.HTML(http.StatusNotFound,"myOrders.html",gin.H{"error":"Order not found","user":"done"})
		return 
	}

	if order.Status == "Returned" {
		c.HTML(http.StatusBadRequest,"myOrders.html",gin.H{"error":"Cannot return order","user":"done"})
		return 
	}

	var couponUsed models.UsedCoupon

	if err := db.Db.Where("user_id = ? AND order_id = ?",order.UserID,orderId).First(&couponUsed).Error; err == nil{

		if err := db.Db.Delete(&models.UsedCoupon{},couponUsed.ID).Error; err != nil{
			c.HTML(http.StatusInternalServerError,"myOrders.html",gin.H{"error":"Failed to update order please try again later"})
			return 
		}

	}else{
		log.Println(err)
	}

	var WalletTransaction models.WalletTransaction
	if err := db.Db.Where("user_id = ? AND order_id = ? AND type = ?",order.UserID,orderId,"Debit").First(&WalletTransaction).Error; err != nil{
		if err != gorm.ErrRecordNotFound {
			log.Println(err)
			c.HTML(http.StatusInternalServerError,"myOrders.html",gin.H{"error":"Failed to load wallet details"})
			return 
		}
	}

	var walletAmount float64
	if WalletTransaction.ID != 0{
		walletAmount = math.Abs(WalletTransaction.Amount)
	}

	if order.PaymentMethod != "cod" || order.Status == "Delivered" {
		
		walletTransaction := models.WalletTransaction{
			UserID: order.UserID,
			OrderID: order.ID,
			Amount: order.TotalAmount+walletAmount,
			Type: "Credit",
			Description: "Refund",
			RefundStatus: true,
		}

		db.Db.Create(&walletTransaction)
		
	}else if order.PaymentMethod == "cod" && WalletTransaction.ID != 0{
		
		newTransaction := models.WalletTransaction{
			UserID: WalletTransaction.UserID,
			OrderID: WalletTransaction.OrderID,
			Amount: walletAmount,
			Type: "Credit",
			Description: "Refund request for order :"+strconv.Itoa(int(WalletTransaction.OrderID)),
			RefundStatus: true,
		}

		db.Db.Create(&newTransaction)
	}
	
	
	if WalletTransaction.ID != 0 || order.PaymentMethod != "cod" || order.Status == "Delivered"{
		order.Status = "Cancelled"
		order.PaymentStatus = "Refunded"
		order.Reason = reason
	}else{
		order.Status = "Cancelled"
		order.PaymentStatus = "Failed"
		order.Reason = reason
	}

	if err := db.Db.Save(&order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"myOrders.html",gin.H{"error":"Failed to return item","user":"done"})
		return 
	}

	for _,item := range order.OrderItems {

		if item.Status != "Returned"{
			item.Status = "Cancelled"
			item.Reason =  reason
			db.Db.Save(&item)
		}
		
	}
	c.Redirect(http.StatusSeeOther,"/user/order/"+OrderIDStr)
}