package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CheckOutPage(c *gin.Context) {

	var CartItems []models.CartItem
	var Addresses []models.Address
	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)

	if err := db.Db.Preload("Product").Where("user_id = ? AND ",userID).Find(&CartItems).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to load data from DB"})
		return 
	}

	if err := db.Db.Where("user_id = ?",userID).Find(&Addresses).Error; err != nil{

		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Not able to fetch address from db"})
			return
		}
	
	}

	// product name, imageurl, total sum, total tax, applicable discount 

	var Response []struct {
		ID 				uint 
		Name 			string
		ImageUrl 		string 
		Quantity 		int 
		TotalSum		float64
		TotalTax		float64
		TotalDiscount	float64
		GrandTotal		float64
	}

	for _,item := range CartItems{

		res := helper.ValidateProduct(item.ProductID,item.Quantity)

		if res {

			Response = append(Response, struct{ID uint; Name string; ImageUrl string; Quantity int; TotalSum float64; TotalTax float64; TotalDiscount float64; GrandTotal float64}{
				ID: item.ProductID,
				Name: item.Product.Variant_name,
				ImageUrl: item.Product.Product_images[0].Image_url,
				Quantity: item.Quantity,
				TotalSum: (item.Price * float64(item.Quantity)),
				TotalTax: (item.Product.Tax*float64(item.Quantity)),
				TotalDiscount: 0.0,
				GrandTotal: (item.Price * float64(item.Quantity))+(item.Product.Tax*float64(item.Quantity)),
			})

		}

	}

	c.JSON(http.StatusOK,gin.H{"items":Response,"addresses":Addresses})

}

func CheckOutOrder(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)
	addressOption := c.PostForm("address_option")
	paymentOption := c.PostForm("payment_method")
	var addressID uint 

	if addressOption == "new" {
		newAddress := models.Address{
			UserID: uint(userID),
			AddressLine1: c.PostForm("line1"),
			AddressLine2: c.PostForm("line2"),
			Country: c.PostForm("country"),
			State: c.PostForm("state"),
			PostalCode: c.PostForm("postalCode"),
		}

		if err := db.Db.Create(&newAddress).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to save address"})
			return 
		}

		addressID = newAddress.AddressID
	}else {

		id,_ := strconv.Atoi(addressOption)
		addressID = uint(id)

	}

	var CartItems []models.CartItem

	if err := db.Db.Where("user_id = ?",userID).Find(&CartItems).Error; err != nil {
		c.JSON(http.StatusNotFound,gin.H{"error":"Not able to load cart items"})
		return 
	}

	if len(CartItems) == 0 {
		c.JSON(http.StatusBadRequest,gin.H{"error":"Cart is empty"})
		return 
	}

	var total float64

	for _,item := range CartItems{
		total += item.Price * float64(item.Quantity)
	}

	order := models.Order{
		UserID: uint(userID),
		AddressID: addressID,
		TotalAmount: total,
		Status: "Processing",
		PaymentStatus: paymentOption,
		CreateAt: time.Now(),
	}

	if err := db.Db.Create(&order).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to create order"})
		return 
	}

	for _,item := range CartItems{

		if err := db.Db.Model(&models.Product_Variant{}).Where("id = ? AND stock >= ?",item.ProductID,item.Quantity).Update("stock",gorm.Expr("stock - ?",item.Quantity)).Error; err != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":"Insufficient stock"})
			return 
		}

		orderItem := models.OrderItem{
			UserID: uint(userID),
			OrderID: order.ID,
			ProductID: item.ProductID,
			Quantity: item.Quantity,
			Price: item.Price,		
		}

		if err := db.Db.Create(&orderItem).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to add order items"})
			return 
		}

	}

	if err := db.Db.Delete(&CartItems).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to clear cart items"})
		return 
	}

	c.JSON(http.StatusOK,gin.H{"message":"Order placed successfully"})

}

