package admin

import (
	"bytes"
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
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
		end = now
	case "monthly":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = now
	case "yearly":
		start = time.Date(now.Year(),1,1,0,0,0,0,now.Location())
		end = now
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
	var totalorders []models.Order

	dbFullordes := db.Db.Model(&models.Order{}).Where("create_at BETWEEN ? AND ?",start,end)

	dbFullordes.Count(&total)
	dbFullordes.Preload("OrderItems").Find(&totalorders)
	totalSales := 0.0
	totalDiscount := 0.0
	totalProducts := 0
	for _,item := range totalorders{
		if item.Status == "Delivered"{
			totalSales += item.TotalAmount
			totalDiscount += item.DiscountTotal
			for _, pitem := range item.OrderItems{
				if pitem.Status == "Delivered"{
					totalProducts += pitem.Quantity
				}
				
			}
		}
	}

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

	
	totalOrders := int(total)
	

	for i, order := range orders{
		var user models.User
		db.Db.Where("id = ?",order.UserID).First(&user)

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
	
	tokenStr,_ := c.Cookie("JWT-Admin")
	_,userId,_ := helper.DecodeJWT(tokenStr)
	var AdminUser models.Admin

	if err := db.Db.Where("id = ?",userId).First(&AdminUser).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Please login again"})
		return 
	}

	name := AdminUser.Username

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
		"start":start.Format("2006-01-02"),
		"end":end.Format("2006-01-02"),
		"filter":rangeType,
		"user":name,
	})

}

func DownloadSalesReport(c *gin.Context){
	
	from := c.PostForm("from")
	to := c.PostForm("to")
	//rangeType := c.PostForm("filter")

	start,_ := time.Parse("2006-01-02",from)
	end,_ := time.Parse("2006-01-02",to)

	var orders []models.Order
	db.Db.Preload("OrderItems").Where("create_at BETWEEN ? AND ? AND status = ?",start,end,"Delivered").Find(&orders)

	pdf := gofpdf.New("P","mm","A4","")
	pdf.AddPage()

	pdf.SetFont("Arial","B",16)

	title := "Sales Report"

	pageWidth,_ := pdf.GetPageSize()
	stringWidth := pdf.GetStringWidth(title)

	pdf.SetX((pageWidth - stringWidth)/2)
	pdf.CellFormat(stringWidth,10,title,"",1,"C",false,0,"")

	pdf.Ln(10)
	pdf.SetFont("Arial","B",12)
	pdf.CellFormat(25,10,"Company","1",0,"C",false,0,"")
	pdf.SetFont("Arial","",12)
	pdf.CellFormat(35,10,"FashionArt","1",0,"C",false,0,"")
	pdf.SetFont("Arial","B",12)
	pdf.CellFormat(20,10,"Email ID:","1",0,"C",false,0,"")
	pdf.SetFont("Arial","",12)
	pdf.CellFormat(55,10,"fashionartify@gmail.com","1",1,"C",false,0,"")
	pdf.SetFont("Arial","B",12)
	pdf.CellFormat(30,10,"From","1",0,"C",false,0,"")
	pdf.SetFont("Arial","",12)
	pdf.CellFormat(30,10,from,"1",0,"C",false,0,"")
	pdf.SetFont("Arial","B",12)
	pdf.CellFormat(30,10,"To","1",0,"C",false,0,"")
	pdf.SetFont("Arial","",12)
	pdf.CellFormat(45,10,to,"1",1,"C",false,0,"")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(30, 10, "Order ID", "1", 0, "C", false, 0, "")
	pdf.CellFormat(35, 10, "Date", "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 10, "Customer", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Amount", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Discount", "1", 1, "C", false, 0, "") // end of row

	pdf.SetFont("Arial", "", 11)

	if len(orders) < 1{
		pdf.CellFormat(30, 10, "---------", "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 10, "---------", "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 10, "No Sales", "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 10, "---------", "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 10, "---------", "1", 1, "C", false, 0, "")
	}

	var totalAmount float64

	for _, order := range orders {
		var user models.User
		db.Db.Where("id = ?",order.UserID).First(&user)
		orderDate := order.CreateAt.Format("02-01-2006")

		pdf.CellFormat(30, 10, fmt.Sprintf("%d", order.ID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 10, orderDate, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 10, user.Username, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("Rs. %.2f", order.TotalAmount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("Rs. %.2f", order.DiscountTotal), "1", 1, "C", false, 0, "")
		totalAmount += order.TotalAmount
	}

	pdf.Ln(10) // some space before total

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(105, 10, "Grand Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("Rs. %.2f", totalAmount), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "", "1", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)

	if err != nil{
		c.HTML(http.StatusInternalServerError,"sales_report.html",gin.H{"error":"Failed to generate PDF"})
		return 
	}

	c.Header("Content-Disposition","attachment; filename=sales_report.pdf")
	c.Data(http.StatusOK,"application/pdf",buf.Bytes())
}

func DownloadExcel(c *gin.Context){
	from := c.PostForm("from")
	to := c.PostForm("to")

	start,_ := time.Parse("2006-01-02",from)
	end,_ := time.Parse("2006-01-02",to)

	var orders []models.Order
	db.Db.Preload("OrderItems").Where("create_at BETWEEN ? AND ? AND status = ?",start,end,"Delivered").Find(&orders)

	f := excelize.NewFile()
	sheet := "SalesReport"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	// --- Title ---
	title := "Sales Report"
	f.SetCellValue(sheet, "A1", title)
	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.MergeCell(sheet, "A1", "E1")
	f.SetCellStyle(sheet, "A1", "E1", styleTitle)

	// --- Company Info ---
	info := [][]string{
		{"Company", "FashionArt", "Email ID:", "fashionartify@gmail.com"},
		{"From", from, "To", to},
	}

	borderStyle, _ := f.NewStyle(&excelize.Style{
	Border: []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
	},
	Alignment: &excelize.Alignment{Horizontal: "center"},})

	styleInfoBold, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
	})
	rowIdx := 3
	for _, row := range info {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
			f.SetCellValue(sheet, cell, val)
			f.SetCellStyle(sheet,cell,cell,borderStyle)
			if colIdx%2 == 0 {
				f.SetCellStyle(sheet, cell, cell, styleInfoBold)
			}
		}
		rowIdx++
	}

	// --- Header Row (start from row 6) ---
	headers := []string{"Order ID", "Date", "Customer", "Amount", "Discount"}
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
	})
	startRow := 6
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, startRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, styleHeader)
	}

	// --- Orders ---
	var totalAmount float64
	if len(orders) < 1 {
		values := []interface{}{"------", "------", "No Sales", "------", "------"}
		for col, v := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, startRow+1)
			f.SetCellValue(sheet, cell, v)
		}
	} else {
		for row, order := range orders {
			var user models.User
			db.Db.First(&user, order.UserID)

			values := []interface{}{
				order.ID,
				order.CreateAt.Format("02-01-2006"),
				user.Username,
				fmt.Sprintf("Rs. %.2f", order.TotalAmount),
				fmt.Sprintf("Rs. %.2f", order.DiscountTotal),
			}

			for col, v := range values {
				cell, _ := excelize.CoordinatesToCellName(col+1, startRow+1+row)
				f.SetCellValue(sheet, cell, v)
				f.SetCellStyle(sheet,cell,cell,borderStyle)
			}
			totalAmount += order.TotalAmount
		}
	}

	// --- Grand Total ---
	if len(orders) > 0 {
		totalRow := startRow + len(orders) + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "Grand Total")
		f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("C%d", totalRow))

		styleTotal, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true},
			Alignment: &excelize.Alignment{Horizontal: "right"},
		})
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("C%d", totalRow), styleTotal)

		amountCell := fmt.Sprintf("D%d", totalRow)
		f.SetCellValue(sheet, amountCell, fmt.Sprintf("Rs. %.2f", totalAmount))
		f.SetCellStyle(sheet, amountCell, amountCell, styleTotal)
	}

	// --- Adjust Column Width ---
	for i := 1; i <= len(headers); i++ {
		col, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheet, col, col, 20)
	}

	c.Header("Content-Disposition","attachment; filename=sales_report.xlsx")
	c.Header("Content-Type","application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	if err := f.Write(c.Writer); err != nil{
		c.JSON(500,"Error generating Excel file")
	}

}