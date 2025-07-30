package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ResendOTP(c *gin.Context) {

	email := c.PostForm("email")
	var user models.User

	if err := db.Db.Where("email = ?",email).First(&user).Error;err != nil{
		c.HTML(http.StatusBadRequest,"userLogin.html",gin.H{"error":"User not found","email":email})
		return 
	}
	
	go func(){
		otp, _ := helper.GenerateAndSaveOtp(email)

		err := helper.SendOTPEmail(user.Username,email,otp)
		
		if err != nil{
			log.Println(err)
		}
		
	}()
	

	c.HTML(http.StatusOK,"verifyOtp.html",gin.H{"message":"OTP resend to mail","email":email})
}