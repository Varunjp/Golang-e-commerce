package admin

import (
	db "first-project/DB"
	"first-project/models"
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
	

	c.HTML(http.StatusOK,"user_list.html",gin.H{
		"Users":users,
		"page":page,
		"limit":limit,
		"totalPages":totalPages,
	})
}

func FindUser (c *gin.Context){
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

	keyword := c.Query("search")

	if keyword != ""{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid data passed"})
		return
	}

	adDb := db.Db.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", "%"+strings.ToLower(keyword)+"%","%"+strings.ToLower(keyword)+"%").Order("id desc").Limit(limit).Offset(offset).Find(&users)

	if err := adDb.Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Database error"})
		return 
	}
	
	adDb.Count(&total)

	if total < 1 {
		c.JSON(http.StatusNotFound,gin.H{"message":"User not found"})
		return 
	}

	totalPages := int(math.Ceil(float64(total)/ float64(limit)))
	
	// Add html
	// c.JSON(http.StatusOK,gin.H{"users":users,"keyword":keyword})


	c.HTML(http.StatusOK,"user_list.html",gin.H{
		"Users":users,
		"page":page,
		"limit":limit,
		"totalPages":totalPages,
	})
	
}

func BlockUser(c *gin.Context){
	userID := c.Param("id")
	var user models.User

	if err := db.Db.First(&user, userID).Error; err != nil{
		c.JSON(http.StatusNotFound, gin.H{"error":"User not found"})
		return 
	}

	if user.Status == "Blocked"{
		c.JSON(http.StatusBadRequest, gin.H{"message":"User already blocked"})
		return 
	}

	user.Status = "Blocked"
	if err := db.Db.Save(&user).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error": "Could not save user"})
		return 
	}

	c.JSON(http.StatusOK,gin.H{"message": "User blocked successfully"})
}

func UnblockUser(c *gin.Context){
	userID := c.Param("id")
	var user models.User

	if err := db.Db.First(&user, userID).Error; err != nil{
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return 
	}

	if user.Status == "Available"{
		c.JSON(http.StatusBadRequest, gin.H{"message": "User already active"})
		return 
	}

	user.Status = "Available"
	if err := db.Db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save user"})
		return 
	}
	c.JSON(http.StatusOK,gin.H{"message":"User unblocked successfully"})
}