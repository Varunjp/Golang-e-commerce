package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"first-project/utils"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterPage(c *gin.Context) {
	ref := c.Query("ref")
	c.HTML(http.StatusOK,"register.html",gin.H{"ReferralCode":ref})
}

func RegisterUser(c *gin.Context){
	
	var input struct {
		Username string		`form:"username" binding:"required"`
		Email	 string		`form:"email" binding:"required"`
		Password string		`form:"password" binding:"required"`
		Confirmpass string 	`form:"confirm_password" binding:"required"`
		Phone	 string		`form:"phone" binding:"required"`
	}

	referralCode := c.PostForm("referral_code")

	newreferralCode := helper.GenerateUniqueReferralCode()

	if err:= c.ShouldBind(&input); err != nil{
		c.HTML(http.StatusBadRequest,"register.html",gin.H{
			"error":"Invalid data",
		})
		return 
	}

	phonePattern := regexp.MustCompile(`^[0-9]{10}$`)

	if !phonePattern.MatchString(input.Phone){
		c.HTML(http.StatusBadRequest,"register.html",gin.H{
			"error":"Phone number must be exactly 10 digits",
		})
		return
	}

	if helper.IsSameDigitPhone(input.Phone){
		c.HTML(http.StatusBadRequest,"register.html",gin.H{
			"error":"Phone number cannot contain all same digits",
		})
		return
	}

	if !helper.IsValidPassword(input.Password){
		c.HTML(http.StatusBadRequest,"register.html",gin.H{
			"error":"Password must be at least 8 characters with uppercase, lowercase, number, and special character",
		})
		return
	}

	if input.Password != input.Confirmpass{
		c.HTML(http.StatusBadRequest,"register.html",gin.H{
			"error":"Passwords do not match",
		})
		return
	}

	hashedPassword,_ := utils.HashPassword(input.Password)
	user := models.User{
		Username: input.Username,
		Email: input.Email,
		Password: hashedPassword,
		Phone: input.Phone,
		Status: "inactive",
		ReferralCode: newreferralCode,
		Created_at: time.Now(),
	}

	if referralCode != ""{
		user.ReferredBy = referralCode
	}

	otp,_ := helper.GenerateAndSaveOtp(user.Email)

	err := helper.SendOTPEmail(user.Email,otp)

	
	if err != nil{
		c.HTML(http.StatusInternalServerError,"register.html",gin.H{"error":"Failed to generate Otp"})
		fmt.Println("Error :",err)
		return
	}

	if err := db.Db.Create(&user).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"register.html",gin.H{"error":"Failed to create account"})
		return 
	}

	c.HTML(http.StatusOK,"verifyOtp.html",gin.H{
		"email":user.Email,
	})
}


func VerfiyOTP (c *gin.Context){

	var input struct {
		Email	string	`form:"email" binding:"required"`
		OTP		string	`form:"otp" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil{
		
		c.HTML(http.StatusBadRequest,"verifyOtp.html",gin.H{
			"email":input.Email,
			"error":"Invalid otp",
		})

		return 
	}

	otpcheck, err := helper.VerfiyOTP(input.Email,input.OTP)


	if !otpcheck || err != nil{
		
		c.HTML(http.StatusBadRequest,"verifyOtp.html",gin.H{
			"email":input.Email,
			"error":"Invalid otp",
		})

		return 
	}

	var user models.User
	db.Db.Model(&models.User{}).Where("email = ?",input.Email).Update("status","Active")
	db.Db.Model(&models.User{}).Where("email = ?",input.Email).First(&user)

	Couperr := helper.CreateCoupon(user.ID)

	if Couperr != nil{
		c.HTML(http.StatusInternalServerError,"userLogin.html",gin.H{"error":err})
		return 
	}

	c.HTML(http.StatusOK,"userLogin.html",gin.H{
		"message":"Email verified. You can now log in.",
	})

}