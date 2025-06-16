package helper

import (
	db "first-project/DB"
	"first-project/models"
	"math/rand"
	"time"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomCode(length int) string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	code := make([]byte,length)
	for i := range code{
		code[i] = charset[r.Intn(len(charset))]
	}
	return string(code)
}

func GenerateUniqueReferralCode() string{
	for {

		code := GenerateRandomCode(8)
		var count int64
		db.Db.Model(&models.User{}).Where("referral_code = ?",code).Count(&count)
		if count == 0 {
			return code 
		}

	}
}

func CreateCoupon(userId uint) error{

	var user models.User

	if err := db.Db.Where("id = ?",userId).First(&user).Error; err != nil{
		return err
	}

	if user.ReferredBy != ""{
		var referrer models.User

		if err := db.Db.Where("referral_code = ?",user.ReferredBy).First(&referrer).Error; err != nil{
			return err 
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
		db.Db.Create(&models.Coupons{UserID: user.ID, 
			Code: welcome.Code,
			Description: welcome.Description,
			Discount: welcome.Discount,
			IsActive: true,
			MinAmount: welcome.MinAmount,
			MaxAmount: welcome.MaxAmount,
			Type: "Referral",
			CouponID: welcome.ID})

	}

	return nil 
}