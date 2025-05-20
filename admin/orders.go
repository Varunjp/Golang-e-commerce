package admin

import (
	db "first-project/DB"
	"first-project/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AdminOrdersPage(c *gin.Context){

	var Orders []models.Order
	type Response struct {
		ID				uint
		UserName		string
		TotalPrice		float64
		ItemCount		int
		CreatedAt		time.Time
		Status 			string
	}

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

	var total int

	if err := db.Db.Preload("OrderItems").Order("id DESC").Offset(offset).Limit(limit).Find(&Orders).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_orders.html",gin.H{"error":"Failed to load orders list, please try again later"})
		return 
	}

	total = len(Orders)

	if total == 0 {
		c.HTML(http.StatusNotFound,"admin_orders.html",gin.H{"error":"No orders to be found"})
		return 
	}

	responseOrder := make([]Response,len(Orders))

	for i, order := range Orders{

		itemcount := 0

		for _,item := range order.OrderItems{
			itemcount += item.Quantity
		}

		var User models.User

		if err := db.Db.Where("id = ?",order.UserID).First(&User).Error; err != nil{
			log.Println("Failed to fetch user details")
		}

		responseOrder[i] = Response{
			ID: order.ID,
			UserName: User.Username,
			TotalPrice: order.TotalAmount,
			ItemCount: itemcount,
			CreatedAt: order.OrderDate,
			Status: order.Status,
		}

	}

	session := sessions.Default(c)
	name := session.Get("name").(string)

	totalPages := int(math.Ceil(float64(total)/ float64(limit)))

	c.HTML(http.StatusOK,"admin_orders.html",gin.H{
		"user":name,
		"orders":responseOrder,
		"page":page,
		"totalPages":totalPages,
		"limit":limit,
	})
}