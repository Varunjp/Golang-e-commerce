package user

import (
	"context"
	"encoding/json"
	"errors"
	db "first-project/DB"
	"first-project/middleware"
	"first-project/models"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func HandleGoogleLogin(c *gin.Context){

	GetValues()

	url := GoogleOauthConfig.AuthCodeURL("random-state")
	c.Redirect(http.StatusTemporaryRedirect,url)

}

func HandleGoogleCallback(c *gin.Context){
	
	code := c.Query("code")
	if code == ""{
		c.HTML(http.StatusBadRequest,"userLogin.html",gin.H{"error":"No code in callback"})
		return 
	}

	token, err := GoogleOauthConfig.Exchange(context.Background(),code)

	if err != nil {
		c.HTML(http.StatusInternalServerError,"userLogin.html",gin.H{"error":"Token exchange failed"})
		return 
	}

	client := GoogleOauthConfig.Client(context.Background(),token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")

	if err != nil{
		c.HTML(http.StatusInternalServerError,"userLogin.html",gin.H{"error":"Failed to get user info"})
		return 
	}

	defer resp.Body.Close()

	var googleUser struct {
		ID 		string	`json:"id"`
		Email	string	`json:"email"`
		Name	string	`json:"name"`
		Picture	string	`json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil{
		c.HTML(http.StatusInternalServerError,"userLogin.html",gin.H{"error":"Decode failed"})
		return 
	}



	var user models.User
	
	if err := db.Db.Where("email = ?",googleUser.Email).First(&user).Error; err != nil{
		
		if errors.Is(err, gorm.ErrRecordNotFound){
			
			user = models.User{
				Username: googleUser.Name,
				Email: googleUser.Email,
				Status:	"Active",
				Created_at: time.Now(),
			}

			if err := db.Db.Create(&user).Error; err != nil{
				c.HTML(http.StatusInternalServerError,"userLogin.html",gin.H{"error":"Failed to create account"})
				return 
			}
		} else{
			c.HTML(http.StatusInternalServerError,"userLogin.html",gin.H{"error":"Unknow issue occured"})
			return
		} 
	}

	if user.Status != "Active"{
		c.HTML(http.StatusInternalServerError,"userLogin.html",gin.H{"error":"Account blocked contact adminstrator"})
		return
	}

	tokenString, err := middleware.CreateToken("user",user.Email,user.ID)

	if err != nil{
		c.HTML(http.StatusInternalServerError,"userLogin.html",gin.H{"error":"Failed to create JWT token"})
		return
	}

	session := sessions.Default(c)
	session.Set("name",user.Username)
	session.Save()

	c.SetCookie("JWT-User",tokenString,3600,"/","",false,true)

	c.Redirect(http.StatusFound,"/user/home")


}