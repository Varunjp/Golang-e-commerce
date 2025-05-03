package admin

import (
	db "first-project/DB"
	"first-project/models/responsemodels"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func ListUsers(c *gin.Context){
	var users [] responsemodels.User

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

	if err := db.Db.Model(&responsemodels.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to retrive users"})
		return 
	}

	AdDb := db.Db.Model(&responsemodels.User{})
	
	AdDb.Count(&total)

	AdDb.Order("id desc").Limit(limit).Offset(offset).Find(&users)

	totalPages := int(math.Ceil(float64(total)/ float64(limit)))
	
	// if err := db.Db.Find(&users).Error; err != nil{
	// 	c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to retrive users"})
	// 	return 
	// }

	// Add html

	// c.JSON(http.StatusOK,gin.H{
	// 	"Users":users,
	// 	"page":page,
	// 	"limit":limit,
	// 	"totalPages":totalPages,
	// })

	c.HTML(http.StatusOK,"user_list.html",gin.H{
		"Users":users,
		"page":page,
		"limit":limit,
		"totalPages":totalPages,
	})
}

func FindUser (c *gin.Context){
	var users [] responsemodels.User

	keyword := c.Query("search")

	if keyword != ""{
		if err := db.Db.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", "%"+strings.ToLower(keyword)+"%","%"+strings.ToLower(keyword)+"%").Find(&users).Error; err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"Database error"})
			return 
		}
	
		// Add html
		c.JSON(http.StatusOK,gin.H{"users":users,"keyword":keyword})
	}
	
}