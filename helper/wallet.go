package helper

import (
	db "first-project/DB"
	"first-project/models"
	"time"

	"gorm.io/gorm"
)

func CreditWallet(userId uint, amount float64, reason string) error {


	if err := db.Db.Model(&models.Wallet{}).Where("user_id = ?",userId).Update("balance",gorm.Expr("balance + ?",amount)).Error; err != nil{
		return err 
	}

	transaction := models.WalletTransaction{
		UserID: userId,
		Amount: amount,
		Type: "credit",
		Description: reason,
	}

	if err := db.Db.Create(&transaction).Error; err != nil{
		return err 
	}

	return nil 
}

func DebitWallet(userId uint, amount float64, orderID uint,reason string) error {

	if err := db.Db.Model(&models.Wallet{}).Where("user_id = ?",userId).Update("balance",gorm.Expr("balance - ?",amount)).Error; err != nil{
		return err 
	}

	transaction := models.WalletTransaction{
		UserID: userId,
		Amount: -amount,
		Type: "Debit",
		Description: reason,
		CreatedAt: time.Now(),
	}

	if err := db.Db.Create(&transaction).Error; err != nil{
		return err 
	}

	return nil 

}