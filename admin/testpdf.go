package admin

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/wcharczuk/go-chart/v2"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

func GenerateStyledSalesReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		startDate := time.Now().AddDate(0, 0, -30)
		endDate := time.Now()

		// Sample values (replace with DB queries)
		totalSales := 245000.00
		totalOrders := 1250
		totalCustomers := 900
		newCustomers := 230

		topProducts := []struct {
			Name       string
			UnitsSold  int
			Revenue    float64
		}{
			{"Men's Running Shoes", 530, 106000},
			{"Women's Handbag", 420, 84000},
			{"Wireless Earbuds", 390, 78000},
			{"Smartwatch X Pro", 360, 72000},
			{"Cotton T-Shirts Pack", 330, 66000},
		}

		// PDF INIT
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.SetMargins(15, 20, 15)
		pdf.AddPage()
		pdf.SetFont("Arial", "", 12)

		// HEADER (Logo + Title)
		pdf.ImageOptions("static/icons/logo.png", 10, 10, 30, 0, false, gofpdf.ImageOptions{}, 0, "")
		pdf.SetFont("Arial", "B", 18)
		pdf.CellFormat(0, 15, "Monthly Sales Report", "", 1, "C", false, 0, "")
		pdf.SetFont("Arial", "", 12)
		pdf.CellFormat(0, 8, fmt.Sprintf("Period: %s - %s", startDate.Format("02 Jan 2006"), endDate.Format("02 Jan 2006")), "", 1, "C", false, 0, "")
		pdf.Ln(5)

		// Draw a line
		pdf.SetDrawColor(100, 100, 100)
		pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
		pdf.Ln(10)

		// Section Box
		drawSectionBox(pdf, "1. Overview")
		pdf.Ln(4)

		addRow(pdf, "Total Sales", fmt.Sprintf("Rs.%.2f", totalSales))
		addRow(pdf, "Total Orders", fmt.Sprintf("%d", totalOrders))
		addRow(pdf, "Total Customers", fmt.Sprintf("%d", totalCustomers))
		addRow(pdf, "New Customers", fmt.Sprintf("%d", newCustomers))

		// Top Products Table
		pdf.Ln(8)
		drawSectionBox(pdf, "2. Top 5 Selling Products")
		pdf.Ln(4)

		pdf.SetFillColor(240, 240, 240)
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(80, 10, "Product Name", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 10, "Units Sold", "1", 0, "C", true, 0, "")
		pdf.CellFormat(50, 10, "Revenue", "1", 0, "C", true, 0, "")
		pdf.Ln(-1)

		pdf.SetFont("Arial", "", 12)
		for _, p := range topProducts {
			pdf.CellFormat(80, 10, p.Name, "1", 0, "", false, 0, "")
			pdf.CellFormat(40, 10, fmt.Sprintf("%d", p.UnitsSold), "1", 0, "C", false, 0, "")
			pdf.CellFormat(50, 10, fmt.Sprintf("Rs.%.2f", p.Revenue), "1", 0, "R", false, 0, "")
			pdf.Ln(-1)
		}

		GenerateChart()
		// Optional Chart (replace with actual chart image)
		pdf.Ln(10)
		drawSectionBox(pdf, "3. Visual Summary (Chart)")
		pdf.Ln(4)
		pdf.ImageOptions("static/chart-sample.png", 40, pdf.GetY(), 130, 0, false, gofpdf.ImageOptions{}, 0, "")
		pdf.Ln(60)

		// Footer
		pdf.SetY(280)
		pdf.SetFont("Arial", "I", 10)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 10, fmt.Sprintf("Generated on %s", time.Now().Format("02 Jan 2006 15:04")), "", 0, "C", false, 0, "")

		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "inline; filename=sales_report.pdf")
		if err := pdf.Output(c.Writer); err != nil {
			log.Println("Error in pdf :",err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		}
	}
}

// --- helper functions below ---

func drawSectionBox(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFillColor(200, 220, 255)
	pdf.SetTextColor(0, 70, 140)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, title, "", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

func addRow(pdf *gofpdf.Fpdf, label, value string) {
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(70, 8, label, "0", 0, "L", false, 0, "")
	pdf.CellFormat(0, 8, value, "0", 1, "L", false, 0, "")
}

func GenerateChart() {
	graph := chart.BarChart{
		Title: "Top Products Revenue",
		Height: 300,
		BarWidth: 50,
		Bars: []chart.Value{
			{Value: 106000, Label: "Running Shoes"},
			{Value: 84000, Label: "Handbag"},
			{Value: 78000, Label: "Earbuds"},
			{Value: 72000, Label: "Smartwatch"},
			{Value: 66000, Label: "T-Shirts"},
		},
	}

	f, _ := os.Create("static/chart-sample.png")
	defer f.Close()
	graph.Render(chart.PNG, f)
}
