package helper

import (
	"fmt"
	"os"

	"github.com/go-gomail/gomail"
)

func SendOTPEmail(username,email, otp string) error {

	myMail := os.Getenv("Email")
	Password := os.Getenv("Password")

	msg := gomail.NewMessage()
	msg.SetHeader("From", myMail)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Fashion Art OTP Code")

	body := fmt.Sprintf(`
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: auto; padding: 20px; border: 1px solid #eee;">
            <h2 style="color: #333;">Hello %s,</h2>
            <p style="font-size: 16px; color: #555;">
                Thank you for using our service. Please use the following One-Time Password (OTP) to complete your action:
            </p>
            <div style="text-align: center; margin: 20px 0;">
                <span style="display: inline-block; background-color: #f4f4f4; padding: 12px 24px; font-size: 24px; font-weight: bold; letter-spacing: 2px; border-radius: 6px; color: #222;">%s</span>
            </div>
            <p style="color: #777;">This OTP is valid for the next 1 minute. Please do not share it with anyone.</p>
            <hr style="border: none; border-top: 1px solid #eee;">
            <p style="font-size: 14px; color: #aaa;">If you did not request this, please ignore this email or contact support.</p>
            <p style="font-size: 14px; color: #aaa;">— Fashion Art Security</p>
        </div>
    `, username, otp)

	msg.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, myMail, Password)
	return d.DialAndSend(msg)
}