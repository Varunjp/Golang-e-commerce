package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"first-project/utils"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-gomail/gomail"
)

func RegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK,"register.html",nil)
}

func RegisterUser(c *gin.Context){
	
	var input struct {
		Username string		`form:"username" binding:"required"`
		Email	 string		`form:"email" binding:"required"`
		Password string		`form:"password" binding:"required"`
		Phone	 string		`form:"phone" binding:"required"`
	}

	if err:= c.ShouldBind(&input); err != nil{
		c.HTML(http.StatusBadRequest,"register.html",gin.H{
			"error":"Invalid data",
		})
		return 
	}

	hashedPassword,_ := utils.HashPassword(input.Password)
	user := models.User{
		Username: input.Username,
		Email: input.Email,
		Password: hashedPassword,
		Status: "inactive",
		Created_at: time.Now(),
	}

	
	// need to continue

	otp,_ := helper.GenerateAndSaveOtp(user.Email)
	err := SendOTPEmail(user.Email,otp)


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

func SendOTPEmail(email, otp string) error{

	myMail := os.Getenv("Email")
	Password := os.Getenv("Password")

	msg := gomail.NewMessage()
	msg.SetHeader("From",myMail)
	msg.SetHeader("To",email)
	msg.SetHeader("Subject","Your OTP Code")
	msg.SetBody("text/plain","Your OTP code is: "+otp)

	d := gomail.NewDialer("smtp.gmail.com",587,myMail,Password)
	return d.DialAndSend(msg)
}

func VerfiyOTP (c *gin.Context){

	var input struct {
		Email	string	`form:"email" binding:"required"`
		OTP		string	`form:"otp" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil{
		
		c.HTML(http.StatusBadRequest,"verfiyOtp.html",gin.H{
			"email":input.Email,
			"error":"Invalid otp",
		})

		return 
	}

	otpcheck, err := helper.VerfiyOTP(input.Email,input.OTP)

	if !otpcheck || err != nil{
		
		c.HTML(http.StatusBadRequest,"verfiyOtp.html",gin.H{
			"email":input.Email,
			"error":"Invalid otp",
		})

		return 
	}

	db.Db.Model(&models.User{}).Where("email = ?",input.Email).Update("status","Active")


	c.HTML(http.StatusOK,"userLogin.html",gin.H{
		"message":"Email verified. You can now log in.",
	})

}