package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"first-project/utils"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

	_,id,err := helper.DecodeJWT(tokenstr)

	if err != nil{
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest,gin.H{"error":"Error found while fetching email or id"})
		return 
	}


	if err := db.Db.Preload("Addresses").Preload("ProfileImages",func(db *gorm.DB)*gorm.DB{
		return db.Order("id DESC")
	}).Where("id = ?",id).First(&User).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"User not found"})
		return 
	}
	
	
	var image models.ProfileImage
	var imageUrl string 

	if err :=  db.Db.Where("user_id = ?",User.ID).First(&image).Error; err != nil{
		log.Println(err)
	}

	if image.ImageUrl != ""{
		imageUrl = image.ImageUrl
	}else{
		imageUrl = "static/images/dummy-profile-pic-300x300.png"
	}

	var wallet models.Wallet

	if err := db.Db.Where("user_id = ?",id).First(&wallet).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			errCreate := helper.CreateWallet(uint(id))
			if errCreate == nil{
				db.Db.Where("user_id = ?",id).First(&wallet)
			}else{
				c.HTML(http.StatusInternalServerError,"user_profile.html",gin.H{"error":"Failed to load wallet details, please try again later"})
				return
			}
		}else{
			c.HTML(http.StatusInternalServerError,"user_profile.html",gin.H{"error":"Failed to load wallet details, please try again later"})
			return 
		}
	}

	pageStr := c.Query("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	var orders []models.Order
	var totalCount int64

	db.Db.Model(&models.Order{}).Where("user_id = ?", id).Count(&totalCount)
	db.Db.Where("user_id = ?", id).Order("id DESC").Limit(limit).Offset(offset).Find(&orders)

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	session := sessions.Default(c)
	flash := session.Get("flash")
	sussess := session.Get("profilesuccess")

	if sussess != nil{
		session.Delete("profilesuccess")
		session.Save()
	}

	if flash != nil{
		session.Delete("flash")
		session.Save()
		c.HTML(http.StatusOK,"user_profile.html",gin.H{
			"user": User,
			"Image" : imageUrl,
			"Addresses": User.Addresses,
			"Balance":wallet.Balance,
			"error":flash,
			"Orders": orders,
			"CurrentPage": page,
			"TotalPages":  totalPages,
		})
		return 
	}

	c.HTML(http.StatusOK,"user_profile.html",gin.H{
		"user": User,
		"Image" : imageUrl,
		"Addresses": User.Addresses,
		"Balance":wallet.Balance,
		"Orders": orders,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"message": sussess,
	})

}

func EditProfilePage(c *gin.Context){

	var User models.User
	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)
	session := sessions.Default(c)
	errmsg := session.Get("flash")

	if err := db.Db.Preload("Addresses").Where("id = ?",id).First(&User).Error; err != nil{
		c.HTML(http.StatusNotFound,"user_profile.html",gin.H{"error":"User not found"})
		return 
	}

	if errmsg != nil{
		session.Delete("flash")
		session.Save()
		c.HTML(http.StatusOK,"edit_profile.html",gin.H{
			"user":User,
			"error":errmsg,
		})
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
	Errors := make(map[string]string)

	if strings.TrimSpace(NewName) == "" {
		Errors["username"] = "Please provide correct name"
	}

	if strings.TrimSpace(email) == ""{
		Errors["email"] = "Please provide email"
	}

	if strings.TrimSpace(NewPhone) == ""{
		Errors["phone"] = "Please provide phone number"
	}

	if helper.IsName(NewName){
		Errors["username"] = "Name cannot contain special characters"
	}

	phonePattern := regexp.MustCompile(`^[0-9]{10}$`)

	if !phonePattern.MatchString(NewPhone){
		Errors["phone"] = "Phone number must be exactly 10 digits"
	} 

	if helper.IsSameDigitPhone(NewPhone){
		Errors["phone"] = "Phone number cannot contain all same digits"
	}

	if len(Errors) > 0{
		c.JSON(http.StatusBadRequest,gin.H{"status":"error","errors":Errors})
		return 
	}

	if err := db.Db.Where("id = ?",id).First(&User).Error; err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"status":"error","message":"Failed to get user details"})
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
			c.JSON(http.StatusBadRequest,gin.H{"status":"error"})
			
			return 
		}

		senterr := helper.SendOTPEmail(NewName,email,otp)

		if senterr != nil{
			c.JSON(http.StatusBadRequest,gin.H{"status":"error"})
			return 
		}

		redirect := fmt.Sprintf("/user/changeEmail?email=%s&name=%s&phone=%s",email,NewName,NewPhone)

		c.JSON(http.StatusOK,gin.H{"redirect":redirect})

	}else{

		if err := db.Db.Save(&User).Error; err != nil{
		c.HTML(http.StatusBadRequest,"edit_profile.html",gin.H{"error":"Failed to update user details"})
		return 
		}

		c.JSON(http.StatusOK,gin.H{"redirect": "/user/profile"})
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

	session := sessions.Default(c)

	if strings.TrimSpace(NewName) == "" || strings.TrimSpace(NewPhone) == "" || strings.TrimSpace(email) == ""{
		session.Set("flash","Invalid content passed")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/edit-profile")
		return 
	}

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

	senterr := helper.SendOTPEmail(NewName,email,otp)

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
	session := sessions.Default(c)

	if err != nil{
		session.Set("flash","No file uploaded")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/profile")
		return
	}

	// Open the uploaded file
	openedFile, err := file.Open()
	if err != nil {
		session.Set("flash","Unable to open file")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/profile")
		return
	}
	defer openedFile.Close()

	// Read the first 512 bytes to detect the content type
	buffer := make([]byte, 512)
	_, err = openedFile.Read(buffer)
	if err != nil {
		session.Set("flash","Unable to read file")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/profile")
		return
	}

	// Detect content type (MIME type)
	contentType := http.DetectContentType(buffer)

	// Check if it's an image
	if !strings.HasPrefix(contentType, "image/") {
		session.Set("flash","Only image files are allowed")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/profile")
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
		session.Set("flash","Unable save file")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/profile")
		return
	}

	if err := helper.UpdateUserImage(int(id),"static/images/profiles/"+filename); err != nil{
		session.Set("flash","Failed to update image")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/profile")
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
	session := sessions.Default(c)

	if strings.TrimSpace(line1) == "" || strings.TrimSpace(line2) == "" || strings.TrimSpace(country) == "" || strings.TrimSpace(state) == "" || strings.TrimSpace(city) == "" || strings.TrimSpace(postalCode) == ""{	
		session.Set("flash","Invalid content passed")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/edit-profile")
		return 
	}

	if len(postalCode) != 6{
		session.Set("flash","Invalid postal code")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/edit-profile")
		return
	}

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

	session := sessions.Default(c)

	if err := db.Db.Where("address_id = ?",AddressID).First(&address).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Address not found"})
		return 
	}

	if strings.TrimSpace(line1) == "" || strings.TrimSpace(line2) == "" || strings.TrimSpace(country) == "" || strings.TrimSpace(state) == "" || strings.TrimSpace(city) == "" || strings.TrimSpace(postalCode) == ""{	
		session.Set("flash","Invalid content passed")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/edit-profile")
		return 
	}

	if len(postalCode) != 6{
		session.Set("flash","Invalid postal code")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/edit-profile")
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

	passerr := session.Get("password_error")

	if passerr != nil{
		session.Delete("password_error")	
		session.Save()
	}

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
		c.HTML(http.StatusOK,"change_password.html",gin.H{"user":username,"hasPassword":true,"error":passerr})
	}else{
		c.HTML(http.StatusOK,"change_password.html",gin.H{"user":username,"error":passerr})
	}	

}

func ChangePassword(c *gin.Context){

	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")
	session := sessions.Default(c)
	if strings.TrimSpace(currentPassword) == "" || strings.TrimSpace(newPassword) == "" || strings.TrimSpace(confirmPassword) == ""{
		session.Set("password_error","Please fill all fields")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/change-password")
		return
	}

	if newPassword != confirmPassword {
		session.Set("password_error","New Password mismatch")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/change-password")
		return
	}

	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)
	var user models.User

	if err := db.Db.First(&user,id).Error; err != nil {
		session.Set("password_error","Failed to load user details. Please try again later")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/change-password")
		return
	}

	if user.Password != ""{
		
		if !utils.CheckPasswordHash(currentPassword,user.Password){
			session.Set("password_error","Incorrect old password")
			session.Save()
			c.Redirect(http.StatusSeeOther,"/user/change-password")
			return
		}

	}

	if !helper.IsValidPassword(newPassword){
		session.Set("password_error","Password must be at least 8 characters with uppercase, lowercase, number, and special character")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/change-password")
		return
	}

	hashedPass, hasherr := utils.HashPassword(newPassword)

	if hasherr != nil {
		session.Set("password_error","Failed to generator hash of new password")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/change-password")
		return
	}

	user.Password = hashedPass

	if err := db.Db.Save(&user).Error; err != nil{
		session.Set("password_error","Failed to update new password")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/change-password")
		return
	}

	session.Set("profilesuccess","Password updated successfully")
	session.Save()
	c.Redirect(http.StatusFound,"/user/profile")

}	



func DeleteAddress(c *gin.Context){

	addressID := c.PostForm("address_id")

	var address models.Address
	var addresses []models.Address

	if err := db.Db.Where("address_id = ?",addressID).First(&address).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Failed to get address"})
		return 
	}

	if err := db.Db.Where("user_id = ?",address.UserID).Find(&addresses).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Failed to get address"})
		return 
	}

	if address.IsDefault {
		for _,ad := range addresses{
			if ad.AddressID != address.AddressID {
				ad.IsDefault = true
				db.Db.Save(&ad)
				break
			}
		}
	}

	if err := db.Db.Delete(&address,addressID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Address not found"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/user/profile")

}