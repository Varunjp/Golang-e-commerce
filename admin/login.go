package admin

import (
	db "first-project/DB"
	"first-project/middleware"
	"first-project/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoginPage(c *gin.Context){
	c.HTML(http.StatusOK,"admin_login.html",nil)
}

func Login(c *gin.Context){
	//var input models.Userinput
	var admin models.Admin

	email := c.PostForm("email")
	password := c.PostForm("password")

	// if err := c.ShouldBind(&input); err != nil{
	// 	c.JSON(http.StatusBadRequest,gin.H{"error": err.Error()})
	// 	return
	// }

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

	token, err := middleware.CreateToken("admin",admin.Email,admin.ID)
	if err != nil{
		c.JSON(http.StatusOK,gin.H{"error": "Error Generating JWT"})
	}
	c.Header("Authorization","Bearer"+token)

	username := admin.Username

	//c.JSON(http.StatusOK, gin.H{"message":"Login successfull", "token":token})

	c.HTML(http.StatusOK,"admin_dashboard.html",gin.H{
		"username" : username,
		"totalUsers": 10,
		"totalProducts": 100,
		"totalSales": 10000,
	})
}