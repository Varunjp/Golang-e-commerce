package user

import (
	db "first-project/DB"
	"first-project/middleware"
	"first-project/models"
	"first-project/utils"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Login(c * gin.Context){
	
	var input struct {
		Email 		string `form:"email" binding:"required,email"`
		Password	string `form:"password" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil{
		c.HTML(http.StatusBadRequest,"userLogin.html",gin.H{"error":err.Error()})
		return 
	}

	var user models.User
	if err := db.Db.Where("email = ?",input.Email).First(&user).Error; err != nil{
		c.HTML(http.StatusBadRequest,"userLogin.html",gin.H{"error":"User not found"})
		return 
	}

	if !utils.CheckPasswordHash(input.Password,user.Password){
		c.HTML(http.StatusUnauthorized,"userLogin.html",gin.H{"error":"Invalid email or password"})
		return
	}

	session := sessions.Default(c)
	session.Set("name",user.Username)
	session.Save()

	token, err := middleware.CreateToken("user",user.Email,user.ID)
	
	if err != nil{
		c.HTML(http.StatusUnauthorized,"userLogin.html",gin.H{"error": err.Error()})
		return
	}
	
	c.SetCookie("JWT",token,3600,"/","",false,true)

	// Prevent caching
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")

	c.JSON(http.StatusOK,gin.H{"message":"login successfull"})
}