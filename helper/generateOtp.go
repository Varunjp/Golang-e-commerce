package helper

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"log"
	"math/rand"
	"time"
)

func GenerateAndSaveOtp(email string) (string, error) {

	err := db.Db.Where("expires_at < ?",time.Now()).Delete(&models.OTPVerification{}).Error

	if err != nil {
		log.Fatal("Error on clearing old otp :",err.Error())
	}

	if err := db.Db.Model(&models.OTPVerification{}).Where("email = ?",email).Update("is_used",true).Error; err != nil{
		log.Fatal("Error while removing old otp :",err)
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiresAt := time.Now().Add(5 * time.Minute)

	record := models.OTPVerification{
		Email: email,
		OTP: otp,
		ExpiresAt: expiresAt,
	}

	if err := db.Db.Create(&record).Error; err != nil{
		return "",err 
	}

	return otp,nil
}