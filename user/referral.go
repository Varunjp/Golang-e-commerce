package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SubmitReferralPage(c *gin.Context){
	c.HTML(http.StatusOK,"referral.html",nil)
}

func SubmitReferralCode(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")
	_,userId,_ := helper.DecodeJWT(tokenStr)
	var newUser models.User

	if err := db.Db.Where("id = ?",userId).First(&newUser).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"referral.html",gin.H{"error":err})
		return 
	}

	// get referral code
	referralCode := c.PostForm("referral_code")

	if referralCode != ""{
		var referrer models.User
		if err := db.Db.Where("referral_code = ?",referralCode).First(&referrer).Error; err == nil{
			newUser.ReferredBy = referralCode
		}

		var welcome,referral models.Coupons
		db.Db.Where("code = ?","REFERRAL100").First(&referral)
		db.Db.Where("code = ?","WELCOME50").First(&welcome)

		db.Db.Create(&models.Coupons{UserID: referrer.ID,
			Code: referral.Code,
			Description: referral.Description,
			Discount: referral.Discount,
			MinAmount: referral.MinAmount,
			MaxAmount: referral.MaxAmount,
			IsActive: true,
			Type: "Referral", 
			CouponID: referral.ID})
		db.Db.Create(&models.Coupons{UserID: newUser.ID, 
			Code: welcome.Code,
			Description: welcome.Description,
			Discount: welcome.Discount,
			MinAmount: welcome.MinAmount,
			IsActive: true,
			MaxAmount: welcome.MaxAmount,
			Type: "Referral",
			CouponID: welcome.ID})
	}

	c.Redirect(http.StatusFound,"/")

}

func GenerateReferralCode(c *gin.Context){
	tokenStr,_ := c.Cookie("JWT-User")
	_,userId,_ := helper.DecodeJWT(tokenStr)
	var user models.User
	if err := db.Db.Where("id = ?",userId).First(&user).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"user_profile.html",gin.H{"error":err})
		return 
	}

	if user.ReferralCode == ""{
		code := helper.GenerateUniqueReferralCode()
		user.ReferralCode = code
	}
	
	db.Db.Save(&user)
	c.Redirect(http.StatusSeeOther,"/user/profile")
}