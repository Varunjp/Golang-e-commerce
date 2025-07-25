package user

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	razorpay "github.com/razorpay/razorpay-go"
)

func RetryOrderPage(c *gin.Context) {
	orderId := c.Query("orderId")

	c.HTML(http.StatusOK,"orderRetry.html",gin.H{"OrderId":orderId})
}

type RetryOrderRequest struct{
	OrderId string `json:"order_id"`
}

func RetryOrderCreate(c *gin.Context){
	var req RetryOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil || req.OrderId == ""{
		c.JSON(http.StatusBadRequest,gin.H{"success":false,"error":"Invalid request"})
		return 
	}

	var order models.Order

	if err := db.Db.Where("order_id = ?",req.OrderId).First(&order).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"success":false,"error":"Order not found"})
		return 
	}

	if order.Status != "Failed"{
		c.JSON(http.StatusBadRequest,gin.H{"success":false,"error":"Order already updated"})
		return 
	}

	err := helper.CheckProduct(order.ID)

	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"success":false,"error":"Product could not meet requirment"})
		return 
	}

	amount := order.TotalAmount

	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	data := map[string]interface{}{
		"amount": amount*100,
		"currency":"INR",
		"receipt":order.OrderID,
		"notes":map[string]interface{}{
			"retry":true,
			"user":order.UserID,
		},
	}

	resp,err := client.Order.Create(data,nil)

	if err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"success":false,"error":"Failed to create Razorpay order"})
		return 
	}

	razorOrderId := resp["id"].(string)

	c.JSON(http.StatusOK,gin.H{
		"success":true,
		"order_id":razorOrderId,
		"amount":amount *100,
		"currency":"INR",
		"key":os.Getenv("RAZORPAY_KEY_ID"),
	})
}

type RazorpaySuccessPayload struct{
	RazorpayPaymentId	string `json:"razorpay_payment_id"`
	RazorpayOrderID 	string `json:"razorpay_order_id"`
	RazorpaySignature 	string `json:"razorpay_signature"`
	OrderID				string `json:"order_id"`
}

func RetryPaymentSuccess(c *gin.Context){
	var payload RazorpaySuccessPayload
	if err := c.ShouldBindJSON(&payload); err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"success":false,"error":"Invalid request"})
		return 
	}


	secret := os.Getenv("RAZORPAY_KEY_SECRET")
	data := payload.RazorpayOrderID + "|" + payload.RazorpayPaymentId

	h := hmac.New(sha256.New,[]byte(secret))
	h.Write([]byte(data))
	generatedSignature := hex.EncodeToString(h.Sum(nil))

	if generatedSignature != payload.RazorpaySignature {
		c.JSON(http.StatusUnauthorized,gin.H{"success":false})
		return 
	}

	var order models.Order
	if err := db.Db.Preload("OrderItems").Where("order_id = ?",payload.OrderID).First(&order).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"success":false})
		return 
	}

	order.Status = "Processing"
	order.PaymentStatus = "Successful"
	order.OrderDate = time.Now()

	for _,item := range order.OrderItems{
		item.Status = "Processing"
		db.Db.Save(&item)
	}

	db.Db.Save(&order)

	c.JSON(http.StatusOK,gin.H{"success":true,"redirect":fmt.Sprintf("/order/confirmation/%d",order.ID)})
}

type OrderFailedPayload struct {
	OrderId string `json:"order_id"`
}

func RetryFailed(c *gin.Context){
	var req OrderFailedPayload
	if err := c.ShouldBindJSON(&req); err != nil || req.OrderId == "" {
		c.JSON(http.StatusBadRequest,gin.H{"success":false})
		return
	}

	var order models.Order
	if err := db.Db.Where("order_id = ?",req.OrderId).First(&order).Error; err != nil{
		order.Status = "Failed"
		order.OrderDate = time.Now()
		db.Db.Save(&order)
	}

	c.JSON(http.StatusOK,gin.H{
		"success":true,
		"redirect":"/user/order/"+strconv.Itoa(int(order.ID)),
	})
}