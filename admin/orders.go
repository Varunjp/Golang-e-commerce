package admin

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	orderId := c.Query("order_id")
	username := c.Query("user_name")
	startdate := c.Query("start_date")
	enddate := c.Query("end_date")
	orderStatus := c.Query("status")

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

	// adding query - filters

	dbQuery := db.Db.Preload("OrderItems",func(db *gorm.DB) *gorm.DB{
		return db.Unscoped()
	}).Model(&models.Order{}).Order("create_at DESC")
	
	

	if orderId != ""{
		dbQuery =dbQuery.Where("id = ?",orderId)
	}

	if username != ""{
		dbQuery = dbQuery.Joins("JOIN users ON users.id = orders.user_id").Where("users.username ILIKE ?","%"+username+"%")
	}

	if startdate != "" && enddate != ""{
		dbQuery = dbQuery.Where("orders.create_at BETWEEN ? AND ?", startdate+" 00:00:00", enddate+" 23:59:59")
	}else if startdate != ""{
		dbQuery = dbQuery.Where("orders.create_at >= ?", startdate+" 00:00:00")
	}else if enddate != ""{
		dbQuery = dbQuery.Where("orders.create_at <= ?", enddate+" 23:59:59")
	}

	if orderStatus != ""{
		dbQuery = dbQuery.Where("orders.status = ?",orderStatus)
	}

	//

	dbQuery.Count(&total)

	if err := dbQuery.Order("id DESC").Offset(offset).Limit(limit).Find(&Orders).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_orders.html",gin.H{"error":"Failed to load orders list, please try again later"})
		return 
	}

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

	tokenStr,_ := c.Cookie("JWT-Admin")
	_,userId,_ := helper.DecodeJWT(tokenStr)
	var AdminUser models.Admin

	if err := db.Db.Where("id = ?",userId).First(&AdminUser).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Please login again"})
		return 
	}

	name := AdminUser.Username

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.HTML(http.StatusOK,"admin_orders.html",gin.H{
		"user":name,
		"orders":responseOrder,
		"page":page,
		"totalPages":totalPages,
		"limit":limit,
		"Filters": gin.H{
			"OrderID": orderId,
			"UserName": username,
			"StartDate": startdate,
			"EndDate": enddate,
			"Status": orderStatus,
		},
	})
}

func AdminOrderCancel(c *gin.Context){
	
	orderId,_ := strconv.Atoi(c.Param("id"))
	reason := c.PostForm("reason")
	var order models.Order

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderId).First(&order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_orders.html",gin.H{"error":"Failed to load order details"})
		return 
	}

	order.Reason = reason
	order.Status = "Returned"

	if err := db.Db.Save(&order).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_orders.html",gin.H{"error":"Failed to update order"})
		return 
	}

	for _,item := range order.OrderItems{

		db.Db.Model(&models.Product_Variant{}).Where("id = ?",item.ProductID).Update("stock",gorm.Expr("stock + ?",item.Quantity))

	}

	c.Redirect(http.StatusSeeOther,"/admin/orders")

}