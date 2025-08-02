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

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

func ListOrders(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")
	_,userId,_ := helper.DecodeJWT(tokenStr)

	qstatus := c.Query("status")
	page,_ := strconv.Atoi(c.DefaultQuery("page","1"))
	limit := 10
	offset := (page - 1) * limit

	var orders []models.Order
	var total int64 
	var query *gorm.DB

	if qstatus != ""{
		query = db.Db.Model(&models.Order{}).Where("user_id = ? AND status = ?",userId,qstatus).Count(&total)
	}else{
		query = db.Db.Model(&models.Order{}).Where("user_id = ?",userId).Count(&total)
	}

	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil{

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
		"FilterStatus":qstatus,
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

	if order.Status == "Returned" || order.Status == "Return requested"{
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

		desc := fmt.Sprintf("Refund request for order : %s",order.OrderID)
		
		walletTransaction := models.WalletTransaction{
			UserID: order.UserID,
			OrderID: order.ID,
			Amount: order.TotalAmount+walletAmount,
			Type: "Credit",
			Description: desc,
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
		order.Status = "Return requested"
		order.PaymentStatus = "Refund is being processed"
		order.Reason = reason
	}else{
		order.Status = "Return requested"
		order.PaymentStatus = "Failed"
		order.Reason = reason
	}

	if err := db.Db.Save(&order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"myOrders.html",gin.H{"error":"Failed to return item","user":"done"})
		return 
	}

	for _,item := range order.OrderItems {

		if item.Status != "Returned"{
			item.Status = "Return requested"
			item.Reason =  reason
			db.Db.Save(&item)
		}
		
	}

	c.Redirect(http.StatusSeeOther,"/user/orders")

}

func OrderItems(c *gin.Context){

	orderID := c.Param("id")
	var Order models.Order
	var address models.OrderAddress

	type Response struct {
		ID 				uint 
		ImageURL		string
		ProductID 		uint 
		ProductName		string
		Quantity		int
		Status 			string 
		Size 			string
		Price 			float64
		Discount		float64
		Tax 			float64
		ReviewCheck 	bool
		ParentProduct  	uint
	}
	
	
	if err := db.Db.Preload("OrderItems",func(db *gorm.DB)*gorm.DB{
		return db.Unscoped()
	}).Where("id = ?",orderID).First(&Order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"orderDetails.html",gin.H{"error":"Unable to find order details"})
		return 
	}

	if err := db.Db.Where("order_id = ?",Order.ID).First(&address).Error; err != nil{
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


	response := make([]Response,len(Order.OrderItems))

	for i, item := range Order.OrderItems{
		
		var Product models.Product_Variant
		var count int64
		err := db.Db.Preload("Product_images",func(DB *gorm.DB)*gorm.DB{
			return DB.Order("order_no,product_image_id ASC")
		}).Where("id = ?",item.ProductID).Unscoped().First(&Product).Error

		if err := db.Db.Model(&models.Review{}).Where("user_id = ? AND product_id = ? AND order_id = ?",Order.UserID,Product.ProductID,Order.ID).Count(&count).Error; err != nil{
			log.Println(err)
		}

		var check bool

		if count > 0{
			check = false
		}else{
			check = true
		}

		if err != nil {
			c.HTML(http.StatusNotFound,"orderDetails.html",gin.H{"error":"Product details not found"})
			return 
		}

		if len(Product.Product_images) != 0{
			response[i] = Response{
				ID: item.ID,
				ProductID: Product.ID,
				ProductName: Product.Variant_name,
				ImageURL: Product.Product_images[0].Image_url,
				Quantity: item.Quantity,
				Status: item.Status,
				Size: Product.Size,
				Price: item.Price,
				Discount: 0.0,
				Tax: Product.Tax,
				ReviewCheck: check,
				ParentProduct: Product.ProductID,
			}
		}else{
			response[i] = Response{
				ID: item.ID,
				ProductID: Product.ID,
				ProductName: Product.Variant_name,
				ImageURL: "",
				Quantity: item.Quantity,
				Status: item.Status,
				Size: Product.Size,
				Price: item.Price,
				Discount: 0.0,
				Tax: Product.Tax,
				ReviewCheck: check,
				ParentProduct: Product.ProductID,
			}
		}

	}

	session := sessions.Default(c)
	flash := session.Get("flash")

	if flash != nil{
		session.Delete("flash")
		session.Save()
		c.HTML(http.StatusOK,"orderDetails.html",gin.H{
			"OrderItems":response,
			"address":address,
			"Order": Order,
			"user": "done",
			"error":flash,
		})
		return 
	}

	c.HTML(http.StatusOK,"orderDetails.html",gin.H{
		"OrderItems":response,
		"address":address,
		"Order": Order,
		"user": "done",

	})

}

func ReturnItem (c *gin.Context){
	orderID := c.PostForm("order_id")
	itemId := c.PostForm("item_id")
	reason := c.PostForm("reason")
	var Order models.Order
	var OrderItem models.OrderItem

	if err := db.Db.Where("id = ?",orderID).First(&Order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"orderDetails.html",gin.H{"error":"Failed to load order details,please try again later."})
		return 
	}

	if err := db.Db.Where("id = ? AND status = ?",itemId,"Delivered").First(&OrderItem).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"orderDetails.html",gin.H{"error":"Order already returned"})
		return 
	}

	if Order.PaymentMethod != "cod" {
		err := helper.ItemReturnOnline(orderID,itemId,reason)

		if err != nil{
			c.HTML(http.StatusInternalServerError,"orderDetails.html",gin.H{"error":err})
			return 
		}
	}else{

		err := helper.ItemReturnCod(orderID,itemId,reason)
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
	var address models.OrderAddress

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderID).First(&order).Error; err != nil {
		c.HTML(http.StatusInternalServerError,"myOrder.html",gin.H{"error":"Failed to fetch order details, please try again later"})
		return 
	}

	if err := db.Db.Where("id = ?",order.UserID).First(&User).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"myOrder.html",gin.H{"error":"Failed to fetch user details, please try again later"})
		return
	}

	if err := db.Db.Where("order_id = ?",order.ID).First(&address).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"myOrder.html",gin.H{"error":"Failed to fetch user details, please try again later"})
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Heading
	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(0, 12, "INVOICE", "", 1, "C", false, 0, "")

	pdf.Ln(5)

	// Company Details
	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(0, 6, "Sold By: Fashion Art Private Limited")
	pdf.Ln(6)
	pdf.Cell(0, 6, "From Address: FashionArtify, Calicut")
	pdf.Ln(6)
	pdf.Cell(0, 6, "GSTIN: 39AACCB8899Z1Z")
	pdf.Ln(10)

	// Invoice Info
	pdf.Cell(0, 6, fmt.Sprintf("Order ID: %v", order.OrderID))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Order Date: %s", order.OrderDate.Format("1/2/2006")))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Invoice Date: %s", order.OrderDate.Format("1/2/2006")))
	pdf.Ln(10)

	// Billing Address
	pdf.SetFont("Helvetica", "BU", 12)
	pdf.Cell(0, 6, "Billing Address:")
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(0, 6, User.Username)
	pdf.Ln(6)
	pdf.Cell(0, 6,  address.AddressLine1)
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("%s, %s - %s", address.City, address.State, address.PostalCode))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Phone: %s", User.Phone))
	pdf.Ln(10)

	// Table Header
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(80, 10, "Description", "1", 0, "", false, 0, "")
	pdf.CellFormat(20, 10, "Qty", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "Price", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "Discount", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "Total", "1", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetFillColor(255, 255, 255)

	var subtotal float64

	for _, item := range order.OrderItems {
		var product models.Product_Variant
		if err := db.Db.Unscoped().Where("id = ?", item.ProductID).First(&product).Error; err != nil {
			c.String(http.StatusInternalServerError, "Error fetching product: %v", err)
			return
		}

		lineTotal := float64(item.Quantity) * item.Price
		subtotal += lineTotal

		desc := product.Variant_name
		if pdf.GetStringWidth(desc) > 75 { // 75 to keep within 80mm width
			for len(desc) > 0 && pdf.GetStringWidth(desc+"...") > 75 {
				desc = desc[:len(desc)-1]
			}
			desc += "..."
		}

		// Continue the rest of the row at same height
		pdf.CellFormat(80, 16, desc, "1", 0, "", false, 0, "")
		pdf.CellFormat(20, 16, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 16, fmt.Sprintf("Rs. %.2f", item.Price), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 16, "Rs. 0.00", "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 16, fmt.Sprintf("Rs. %.2f", lineTotal), "1", 1, "C", false, 0, "")
	}

	// Summary
	pdf.SetFont("Helvetica", "B", 10)

	pdf.CellFormat(160, 10, "Subtotal", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("Rs. %.2f", order.SubTotal), "1", 1, "C", false, 0, "")

	pdf.CellFormat(160, 10, "Coupon Discount", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("- Rs. %.2f", order.DiscountTotal), "1", 1, "C", false, 0, "")

	finalPayable := order.SubTotal - order.DiscountTotal
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetFillColor(230, 255, 230) // Light green for total
	pdf.CellFormat(160, 10, "Final Payable", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("Rs. %.2f", finalPayable), "1", 1, "C", false, 0, "")

	// Footer / Signature
	pdf.Ln(15)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, "Fashion Art Private Limited", "", 1, "R", false, 0, "")
	pdf.CellFormat(0, 6, "(Signature Placeholder)", "", 1, "R", false, 0, "")
	pdf.CellFormat(0, 6, "Authorized Signatory", "", 1, "R", false, 0, "")

	// Send as response
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=invoice.pdf")

	if err := pdf.Output(c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF: %v", err)
	}
}