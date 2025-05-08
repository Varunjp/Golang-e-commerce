package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"first-project/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ResetPasswordOTP(c *gin.Context){

	c.HTML(http.StatusOK,"resetPassword.html",nil)

}

func ResetPasswordOTPSend(c *gin.Context){

	email := c.PostForm("email")

	var user models.User
	if err := db.Db.Where("email = ?", email).First(&user).Error; err != nil{
		c.HTML(http.StatusBadRequest,"resetPassword.html",gin.H{
			"error": "User not found",
		})
		return 
	}

	otp,err := helper.GenerateAndSaveOtp(email)

	if err != nil{
		c.HTML(http.StatusBadRequest,"resetPassword.html",gin.H{
			"error": "Failed to create OTP",
		})
		return 
	}

	err = helper.SendOTPEmail(email,otp)

	if err != nil{
		c.HTML(http.StatusBadRequest,"resetPassword.html",gin.H{
			"error": "Failed to send otp",
		})
		return
	}

	c.Redirect(http.StatusFound,"/reset-password/verify-otp?email="+email)
}

func ResetPasswordOTPpage(c *gin.Context){
	
	email := c.Query("email")
	
	c.HTML(http.StatusOK,"resetpassword_verifyOtp.html",gin.H{"email":email})

}

func ResetPasswordOTPVerify(c * gin.Context){

	email := c.PostForm("email")
	enteredOTP := c.PostForm("otp")

	otpcheck,err := helper.VerfiyOTP(email,enteredOTP)

	if err != nil || !otpcheck{
		c.HTML(http.StatusBadRequest,"resetpassword_verifyOtp.html",gin.H{
			"error":"Invalid otp",
		})
		return 
	}

	c.Redirect(http.StatusFound,"/user/reset-password?email="+email)

}

func Resetpassword_ResendOTP(c *gin.Context){

	email := c.PostForm("email")
	var user models.User

	if err := db.Db.Where("email = ?",email).First(&user).Error;err != nil{
		c.HTML(http.StatusBadRequest,"resetpassword_verifyOtp.html",gin.H{"error":"User not found"})
		return 
	}

	otp, _ := helper.GenerateAndSaveOtp(email)

	err := helper.SendOTPEmail(email,otp)

	if err != nil{
		c.HTML(http.StatusConflict,"resetpassword_verifyOtp.html",gin.H{"error":"Failed to send OTP"})
		return 
	}

	c.HTML(http.StatusOK,"resetpassword_verifyOtp.html",gin.H{"message":"OTP resend to mail","email":email})

}

func ShowRestPasswordPage(c *gin.Context){
	
	c.HTML(http.StatusOK,"reset_password.html",gin.H{"email":c.Query("email")})

}

func ResetPassword(c *gin.Context){

	email := c.PostForm("email")
	newPassword := c.PostForm("password")

	hashed,_ := utils.HashPassword(newPassword)

	if err := db.Db.Model(&models.User{}).Where("email = ?",email).Update("password",hashed).Error; err != nil{
		c.HTML(http.StatusBadRequest,"reset_password.html",gin.H{"error": "Failed to reset password", "email": email})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/user/login")

}

