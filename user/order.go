package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

func ListOrders(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")

	_,userId,_ := helper.DecodeJWT(tokenStr)

	page,_ := strconv.Atoi(c.DefaultQuery("page","1"))
	limit := 10
	offset := (page - 1) * limit

	var orders []models.Order
	var total int64 

	db.Db.Model(&models.Order{}).Where("user_id = ?",userId).Count(&total)

	if err := db.Db.Where("user_id = ?",userId).Order("id DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil{

		if err == gorm.ErrRecordNotFound {
			c.HTML(http.StatusNotFound,"myOrders.html",gin.H{"user":"done"})
			return
		}else{
			c.HTML(http.StatusInternalServerError,"myOrders.html",gin.H{"error":"Failed to retrieve order details, please try again later"})
			return 
		}
		
	}

	for i := range orders {
		switch orders[i].Status{
		case "Delivered":
			orders[i].BadgeClass = "success"
		case "Processing", "Pending":
			orders[i].BadgeClass = "warning"
		case "Cancelled", "Returned":
			orders[i].BadgeClass = "danger"
		default:
			orders[i].BadgeClass = "secondary"
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	var pages []map[string]int
	for i := 1; i <= totalPages; i++{
		pages = append(pages, map[string]int{"Number":i})
	}

	c.HTML(http.StatusOK,"myOrders.html",gin.H{
		"user":"done",
		"Orders":orders,
		"CurrentPage": page,
		"Pages": pages,
		"HasPrev": page > 1,
		"HasNext": page < totalPages,
		"PrevPage": page - 1,
		"NextPage": page + 1,
	})

}

func ReturnOrder(c *gin.Context){

	orderId,_ := strconv.Atoi(c.PostForm("order_id"))
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

	var walletTransaction models.WalletTransaction
	if err := db.Db.Where("user_id = ? AND order_id = ?",order.UserID,orderId).First(&walletTransaction).Error; err != nil{
		if err != gorm.ErrRecordNotFound {
			log.Println(err)
			c.HTML(http.StatusInternalServerError,"myOrders.html",gin.H{"error":"Failed to load wallet details"})
			return 
		}
	}

	if order.PaymentMethod != "cod"{
		
		walletTransaction := models.WalletTransaction{
			UserID: order.UserID,
			OrderID: order.ID,
			Amount: order.TotalAmount,
			Type: "Credit",
			Description: "Refund",
			RefundStatus: true,
		}

		db.Db.Create(&walletTransaction)
	}
	
	if walletTransaction.ID != 0{

		newTransaction := models.WalletTransaction{
			UserID: walletTransaction.UserID,
			OrderID: walletTransaction.OrderID,
			Amount: walletTransaction.Amount,
			Type: "Credit",
			Description: "Refund request for "+strconv.Itoa(int(walletTransaction.ID)),
			RefundStatus: true,
		}

		db.Db.Create(&newTransaction)
	}
	
	if walletTransaction.ID != 0 || order.PaymentMethod != "cod"{
		order.Status = "Refund is being processed"
		order.Reason = reason
	}else{
		order.Status = "Returned"
		order.Reason = reason
	}

	if err := db.Db.Save(&order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"myOrders.html",gin.H{"error":"Failed to return item","user":"done"})
		return 
	}

	for _,item := range order.OrderItems {

		db.Db.Model(&models.Product_Variant{}).Where("id = ?",item.ProductID).Update("stock",gorm.Expr("stock + ?",item.Quantity))
		//db.Db.Delete(&models.OrderItem{},item.ID)
		if item.Status != "Return requested"{
			item.Status = "Return"
			db.Db.Save(&item)
		}
		
	}

	c.Redirect(http.StatusSeeOther,"/user/orders")

}

func OrderItems(c *gin.Context){

	orderID := c.Param("id")
	var Order models.Order
	var address models.Address

	type Response struct {
		ID 				uint 
		ImageURL		string
		ProductName		string
		Quantity		int
		Status 			string 
		Price 			float64
		Discount		float64
		Tax 			float64
	}
	
	
	if err := db.Db.Preload("OrderItems",func(db *gorm.DB)*gorm.DB{
		return db.Unscoped()
	}).Where("id = ?",orderID).First(&Order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"orderDetails.html",gin.H{"error":"Unable to find order details"})
		return 
	}

	if err := db.Db.Where("address_id = ?",Order.AddressID).First(&address).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"orderDetails.html",gin.H{"error":err})
		return 
	}


	switch Order.Status{
	case "Delivered":
		Order.BadgeClass = "success"
	case "Processing", "Pending":
		Order.BadgeClass = "warning"
	case "Cancelled", "Returned", "Failed":
		Order.BadgeClass = "danger"
	default:
		Order.BadgeClass = "secondary"
	}


	// if err := db.Db.Where("order_id = ?",orderID).Order("id DESC").Find(&OrderItems).Error; err != nil{
	// 	c.HTML(http.StatusBadRequest,"orderDetails.html",gin.H{"error":"Order items not found"})
	// 	return 
	// }

	response := make([]Response,len(Order.OrderItems))

	for i, item := range Order.OrderItems{
		
		var Product models.Product_Variant
		err := db.Db.Preload("Product_images").Where("id = ?",item.ProductID).First(&Product).Error

		if err != nil {
			c.HTML(http.StatusNotFound,"orderDetails.html",gin.H{"error":"Product details not found"})
			return 
		}

		if len(Product.Product_images) != 0{
			response[i] = Response{
				ID: item.ID,
				ProductName: Product.Variant_name,
				ImageURL: Product.Product_images[0].Image_url,
				Quantity: item.Quantity,
				Status: item.Status,
				Price: item.Price,
				Discount: 0.0,
				Tax: Product.Tax,
			}
		}else{
			response[i] = Response{
				ID: item.ID,
				ProductName: Product.Variant_name,
				ImageURL: "",
				Quantity: item.Quantity,
				Status: item.Status,
				Price: item.Price,
				Discount: 0.0,
				Tax: Product.Tax,
			}
		}

	}

	c.HTML(http.StatusOK,"orderDetails.html",gin.H{
		"OrderItems":response,
		"address":address,
		"Order": Order,
		"user": "done",

	})

}

func CancelItem (c *gin.Context){
	orderID := c.PostForm("order_id")
	itemId := c.PostForm("item_id")
	reason := c.PostForm("reason")
	var Order models.Order

	if err := db.Db.Where("id = ?",orderID).First(&Order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"orderDetails.html",gin.H{"error":"Failed to load order details,please try again later."})
		return 
	}

	if Order.PaymentMethod != "cod" {
		err := helper.ItemCancelOnline(orderID,itemId,reason)

		if err != nil{
			c.HTML(http.StatusInternalServerError,"orderDetails.html",gin.H{"error":err})
			return 
		}
	}else{
		err := helper.ItemCancelCod(orderID,itemId,reason)
		if err != nil{
			c.HTML(http.StatusInternalServerError,"orderDetails.html",gin.H{"error":err})
			return 
		}
	}

	

	c.Redirect(http.StatusSeeOther,"/user/order/"+orderID)

}

func DownloadPdf(c *gin.Context){

	orderID,_ := strconv.Atoi(c.Param("id"))
	var order models.Order
	var User models.User

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderID).First(&order).Error; err != nil {
		c.HTML(http.StatusInternalServerError,"myOrder.html",gin.H{"error":"Failed to fetch order details, please try again later"})
		return 
	}

	if err := db.Db.Where("id = ?",order.UserID).First(&User).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"myOrder.html",gin.H{"error":"Failed to fetch user details, please try again later"})
		return
	}

	pdf := gofpdf.New("P","mm","A4","")
	pdf.AddPage()
	
	pdf.SetFont("Arial","B",16)
	pdf.Cell(40,10,"Invoice")

	pdf.Ln(12)
	pdf.SetFont("Arial","",12)
	pdf.Cell(40,10,fmt.Sprintf("Order ID: %d",order.ID))
	pdf.Ln(8)
	pdf.Cell(40,10,fmt.Sprintf("Customer: %s",User.Username))
	pdf.Ln(8)
	pdf.Cell(40,10,fmt.Sprintf("Date: %s",order.OrderDate))

	pdf.Ln(12)
	pdf.SetFont("Arial","B",12)
	pdf.CellFormat(80, 10, "Product", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "Qty", "1", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, "Price", "1", 1, "", false, 0, "")

	pdf.SetFont("Arial", "", 12)

	for _,item := range order.OrderItems{

		var Product models.Product_Variant

		if err := db.Db.Where("id = ?",item.ProductID).First(&Product).Error; err != nil{
			c.HTML(http.StatusInternalServerError,"myOrder.html",gin.H{"error":"Failed to retrive product details"})
			return 
		}

		pdf.CellFormat(80, 10, Product.Variant_name, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%d", item.Quantity), "1", 0, "", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", item.Price), "1", 1, "", false, 0, "")
	}

	pdf.Ln(8)
	pdf.CellFormat(110, 10, "Total:", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", order.TotalAmount), "1", 1, "", false, 0, "")

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition","attachment; filename=invoice.pdf")
	err := pdf.Output(c.Writer)

	if err != nil{
		c.String(http.StatusInternalServerError,"Failed to generate PDF: %v",err)
	}

}