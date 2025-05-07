package helper

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"math/rand"
	"time"
)

func GenerateAndSaveOtp(email string) (string, error) {

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