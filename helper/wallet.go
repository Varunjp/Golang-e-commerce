package helper

import (
	db "first-project/DB"
	"first-project/models"
	"log"
	"time"

	"gorm.io/gorm"
)

func CreateWallet(userID uint) error {
	
	wallet := models.Wallet{
		UserID: userID,
		Balance: 0.0,
	}

	if err := db.Db.Create(&wallet).Error; err != nil{
		log.Println(err)
		return err 
	}

	return nil
}

func CreditWallet(userId uint, amount float64, reason string) error {


	if err := db.Db.Model(&models.Wallet{}).Where("user_id = ?",userId).Update("balance",gorm.Expr("balance + ?",amount)).Error; err != nil{
		return err 
	}

	transaction := models.WalletTransaction{
		UserID: userId,
		Amount: amount,
		Type: "Credit",
		Description: reason,
		Status: true,
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
		OrderID: orderID,
		Description: reason,
		CreatedAt: time.Now(),
		Status: true,
	}

	if err := db.Db.Create(&transaction).Error; err != nil{
		return err 
	}

	return nil 

}

func UpdateDebitWallet(userID uint,orderId uint)error {
	var latestTranscation models.WalletTransaction

	if err := db.Db.Where("user_id = ? AND order_id = ?",userID,userID).Order("created_at DESC").First(&latestTranscation).Error; err != nil{
		return err 
	}

	latestTranscation.OrderID = orderId

	if err := db.Db.Save(&latestTranscation).Error; err != nil{
		return err 
	}

	return nil 
}

func CreditCancelWallet(userId uint, amount float64, reason string) error {


	if err := db.Db.Model(&models.Wallet{}).Where("user_id = ?",userId).Update("balance",gorm.Expr("balance + ?",amount)).Error; err != nil{
		return err 
	}

	return nil 
}