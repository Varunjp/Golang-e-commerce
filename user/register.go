package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"first-project/utils"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterPage(c *gin.Context) {
	ref := c.Query("ref")
	tokenStr,err := c.Cookie("JWT-User")
	if err == nil{
		if tokenStr != ""{
			c.Redirect(http.StatusSeeOther,"/")
			return 
		}
	}
	c.HTML(http.StatusOK,"register.html",gin.H{"ReferralCode":ref})
}

func RegisterUser(c *gin.Context){
	Errors := make(map[string]string)
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
		c.JSON(http.StatusBadRequest,gin.H{
			"status":"error","error":"Invalid data",
		})
		return 
	}
	var testUser models.User
	if err := db.Db.Where("email = ?",input.Email).First(&testUser).Error; err == nil{
		Errors["email"] = "Email already exist"
	}

	phonePattern := regexp.MustCompile(`^[0-9]{10}$`)

	if !phonePattern.MatchString(input.Phone){
		Errors["phone"] = "Phone number must be exactly 10 digits"
	}

	if helper.IsName(input.Username){
		Errors["username"] = "Name cannot contain special characters"
	}

	if helper.IsSameDigitPhone(input.Phone){
		Errors["phone"] = "Phone number cannot contain all same digits"
	}

	if !helper.IsValidPassword(input.Password){
		Errors["password"] = "Password must be at least 8 characters with uppercase, lowercase, number, and special character"
	}

	if input.Password != input.Confirmpass{
		Errors["password"]="Passwords do not match"
		Errors["confirm_password"] = "Passwords do not match"
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest,gin.H{"status":"error","errors":Errors})
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

	if err := db.Db.Create(&user).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error","message":"Failed to create account"})
		return 
	}

	go func(){
		otp,_ := helper.GenerateAndSaveOtp(user.Email)

		err := helper.SendOTPEmail(user.Email,otp)

		if err != nil{
			log.Println("Error :",err)
		}
	}()



	c.JSON(http.StatusOK,gin.H{
		"redirect": "/user/verifyOtppage",
		"email":user.Email,
	})
}

func VerifyOTPPage(c *gin.Context){

	token,err := c.Cookie("JWT-User")

	if err == nil{
		if token != ""{
			c.Redirect(http.StatusSeeOther,"/")
			return 
		}
	}

	email := c.Query("email")
	c.HTML(http.StatusOK,"verifyOtp.html",gin.H{"email":email})
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