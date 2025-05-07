package admin

import (
	db "first-project/DB"
	"first-project/middleware"
	"first-project/models"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func LoginPage(c *gin.Context){

	session := sessions.Default(c)
	username := session.Get("name")

	c.HTML(http.StatusOK,"admin_dashboard.html",gin.H{
		"username" : username.(string),
		"totalUsers": 10,
		"totalProducts": 100,
		"totalSales": 10000,
	})

	//c.HTML(http.StatusOK,"admin_login.html",nil)
}

func Login(c *gin.Context){
	
	var admin models.Admin

	email := c.PostForm("email")
	password := c.PostForm("password")

	result := db.Db.Where("email=?",email).First(&admin)

	if result.Error != nil{
		c.JSON(http.StatusUnauthorized,gin.H{
			"message":"Invalid email or password",
		})
		return
	}

	if admin.Status == "Blocked"{
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":"User has been blocked by admin",
		})
		return 
	}

	if admin.Password != password {
		c.JSON(http.StatusUnauthorized, gin.H{"message":"Invalid email or password"})
		return 
	}

	session := sessions.Default(c)
	session.Set("name",admin.Username)
	session.Save()

	token, err := middleware.CreateToken("admin",admin.Email,admin.ID)
	if err != nil{
		c.JSON(http.StatusOK,gin.H{"error": "Error Generating JWT"})
	}
	
	c.SetCookie("JWT-Admin",token,3600,"/","",false,true)
	
	// Prevent caching
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")


	//c.JSON(http.StatusOK, gin.H{"message":"Login successfull", "token":token})

	c.Redirect(http.StatusTemporaryRedirect,"/admin")
}