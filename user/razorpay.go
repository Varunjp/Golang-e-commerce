package user

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	razorpay "github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

func CreateRazorpayOrder(c *gin.Context){


	var req struct {
		AddressID 	string		`json:"address_id"`
		Amount 		float64  	`json:"amount"`
		CouponCode	string 		`json:"coupon_code"`
		IsWallet    bool    	`json:"is_wallet"`
	}

	
	if err := c.ShouldBindJSON(&req); err != nil{
		
		c.JSON(http.StatusBadRequest,gin.H{"success":false})
		log.Println("Json :",err)
		return 
	}

	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)
	var usedcouponcheck models.UsedCoupon
	var totalAmount float64

	couponCode := req.CouponCode

	if couponCode != ""{
		if err := db.Db.Where("user_id = ? AND coupon_id = ?",userID,couponCode).First(&usedcouponcheck).Error; err == nil{
			c.JSON(http.StatusBadRequest,gin.H{"success":false})
			return 
		}

	}
	
	var orderitems []models.OrderItem
	
	if err := db.Db.Where("user_id = ? AND deleted_at IS NULL",userID).Find(&orderitems).Error; err != nil{
		if err != gorm.ErrRecordNotFound{
			c.JSON(http.StatusInternalServerError,gin.H{"success":false})
			log.Println(err)
			return 
		}
	}

	var CartItems []models.CartItem

	if err := db.Db.Where("user_id = ?",userID).Find(&CartItems).Error; err != nil {
		c.JSON(http.StatusNotFound,gin.H{"success":false})
		return 
	}
	
	if len(CartItems) < 1 {
		c.JSON(http.StatusNotFound,gin.H{"success":false})
		return
	}
	
	for _,item := range CartItems{

		var Product models.Product_Variant
		db.Db.Where("id = ?",item.ProductID).First(&Product)
		itemCount := 0
		totalAmount += item.Price * float64(item.Quantity)
		totalAmount += Product.Tax * float64(item.Quantity)

		for _, oritems := range orderitems{
			if item.ProductID == oritems.ProductID{
				itemCount = item.Quantity + oritems.Quantity
			}
		}

		if itemCount >= 5 {

			db.Db.Model(&models.Product_Variant{}).Where("id = ?",item.ProductID).Update("stock",gorm.Expr("stock + ?",item.Quantity))

			c.JSON(http.StatusBadRequest,gin.H{"success":false})
			return 
		}

	}

	if couponCode != "" {
		var Coupon models.Coupons
		db.Db.Where("id = ?",couponCode).First(&Coupon)
		if totalAmount > Coupon.MinAmount {
			discount := (totalAmount * Coupon.Discount)/100
			totalAmount = totalAmount - discount
		}
	}

	totalAmount = totalAmount *100

	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	data := map[string]interface{}{
		"amount": int(totalAmount),
		"currency": "INR",
		"receipt":fmt.Sprintf("order_rcptid_%d",time.Now().Unix()),
	}

	body, err := client.Order.Create(data,nil)

	if err != nil {
		log.Println("error from order :",err.Error())
		c.JSON(http.StatusInternalServerError,gin.H{"success":false,"error":"Failed to create Razorpay order"})
		return 
	}

	c.JSON(http.StatusOK,gin.H{
		"key":	os.Getenv("RAZORPAY_KEY_ID"),
		"amount": totalAmount,
		"currency": "INR",
		"order_id": body["id"],
		"success":true,
	})

}


func PaymentSuccess(c *gin.Context){

	var payload struct {
		RazorpayPaymentID 		string		`json:"razorpay_payment_id"`
		RazorpayOrderID			string		`json:"razorpay_order_id"`
		RazorpaySignature 		string 		`json:"razorpay_signature"`
		AddressID				string 		`json:"address_id"`	
		CouponCode				string 		`json:"coupon_code"`
		Amount 					float64  	`json:"amount"`
		IsWallet    			bool    	`json:"is_wallet"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil{
		log.Println(err)
		c.JSON(http.StatusBadRequest,gin.H{"success":false})
		return 
	}

	secret := os.Getenv("RAZORPAY_KEY_SECRET")
	data := payload.RazorpayOrderID + "|" + payload.RazorpayPaymentID

	h := hmac.New(sha256.New,[]byte(secret))
	h.Write([]byte(data))
	generatedSignature := hex.EncodeToString(h.Sum(nil))

	if generatedSignature != payload.RazorpaySignature {
		c.JSON(http.StatusUnauthorized,gin.H{"success":false})
		return 
	}

	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)

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
	var totalTax float64
	var totalAmount float64
	for _,item := range CartItems{
		var product models.Product_Variant
		db.Db.Where("id = ?",item.ProductID).First(&product)
		total += item.Price * float64(item.Quantity)
		totalTax += product.Tax * float64(item.Quantity)
		totalAmount+= item.Price * float64(item.Quantity) + totalTax
	}

	addressintId,_ := strconv.Atoi(payload.AddressID)
	var coupon models.Coupons

	if payload.CouponCode != ""{
		if err := db.Db.Where("id = ?",payload.CouponCode).First(&coupon).Error; err != nil{
			if err != gorm.ErrRecordNotFound {
				c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Failed to load details"})
				return 
			}
		}
	}
	

	var discount float64
	if coupon.ID != 0 {

		if totalAmount > coupon.MinAmount {
			discount = (total * coupon.Discount)/100
		}
		
		if coupon.MaxAmount < discount {
			discount = coupon.MaxAmount	
		}

		if payload.IsWallet {
				wallet := (total+totalTax)-discount
				discount += wallet
		}
	}else if payload.IsWallet {
		discount = (total+totalTax) - payload.Amount
	}
	
	neOrderId := helper.GenerateOrderID()
	totalAmount = totalAmount - discount

	order := models.Order{
		UserID: uint(userID),
		OrderID: neOrderId,
		AddressID: uint(addressintId),
		DiscountTotal: discount,
		SubTotal: total+totalTax,
		TotalTax: totalTax,
		TotalAmount: totalAmount,
		Status: "Processing",
		PaymentStatus: "Successful",
		PaymentMethod: "Razorpay",
		PaymentID: payload.RazorpayOrderID,
		CreateAt: time.Now(),
	}


	if err := db.Db.Create(&order).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to create order"})
		return 
	}

	var address models.Address

	db.Db.Where("address_id  = ?",addressintId).First(&address)

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

	couponCode := payload.CouponCode

	if couponCode != ""{

		
		usedcoupon := models.UsedCoupon{
			UserID: order.UserID,
			CouponID: coupon.ID,
			OrderID: order.ID,
		}

		if err := db.Db.Create(&usedcoupon).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error while saving coupon upadate please try again later."})
			return 
		}

	}

	if payload.IsWallet {

		var walletAmount float64

		if coupon.ID != 0 {
			couponDiscount := math.Round((total+totalTax) * coupon.Discount /100) 
			walletAmount = ((total+totalTax) - payload.Amount)-couponDiscount
		}else{
			walletAmount = (total+totalTax) - payload.Amount
		}

		err := helper.DebitWallet(uint(userID),walletAmount,order.ID,"Order debit")

		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"success":false})
			return 
		}
	}

	var orderitems []models.OrderItem
	
	if err := db.Db.Where("user_id = ? AND deleted_at IS NULL",userID).Find(&orderitems).Error; err != nil{
		if err != gorm.ErrRecordNotFound{
			c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Failed to load user details please try again later"})
			return 
		}
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

		if itemCount >= 5 {
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
			Status: "Processing",
			Price: item.Price,
		}

		if err := db.Db.Create(&orderItem).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to add order items"})
			return 
		}

		// delete from wishlist
		var wishlist models.WishList

		if err := db.Db.Where("user_id = ? AND product_id = ?",userID,item.ProductID).First(&wishlist).Error; err == nil{
			db.Db.Delete(&wishlist)
		}

	}

	if err := db.Db.Delete(&CartItems).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to clear cart items"})
		return 
	}


	c.JSON(http.StatusOK,gin.H{"success":true,"redirect": fmt.Sprintf("/order/confirmation/%d",order.ID)})
}