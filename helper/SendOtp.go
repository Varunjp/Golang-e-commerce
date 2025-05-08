package helper

import (
	"os"

	"github.com/go-gomail/gomail"
)

func SendOTPEmail(email, otp string) error {

	myMail := os.Getenv("Email")
	Password := os.Getenv("Password")

	msg := gomail.NewMessage()
	msg.SetHeader("From", myMail)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Your OTP Code")
	msg.SetBody("text/plain", "Your OTP code is: "+otp)

	d := gomail.NewDialer("smtp.gmail.com", 587, myMail, Password)
	return d.DialAndSend(msg)
}