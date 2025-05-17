package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListOrders(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")

	_,userId,_ := helper.DecodeJWT(tokenStr)

	var orders []models.Order


	if err := db.Db.Where("user_id = ?",userId).Find(&orders).Order("id DESC").Error; err != nil{

		if err == gorm.ErrRecordNotFound {
			c.Redirect(http.StatusTemporaryRedirect,"/user/orders")
			return
		}else{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to retrieve data from db"})
			return 
		}
		
	}

	c.JSON(http.StatusOK,gin.H{"orders":orders})

}

func ReturnOrder(c *gin.Context){

	orderId,_ := strconv.Atoi(c.PostForm("id"))
	reason := c.PostForm("reason")
	var order models.Order

	if reason == "" {
		c.JSON(http.StatusBadRequest,gin.H{"error":"Please provide a reason"})
		return 
	}

	if err := db.Db.Where("id = ?",orderId).First(&order).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Order not found"})
		return 
	}

	if order.Status != "Delivered" || order.Status == "Returned" {
		c.JSON(http.StatusBadRequest,gin.H{"error":"Cannot return order"})
		return 
	}

	order.Status = "Returned"

	if err := db.Db.Save(&order).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to return item"})
		return 
	}

	for _,item := range order.OrderItems {

		db.Db.Model(&models.Product_Variant{}).Where("id = ?",item.ProductID).Update("stock",gorm.Expr("stock + ?",item.Quantity))

	}

	c.JSON(http.StatusOK,gin.H{"message":"Order returned successfully"})

}

func OrderItems(c *gin.Context){

	orderID := c.Param("id")
	var OrderItems []models.OrderItem

	if err := db.Db.Where("order_id = ?",orderID).Order("id DESC").Find(&OrderItems).Error; err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Order items not found"})
		return 
	}

	c.JSON(http.StatusOK,gin.H{"orderItems":OrderItems})

}