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


func LoginPage(c *gin.Context){
	tokenstr,err := c.Cookie("JWT-User")
	session := sessions.Default(c)

	message := session.Get("success")
	if message != nil {
		session.Delete("success")
		session.Save()
	}

	if err == nil{
		if tokenstr != ""{
			c.Redirect(http.StatusSeeOther,"/")
			return 
		}
	}
	c.HTML(http.StatusOK,"userLogin.html",gin.H{
		"message": message,})
}

func Login(c * gin.Context){
	
	var input struct {
		Email 		string `form:"email" binding:"required"`
		Password	string `form:"password" binding:"required"`
	}

	tokenstr,err := c.Cookie("JWT-User")

	if err == nil{
		if tokenstr != ""{
			c.Redirect(http.StatusSeeOther,"/")
			return 
		}
	}

	if err := c.ShouldBind(&input); err != nil{
		c.HTML(http.StatusBadRequest,"userLogin.html",gin.H{"error":err.Error()})
		return 
	}

	var user models.User
	if err := db.Db.Where("email = ? AND deleted_at IS NULL",input.Email).First(&user).Error; err != nil{
		c.HTML(http.StatusBadRequest,"userLogin.html",gin.H{"error":"User not found"})
		return 
	}

	if user.Status != "Active"{
		c.HTML(http.StatusBadRequest,"userLogin.html",gin.H{"error":"Account blocked contact administrator"})
		return 
	}

	if !utils.ChecKPasswordHash(input.Password,user.Password){
		c.HTML(http.StatusUnauthorized,"userLogin.html",gin.H{"error":"Invalid email or password"})
		return
	}

	session := sessions.Default(c)
	session.Set("name",user.Username)
	session.Save()

	token, err := middleware.CReateToken("user",user.Email,user.ID)
	
	if err != nil{
		c.HTML(http.StatusUnauthorized,"userLogin.html",gin.H{"error": err.Error()})
		return
	}
	
	c.SetCookie("JWT-User",token,3600,"/","",false,true)

	// Prevent caching
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")

	c.Redirect(http.StatusFound,"/")
}