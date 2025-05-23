package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"first-project/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserProfilePage(c *gin.Context) {

	var User models.User

	tokenstr,err := c.Cookie("JWT-User")
	
	if err != nil{
		c.JSON(http.StatusUnauthorized,gin.H{"error":"JWT token not found in cookies"})
		return 
	}

	email,id,err := helper.DecodeJWT(tokenstr)

	if err != nil{
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest,gin.H{"error":"Error found while fetching email or id"})
		return 
	}


	if err := db.Db.Preload("Orders",func(db *gorm.DB)*gorm.DB{
		return db.Order("id DESC")
	}).Preload("Addresses").Preload("ProfileImages",func(db *gorm.DB)*gorm.DB{
		return db.Order("id DESC")
	}).Where("email = ? AND id = ?",email,id).First(&User).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"User not found"})
		return 
	}
	
	
	var image models.ProfileImage

	if err :=  db.Db.Where("user_id = ?",User.ID).First(&image).Error; err != nil{
		log.Println(err)
	}


	c.HTML(http.StatusOK,"user_profile.html",gin.H{
		"user": User,
		"Image" : image.ImageUrl,
		"Addresses": User.Addresses,
		"Orders": User.Orders,
	})

}

func EditProfilePage(c *gin.Context){

	var User models.User
	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)

	if err := db.Db.Preload("Addresses").Where("id = ?",id).First(&User).Error; err != nil{
		c.HTML(http.StatusNotFound,"user_profile.html",gin.H{"error":"User not found"})
		return 
	}

	c.HTML(http.StatusOK,"edit_profile.html",gin.H{
		"user":User,

	})

}

func UpdateProfile(c *gin.Context){
	
	NewName := c.PostForm("username")
	NewPhone := c.PostForm("phone")
	email := c.PostForm("email")
	var User models.User
	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)

	if err := db.Db.Where("id = ?",id).First(&User).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"edit_profile.html",gin.H{"error":"Failed to retrive user details"})
		return 
	}

	if NewName != ""{
		User.Username = NewName
	}

	if NewPhone != ""{
		User.Phone = NewPhone
	}

	if User.Email != email && email != ""{

		otp,otperr := helper.GenerateAndSaveOtp(email)

		if otperr != nil{
			c.HTML(http.StatusInternalServerError,"edit_profile.html",gin.H{"error":"Failed to generate otp"})
			return 
		}

		senterr := helper.SendOTPEmail(email,otp)

		if senterr != nil{
			c.HTML(http.StatusInternalServerError,"edit_profile.html",gin.H{"error":"Failed to send email"})
			return 
		}

		c.HTML(http.StatusOK,"changeEmail.html",gin.H{"user":User.Username,"Email":email,"name":User.Username,"phone":User.Phone})

	}else{

		if err := db.Db.Save(&User).Error; err != nil{
		c.HTML(http.StatusBadRequest,"edit_profile.html",gin.H{"error":"Failed to update user details"})
		return 
		}

		c.Redirect(http.StatusFound,"/user/profile")

	}

}

func UpdateEmail(c *gin.Context){

	NewName := c.PostForm("name")
	NewPhone := c.PostForm("phone")
	email := c.PostForm("email")
	otp := c.PostForm("otp")
	var User models.User
	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)

	if err := db.Db.Where("id = ?",id).First(&User).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"changeEmail.html",gin.H{"error":"Failed to retrive user details"})
		return 
	}

	res,err := helper.VerfiyOTP(email,otp)

	if !res || err != nil {
		c.HTML(http.StatusNotFound,"changeEmail.html",gin.H{"user":User.Username,"Email":email,"name":NewName,"phone":NewPhone,"error":"Invalid OTP entered"})
		return 
	}

	User.Username = NewName
	User.Phone = NewPhone
	User.Email = email

	if err := db.Db.Save(&User).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"user_profile.html",gin.H{"error":"Failed to update details in DB"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/user/profile")
}

func ResendEmailOtp(c *gin.Context){

	NewName := c.PostForm("name")
	NewPhone := c.PostForm("phone")
	email := c.PostForm("email")

	var User models.User
	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)

	if err := db.Db.Where("id = ?",id).First(&User).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"changeEmail.html",gin.H{"user":User.Username,"Email":email,"name":NewName,"phone":NewPhone,"error":"Failed to retrieve user details"})
		return 
	}

	otp,otperr := helper.GenerateAndSaveOtp(email)

	if otperr != nil{
		c.HTML(http.StatusInternalServerError,"changeEmail.html",gin.H{"user":User.Username,"Email":email,"name":NewName,"phone":NewPhone,"error":"Failed to generate otp"})
		return 
	}

	senterr := helper.SendOTPEmail(email,otp)

	if senterr != nil{
		c.HTML(http.StatusInternalServerError,"edit_profile.html",gin.H{"user":User.Username,"Email":email,"name":NewName,"phone":NewPhone,"error":"Error occured while senting otp, Please try again later"})
		return 
	}

	c.HTML(http.StatusOK,"changeEmail.html",gin.H{"user":User.Username,"Email":email,"name":NewName,"phone":NewPhone})


}

func UploadProfileImage(c *gin.Context){
	
	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)

	file, err := c.FormFile("profile_image")
	
	if err != nil{
		c.HTML(http.StatusBadRequest,"user_profile.html",gin.H{"error":"No file uploaded"})
		return 
	}

	uploadPath := "./static/images/profiles"

	if err := os.MkdirAll(uploadPath,os.ModePerm); err != nil {
		c.HTML(http.StatusInternalServerError,"user_profile.html",gin.H{"error":"Unable to assess of create path"})
		return 
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("user_%d_%d%s",int(id),time.Now().Unix(),ext)
	filePath := filepath.Join(uploadPath,filename)

	if err := c.SaveUploadedFile(file,filePath); err != nil{
		c.HTML(http.StatusInternalServerError,"user_profile.html",gin.H{"error":"Unable save file"})
		return 
	}

	if err := helper.UpdateUserImage(int(id),"static/images/profiles/"+filename); err != nil{
		c.HTML(http.StatusInternalServerError,"user_profile.html",gin.H{"error":"Failed to update image"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/user/profile")

}

func AddAddress(c *gin.Context){

	userID,_ := strconv.Atoi(c.PostForm("user"))
	line1 := c.PostForm("line1")
	line2 := c.PostForm("line2")
	country := c.PostForm("country")
	state := c.PostForm("state")
	city := c.PostForm("city")
	postalCode := c.PostForm("postal_code")

	address := models.Address{
		UserID: uint(userID),
		AddressLine1: line1,
		AddressLine2: line2,
		Country: country,
		State: state,
		City: city,
		PostalCode: postalCode,
	}

	if err := db.Db.Create(&address).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to add new address"})
		return 
	}

	c.Redirect(http.StatusFound,"/user/profile")

}

func EditAddress(c *gin.Context){

	var address models.Address

	AddressID := c.PostForm("address_id")
	line1 := c.PostForm("line1")
	line2 := c.PostForm("line2")
	country := c.PostForm("country")
	state := c.PostForm("state")
	city := c.PostForm("city")
	postalCode := c.PostForm("postal_code")

	if err := db.Db.Where("address_id = ?",AddressID).First(&address).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Address not found"})
		return 
	}

	address.AddressLine1 = line1
	address.AddressLine2 = line2
	address.Country = country
	address.State = state
	address.City = city
	address.PostalCode =  postalCode

	if err := db.Db.Save(&address).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to update address"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/user/profile")

}

func ChangePasswordPage(c *gin.Context){
	
	session := sessions.Default(c)
	username,err := session.Get("name").(string)

	if !err {
		c.HTML(http.StatusInternalServerError,"change_password.html",gin.H{"error":"Error while fetching user name"})
		return 
	}

	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)
	var user models.User

	if err := db.Db.First(&user,id).Error; err != nil {
		c.HTML(http.StatusInternalServerError,"change_password.html",gin.H{"error":"Failed to load user details. Please try again later"})
		return 
	}

	if user.Password != ""{
		c.HTML(http.StatusOK,"change_password.html",gin.H{"user":username,"hasPassword":true})
	}else{
		c.HTML(http.StatusOK,"change_password.html",gin.H{"user":username})
	}	

}

func ChangePassword(c *gin.Context){

	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	if newPassword != confirmPassword {
		c.HTML(http.StatusBadRequest,"change_password.html",gin.H{"error":"New Password mismatch"})
		return 
	}

	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)
	var user models.User

	if err := db.Db.First(&user,id).Error; err != nil {
		c.HTML(http.StatusInternalServerError,"change_password.html",gin.H{"error":"Failed to load user details. Please try again later"})
		return 
	}

	if user.Password != ""{
		
		if !utils.CheckPasswordHash(currentPassword,user.Password){
		c.HTML(http.StatusConflict,"change_password.html",gin.H{"error":"Incorrect old password"})
		return
		}

	}

	

	hashedPass, hasherr := utils.HashPassword(newPassword)

	if hasherr != nil {
		c.HTML(http.StatusConflict,"change_password.html",gin.H{"error":"Failed to generator hash of new password"})
		return
	}

	user.Password = hashedPass

	if err := db.Db.Save(&user).Error; err != nil{
		c.HTML(http.StatusConflict,"change_password.html",gin.H{"error":"Failed to update new password"})
		return
	}

	c.Redirect(http.StatusFound,"/user/profile")

}	



func DeleteAddress(c *gin.Context){

	addressID := c.PostForm("address_id")

	var address models.Address

	if err := db.Db.Delete(&address,addressID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Address not found"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/user/profile")

}