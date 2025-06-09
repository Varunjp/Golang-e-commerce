package admin

import (
	db "first-project/DB"
	"first-project/models"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WalletTransactions(c *gin.Context){
	
	var WalletTransactions []models.WalletTransaction
	type response struct{
		ID 			uint
		UserName 	string
		Type 		string
		Amount 		float64
		Description string 
		CreatedAt 	time.Time 
	}


	if err := db.Db.Order("id DESC").Find(&WalletTransactions).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Could not transaction details, please try again later"})
		return 
	}

	ResponseTransactions := make([]response,len(WalletTransactions))

	for i, transaction := range WalletTransactions {
		var user models.User
		db.Db.Where("id = ?",transaction.UserID).First(&user)
		ResponseTransactions[i] = response{
			ID: transaction.ID,
			UserName: user.Username,
			Type: transaction.Type,
			Amount: math.Abs(transaction.Amount),
			Description: transaction.Description,
			CreatedAt: transaction.CreatedAt,
		}
	}

	c.HTML(http.StatusOK,"wallet.html",gin.H{"transactions":ResponseTransactions})

}

func WalletRefunds(c *gin.Context){

	var walletTransactions []models.WalletTransaction
	
	type response struct{
		ID 			uint
		UserName 	string
		OrderID 	uint
		Amount 		float64
		Description string 
		CreatedAt 	time.Time 
	}

	

	if err := db.Db.Where("refund_status = ?",true).Find(&walletTransactions).Order("id DESC").Error; err != nil{
		if err != gorm.ErrRecordNotFound {
			c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to fetch wallet transcations, please try again later"})
			return 
		}
	}


	if len(walletTransactions) > 0 {

		ResponseTransactions := make([]response,len(walletTransactions))

		for i, transaction := range walletTransactions {
			var user models.User
			db.Db.Where("id = ?",transaction.UserID).First(&user)
			ResponseTransactions[i] = response{
				ID: transaction.ID,
				UserName: user.Username,
				OrderID: transaction.OrderID,
				Amount: math.Abs(transaction.Amount),
				Description: transaction.Description,
				CreatedAt: transaction.CreatedAt,
			}
		}

		c.HTML(http.StatusOK,"refund_request.html",gin.H{"refundRequests":ResponseTransactions})

	}else{
		c.HTML(http.StatusNotFound,"refund_request.html",gin.H{"message":"No refund transactions"})
	}

}

func WalletRefundApproval (c *gin.Context){
	
	transactionId := c.PostForm("request_id")
	reason := c.PostForm("note")
	var transaction models.WalletTransaction

	if err := db.Db.Where("id = ?",transactionId).First(&transaction).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to get details for specific transcation, please try again later"})
		return 
	}

	if err := db.Db.Model(&models.Wallet{}).Where("user_id = ?",transaction.UserID).Update("balance",gorm.Expr("balance + ?",math.Abs(transaction.Amount))).Error; err != nil {
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to update user wallet, please try again later"})
		return
	}

	transaction.RefundStatus = false
	transaction.Description = reason
	transaction.Type = "Credit"

	if err := db.Db.Save(&transaction).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to save transaction, please try again later"})
		return
	}

	c.Redirect(http.StatusSeeOther,"/admin/wallet-transactions")
}


func WalletRefundDecline (c *gin.Context){
	
	transactionId := c.PostForm("request_id")
	reason := c.PostForm("note")
	var transaction models.WalletTransaction

	if err := db.Db.Where("id = ?",transactionId).First(&transaction).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to get details for specific transcation, please try again later"})
		return 
	}

	transaction.RefundStatus = false
	transaction.Description = reason
	transaction.Type = "Refund declined"

	if err := db.Db.Save(&transaction).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to save transaction, please try again later"})
		return
	}

	c.Redirect(http.StatusSeeOther,"/admin/wallet-transactions")
}