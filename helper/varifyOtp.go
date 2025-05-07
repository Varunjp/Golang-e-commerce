package helper

import (
	db "first-project/DB"
	"first-project/models"
	"time"
)

func VerfiyOTP(email, enteredOTP string) (bool, error) {

	var otprecord models.OTPVerification

	err := db.Db.Where("email = ? AND otp = ? AND is_used = false AND expires_at > ?",email,enteredOTP,time.Now()).First(&otprecord).Error

	if err != nil {
		return false, err
	}

	otprecord.IsUsed = true
	db.Db.Save(otprecord)

	return true,nil 
}