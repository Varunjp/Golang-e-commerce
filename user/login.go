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
	c.HTML(http.StatusOK,"userLogin.html",nil)
}

func Login(c * gin.Context){
	
	var input struct {
		Email 		string `form:"email" binding:"required"`
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

	if user.Status != "Active"{
		c.HTML(http.StatusBadRequest,"userLogin.html",gin.H{"error":"Account blocked contact administrator"})
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
	
	c.SetCookie("JWT-User",token,3600,"/","",false,true)

	// Prevent caching
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")

	c.Redirect(http.StatusFound,"/")

}