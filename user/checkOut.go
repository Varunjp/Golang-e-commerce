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

	if err := db.Db.Preload("Product").Where("user_id = ?",userID).Find(&CartItems).Error; err != nil{
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
		Quantity 		int
		Price  			float64 
		TotalSum		float64
		TotalTax		float64
		TotalDiscount	float64
		GrandTotal		float64
	}

	for _,item := range CartItems{

		res := helper.ValidateProduct(item.ProductID,item.Quantity)

		if res {

			Response = append(Response, struct{ID uint; Name string; Quantity int;Price float64; TotalSum float64; TotalTax float64; TotalDiscount float64; GrandTotal float64}{
				ID: item.ProductID,
				Name: item.Product.Variant_name,
				Quantity: item.Quantity,
				Price: item.Price,
				TotalSum: (item.Price * float64(item.Quantity)),
				TotalTax: (item.Product.Tax*float64(item.Quantity)),
				TotalDiscount: 0.0,
				GrandTotal: (item.Price * float64(item.Quantity))+(item.Product.Tax*float64(item.Quantity)),
			})

		}

	}

	var totalamount float64

	for _, item := range Response{
		totalamount += item.GrandTotal
	}

	c.HTML(http.StatusOK,"checkOut.html",gin.H{"user":"done","CartItems":Response,"Addresses":Addresses,"TotalAmount":totalamount})

}

func CheckOutOrder(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)
	addressOption := c.PostForm("address_id")
	paymentOption := c.PostForm("payment_method")
	var addressID uint 
	var orderitems []models.OrderItem
	
	if addressOption == ""{
		c.HTML(http.StatusBadRequest,"checkOut.html",gin.H{"error":"Need to provide a address details"})
		return 
	}

	if paymentOption != "cod"{
		c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Online option not available yet please change payment method to cod"})
		return
	}

	if err := db.Db.Where("user_id = ? AND deleted_at IS NULL",userID).Find(&orderitems).Error; err != nil{
		if err != gorm.ErrRecordNotFound{
			c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Failed to load user details please try again later"})
			return 
		}
	}

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

		itemCount := 0

		if err := db.Db.Model(&models.Product_Variant{}).Where("id = ? AND stock >= ?",item.ProductID,item.Quantity).Update("stock",gorm.Expr("stock - ?",item.Quantity)).Error; err != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":"Insufficient stock"})
			return 
		}

		for _, oritems := range orderitems{
			if item.ProductID == oritems.ProductID{
				itemCount = item.Quantity + oritems.Quantity
			}
		}

		if itemCount > 5 {
			db.Db.Model(&models.Product_Variant{}).Where("id = ?",item.ProductID).Update("stock",gorm.Expr("stock + ?",item.Quantity))
			db.Db.Delete(&models.Order{},order.ID)
			c.HTML(http.StatusBadRequest,"checkOut.html",gin.H{"user":"done","error":"User exceeded product purchase limit"})
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

	c.HTML(http.StatusOK,"orderSuccess.html",gin.H{"OrderID":order.ID,"user":"done"})

}


func AddNewAddressPage(c *gin.Context){
	c.HTML(http.StatusOK,"addAddress.html",gin.H{"user":"done"})
}

func AddNewAddress(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)

	AddressLine1:= c.PostForm("line1")
	AddressLine2:= c.PostForm("line2")
	Country:= c.PostForm("country")
	State:= c.PostForm("state")
	PostalCode:= c.PostForm("postalcode")
	City := c.PostForm("city")

	// address.AddressLine1 = AddressLine1
	// address.AddressLine2 = AddressLine2
	// address.Country = Country
	// address.State = State
	// address.PostalCode = PostalCode
	// address.City = City
	// address.UserID = uint(userID)

	address := models.Address{
		UserID: uint(userID),
		AddressLine1: AddressLine1,
		AddressLine2: AddressLine2,
		Country: Country,
		State: State,
		PostalCode: PostalCode,
		City: City,
	}

	if err := db.Db.Create(&address).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Failed to save new address"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/user/checkout")

}