package admin

import (
	db "first-project/DB"
	"first-project/models"
	"first-project/models/responsemodels"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ListUsers(c *gin.Context){
	var users [] responsemodels.User

	session := sessions.Default(c)
	name := session.Get("name").(string)

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
	
	

	keyWord := c.Query("search")

	if keyWord != ""{
		param := "%"+keyWord+"%"
		AdDb.Where("username ILIKE ?",param)
	}

	AdDb.Count(&total)

	AdDb.Where("deleted_at IS NULL").Order("id desc").Limit(limit).Offset(offset).Find(&users)

	totalPages := int(math.Ceil(float64(total)/ float64(limit)))
	
	c.HTML(http.StatusOK,"user_list.html",gin.H{
		"users":users,
		"page":page,
		"limit":limit,
		"totalPages":totalPages,
		"user":name,
	})
}

func FindUser (c *gin.Context){
	var users [] responsemodels.User

	session := sessions.Default(c)
	name := session.Get("name").(string)

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

	adDb := db.Db.Where("deleted_at IS NULL AND LOWER(username) LIKE ? OR LOWER(email) LIKE ?", "%"+strings.ToLower(keyword)+"%","%"+strings.ToLower(keyword)+"%").Order("id desc").Limit(limit).Offset(offset).Find(&users)

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
		"user":name,
	})
	
}

func BlockUser(c *gin.Context){
	userID := c.Param("id")
	var user models.User

	if err := db.Db.First(&user, userID).Error; err != nil{
		c.HTML(http.StatusNotFound, "user_list.html", gin.H{"error":"User not found"})
		return 
	}

	if user.Status == "Blocked"{
		c.HTML(http.StatusConflict,"user_list.html",gin.H{"error":"User already blocked"})
		return 
	}

	user.Status = "Blocked"
	if err := db.Db.Save(&user).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error": "Could not save user"})
		return 
	}

	c.Redirect(http.StatusFound,"/admin/users-list")
}

func UnblockUser(c *gin.Context){
	userID := c.Param("id")
	var user models.User

	if err := db.Db.First(&user, userID).Error; err != nil{
		c.HTML(http.StatusNotFound, "user_list.html",gin.H{"error": "User not found"})
		return 
	}

	if user.Status == "Active"{
		c.HTML(http.StatusConflict,"user_list.html", gin.H{"message": "User already active"})
		return 
	}

	user.Status = "Active"
	if err := db.Db.Save(&user).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "user_list.html",gin.H{"error": "Could not save user"})
		return 
	}
	c.Redirect(http.StatusFound,"/admin/users-list")
}

func DeleteUser(c *gin.Context){
	userID := c.Param("id")
	var user models.User

	if err := db.Db.Where("id = ?",userID).First(&user).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"user_list.html",gin.H{"error":"Could not remove user"})
		return 
	}

	if err := db.Db.Delete(&user).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"user_list.html",gin.H{"error":"Could not remove user"})
		return 
	}
	c.Redirect(http.StatusSeeOther,"/admin/users-list")
}