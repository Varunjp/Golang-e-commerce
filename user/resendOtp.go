package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ResendOTP(c *gin.Context) {

	email := c.PostForm("email")
	var user models.User

	if err := db.Db.Where("email = ?",email).First(&user).Error;err != nil{
		c.HTML(http.StatusBadRequest,"userLogin.html",gin.H{"error":"User not found"})
		return 
	}

	otp, _ := helper.GenerateAndSaveOtp(email)

	err := helper.SendOTPEmail(email,otp)

	if err != nil{
		c.HTML(http.StatusConflict,"verifyOtp.html",gin.H{"error":"Failed to send OTP"})
		return 
	}

	c.HTML(http.StatusOK,"verifyOtp.html",gin.H{"message":"OTP resend to mail","email":email})
}