package admin

import (
	db "first-project/DB"
	"first-project/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
	
	orderItem.Status = "Delivered non returnable"
	orderId := orderItem.OrderID
	orderIdStr := strconv.Itoa(int(orderId))
	db.Db.Save(&orderItem)
	db.Db.Save(&order)
	// db.Db.Delete(&orderItem)

	c.Redirect(http.StatusSeeOther,"/admin/order/"+orderIdStr)
}