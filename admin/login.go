package admin

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/middleware"
	"first-project/models"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func LoginPage(c *gin.Context){

	session := sessions.Default(c)
	username := session.Get("name")

	var totalUser int64
	var totalProducts int64

	db.Db.Model(&models.User{}).Count(&totalUser)
	db.Db.Model(&models.Product_Variant{}).Count(&totalProducts)

	products,categories := helper.TopProductCategory()
	totalSales := helper.SalesReport()

	c.HTML(http.StatusOK,"admin_dashboard.html",gin.H{
		"username" : username.(string),
		"totalUsers": totalUser,
		"totalProducts": totalProducts,
		"totalSales": totalSales,
		"topProducts":   products,
    	"topCategories": categories,
	})

}

func Login(c *gin.Context){
	
	var admin models.Admin

	email := c.PostForm("email")
	password := c.PostForm("password")

	result := db.Db.Where("email=?",email).First(&admin)

	if result.Error != nil{
		c.HTML(http.StatusUnauthorized,"admin_login.html",gin.H{
			"error":"Invalid email or password",
		})
		return
	}

	if admin.Status == "Blocked"{
		c.HTML(http.StatusUnauthorized, "admin_login.html",gin.H{
			"error":"User has been blocked by admin",
		})
		return 
	}

	if admin.Password != password {
		c.HTML(http.StatusUnauthorized,"admin_login.html",gin.H{"error":"Invalid email or password"})
		return 
	}

	session := sessions.Default(c)
	session.Set("name",admin.Username)
	session.Save()

	token, err := middleware.CreateToken("admin",admin.Email,admin.ID)
	if err != nil{
		c.HTML(http.StatusOK,"admin_login.html",gin.H{"error": "Error Generating JWT"})
	}
	
	c.SetCookie("JWT-Admin",token,3600,"/","",false,true)
	
	// Prevent caching
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")

	c.Redirect(http.StatusTemporaryRedirect,"/admin")
}