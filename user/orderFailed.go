package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func OrderFailed(c *gin.Context){
	
	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)
	var cartItems []models.CartItem
	var address models.Address
	
	var Payload struct {
		AddressID		string 		`json:"address_id"`
	}


	if err := c.ShouldBindJSON(&Payload); err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":err})
		return 
	}

	if err := db.Db.Where("user_id = ?",userID).Find(&cartItems).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":err})
		return 
	}


	if err := db.Db.Where("address_id = ?",Payload.AddressID).First(&address).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":err})
		return 
	}

	var total float64

	for _,item := range cartItems{
		var product models.Product_Variant
		total += item.Price * float64(item.Quantity)
		db.Db.Where("id = ?",item.ProductID).First(&product)
		tax := product.Tax * float64(item.Quantity)
		total +=tax
	}

	order := models.Order{
		UserID: uint(userID),
		AddressID: address.AddressID,
		TotalAmount: total,
		SubTotal: total,
		DiscountTotal: 0.0,
		Status: "Failed",
		PaymentMethod: "Razorpay",
		PaymentStatus: "Failed",
		CreateAt: time.Now(),
	}


	if err := db.Db.Create(&order).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":err})
		return 
	}

	// var address models.Address

	// db.Db.Where("address_id  = ?",addressID).First(&address)

	OrderAddress := models.OrderAddress{
		OrderID: order.ID,
		AddressLine1: address.AddressLine1,
		AddressLine2: address.AddressLine2,
		Country: address.Country,
		City: address.City,
		State: address.State,
		PostalCode: address.PostalCode,
	}

	db.Db.Create(&OrderAddress)

	for _, item := range cartItems{

		orderItem := models.OrderItem{
			OrderID: order.ID,
			UserID: uint(userID),
			ProductID: item.ProductID,
			Quantity: item.Quantity,
			Price: item.Price,
			Status: "Canceled",
		}


		if err := db.Db.Create(&orderItem).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err})
			return 
		}

	}

	c.JSON(http.StatusOK,gin.H{"success":true,"redirect": fmt.Sprintf("/order/failed-page/%d",order.ID)})
}

func OrderFailedPage(c *gin.Context){
	orderId := c.Param("id")

	c.HTML(http.StatusOK,"orderFailed.html",gin.H{
		"OrderID":orderId,
		"user":"done",
	})
}