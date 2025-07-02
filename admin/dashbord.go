package admin

import (
	db "first-project/DB"
	"first-project/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetSalesData(c *gin.Context){
	type RequestQuery struct {
		Type 	string    `form:"type"`
	}

	var q RequestQuery

	if err := c.ShouldBindQuery(&q); err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid query parameter"})
		return 
	}

	var labels []string 
	var sales []float64

	now := time.Now()

	switch q.Type {
	case "daily":
		for i:= 29;i>=0;i--{
			day := now.AddDate(0,0,-i)
			labels = append(labels, day.Format("Jan 2"))

			var total float64

			db.Db.Model(&models.Order{}).Where("DATE(create_at)= ? AND status = ?",day.Format("2006-01-02"), "Delivered").Select("COALESCE(SUM(total_amount),0)").Scan(&total)
			sales = append(sales, total)
		}

		c.JSON(http.StatusOK,gin.H{"labels":labels,"sales":sales,"label":"Daily Sales (Last 30 Days)"})
		return 
	case "monthly":
		for i := 11; i >= 0; i--{
			monthTime := now.AddDate(0,-i,0)
			start := time.Date(monthTime.Year(),monthTime.Month(),1,0,0,0,0,time.UTC)
			end := start.AddDate(0,1,0)

			labels = append(labels, monthTime.Format("Jan 2006"))

			var total float64
			db.Db.Model(&models.Order{}).Where("create_at >= ? AND create_at < ? AND status = ?", start, end, "Delivered").Select("COALESCE(SUM(total_amount),0)").Scan(&total)

			sales = append(sales, total)
		}

		c.JSON(http.StatusOK,gin.H{
			"labels":labels,
			"sales":sales,
			"label":"Monthly Sales (Last 12 Months)",
		})
		return 

	case "yearly":
		for i := 0; i < 5; i++{
			yearTime := now.AddDate(-i,0,0)
			// start := time.Date(yearTime.Year(),yearTime.Month(),1,0,0,0,0,time.UTC)
			// end := start.AddDate(1,0,0)

			year := now.Year() - i

			
			start := time.Date(year, time.January, 1, 0, 0, 0, 0, now.Location())

			
			var end time.Time
			if i == 0 {
				
				end = now
			} else {
				
				end = time.Date(year+1, time.January, 1, 0, 0, 0, 0, now.Location())
			}

			labels = append(labels, yearTime.Format("2006"))
			var total float64
			db.Db.Model(&models.Order{}).Where("create_at >= ? AND create_at <= ? AND status = ?", start, end, "Delivered").Select("COALESCE(SUM(total_amount),0)").Scan(&total)
			sales = append(sales, total)
		}
		c.JSON(http.StatusOK,gin.H{
			"labels":labels,
			"sales":sales,
			"label":"Yearly Sales (Last 5 years)",
		})
	}

}