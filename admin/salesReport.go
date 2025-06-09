package admin

import (
	"bytes"
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

func SalesReportPage(c *gin.Context){

	// Pagination
	pageStr := c.DefaultQuery("page","1")
	limitStr := c.DefaultQuery("limit","10")

	page, err := strconv.Atoi(pageStr)

	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1{
		limit = 10
	}

	offset := (page - 1) * limit
	var total int64
	
	// filters
	rangeType := c.Query("filter")
	startDate := c.Query("start")
	endDate := c.Query("end")

	var start, end time.Time

	now := time.Now()

	switch rangeType {
	case "daily":
		start = now.Truncate(24 *time.Hour)
		end = start.Add(24 *time.Hour)
	case "weekly":
		start = now.AddDate(0,0,-int(now.Weekday()))
		end = start.AddDate(0,0,7)
	case "montly":
		start = time.Date(now.Year(), now.Month(),1,0,0,0,0,now.Location())
		end = start.AddDate(0,1,0)
	case "yearly":
		start = time.Date(now.Year(),1,1,0,0,0,0,now.Location())
		end = start.AddDate(1,0,0)
	case "custom":
		var err error 
		start, err = time.Parse("2006-01-02",startDate)
		end,_ = time.Parse("2006-01-02",endDate)
		if err != nil{
			c.HTML(http.StatusBadRequest,"sales_report.html",gin.H{"error":"Invalid date format"})
			return 
		}
	default:
		start = time.Time{}
		end = now
	}

	
	var orders []models.Order

	dbFullordes := db.Db.Model(&models.Order{})

	dbFullordes.Count(&total)

	if err := db.Db.Preload("OrderItems").Where("create_at BETWEEN ? AND ?",start,end).Order("id DESC").Limit(limit).Offset(offset).Find(&orders).Error; err != nil{
		if err != gorm.ErrRecordNotFound{
			c.HTML(http.StatusInternalServerError,"sales_report.html",gin.H{"error":"Could not load any orders, please try again later"})
			return 
		}
	}


	type response struct {
		OrderID 		uint
		Date 			time.Time
		UserName 		string 
		Total 			float64
		Discount 		float64
		PaymentMethod 	string
		Status 			string
	}

	ResponseOrders := make([]response,len(orders))

	totalSales := 0.0
	totalOrders := int(total)
	totalDiscount := 0.0
	totalProducts := 0

	for i, order := range orders{
		var user models.User
		db.Db.Where("id = ?",order.UserID).First(&user)
		if order.Status == "delivered"{
			totalSales += order.TotalAmount
			totalDiscount += order.DiscountTotal

			for _, item := range order.OrderItems{
				totalProducts += item.Quantity
			}
		}

		ResponseOrders[i] = response{
			OrderID: order.ID,
			Date: order.OrderDate,
			UserName: user.Username,
			Total: order.TotalAmount,
			Discount: order.DiscountTotal,
			PaymentMethod: order.PaymentMethod,
			Status: order.Status,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	

	c.HTML(http.StatusOK,"sales_report.html",gin.H{
		"sales":ResponseOrders,
		"totalSales":totalSales,
		"totalOrders": totalOrders,
		"totalDiscount":totalDiscount,
		"totalProduct":totalProducts,
		"selectedRange":rangeType,
		"startDate":start.Format("2006-01-02"),
		"endDate":end.Format("2006-01-02"),
		"page": page,
		"totalPages":totalPages,
		"limit":limit,
	})

}

func DownloadSalesReport(c *gin.Context){
	
	from := c.PostForm("from")
	to := c.PostForm("to")
	// rangeType := c.PostForm("range")

	start,_ := time.Parse("2006-01-02",from)
	end,_ := time.Parse("2006-01-02",to)

	var orders []models.Order
	db.Db.Preload("OrderItems").Where("create_at BETWEEN ? AND ?",start,end).Find(&orders)

	pdf := gofpdf.New("P","mm","A4","")
	pdf.AddPage()
	pdf.SetFont("Arial","B",16)
	pdf.Cell(40,10,"Sales Report")

	pdf.SetFont("Arial","",12)
	pdf.Ln(10)
	pdf.Cell(60,10,"From: "+from)
	pdf.Cell(60,10,"To: "+to)
	pdf.Ln(12)

	// delete
	fmt.Println("Date check from :",from)
	fmt.Println("Date check to :",to)
	fmt.Println("--------------")
	fmt.Println("Default start :",start)
	fmt.Println("Default end :",end)
	fmt.Println("---------")


	for _, order := range orders {
		var user models.User
		db.Db.Where("id = ?",order.UserID).First(&user)

		// delete
		fmt.Printf("Order ID: %d | User: %s | Amount: ₹%.2f\n",order.ID,user.Username,order.TotalAmount)

		pdf.Cell(0,10,fmt.Sprintf("Order ID: %d | User: %s | Amount: ₹%.2f",order.ID,user.Username,order.TotalAmount))
		pdf.Ln(8)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)

	if err != nil{
		c.HTML(http.StatusInternalServerError,"sales_report.html",gin.H{"error":"Failed to generate PDF"})
		return 
	}

	c.Header("Content-Disposition","attachment; filename=sales_report.pdf")
	c.Data(http.StatusOK,"application/pdf",buf.Bytes())
}