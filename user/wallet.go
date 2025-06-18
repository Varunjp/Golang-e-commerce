package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WalletDetails(c *gin.Context){

	var Wallet models.Wallet

	tokenStr,_ := c.Cookie("JWT-User")
	_,UserId,_ := helper.DecodeJWT(tokenStr)

	if err := db.Db.Where("user_id = ?",UserId).First(&Wallet).Error; err != nil{

		if err != gorm.ErrRecordNotFound{
			log.Println(err)
			c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to load wallet details, please try again later"})
			return 
		}

		if err == gorm.ErrRecordNotFound{
			Wallet.UserID = uint(UserId)
			Wallet.UpdatedAt = time.Now()
			Wallet.Balance = 0.0

			if errCr := db.Db.Create(&Wallet).Error; errCr != nil{
				c.HTML(http.StatusInternalServerError,"wallet.html",gin.H{"error":"Failed to retrieve wallet data"})
				return 
			}
		}

	}

	c.HTML(http.StatusOK,"wallet.html",gin.H{"wallet":Wallet,"user":"done"})
}

func WalletTransaction(c *gin.Context){
	tokenStr,_ := c.Cookie("JWT-User")
	_,UserId,_ := helper.DecodeJWT(tokenStr)
	var walletTransaction []models.WalletTransaction

	page,_ := strconv.Atoi(c.DefaultQuery("page","1"))
	limit := 10
	offset := (page - 1) * limit

	var total int64
	db.Db.Unscoped().Model(&models.WalletTransaction{}).Where("user_id = ?",UserId).Count(&total)

	if err := db.Db.Unscoped().Where("user_id = ?",UserId).Order("id desc").Offset(offset).Limit(limit).Find(&walletTransaction).Error; err != nil{
		if err != gorm.ErrRecordNotFound {
			c.HTML(http.StatusInternalServerError,"walletTransaction.html",gin.H{"error":"Failed to load wallet transactions, please try again later"})
			return 
		}else{
			c.HTML(http.StatusNotFound,"walletTransaction.html",gin.H{"user":"done"})
		}
	}

	type response struct {
		ID  		uint
		Type 		string 
		Amount 		float64
		Description string
		Date 		string
		Status 		string
		BadgeClass  string  
	}

	transactions := make([]response,len(walletTransaction))
	
	for i,trans := range walletTransaction {

		transactions[i] = response{
			ID: trans.ID,
			Type: trans.Type,
			Amount: trans.Amount,
			Description: trans.Description,
			Date: trans.CreatedAt.Format("02-01-2006"),
		}

		if !trans.RefundStatus && trans.Type == "Debit"{
			transactions[i].Status = "Success"
			transactions[i].BadgeClass = "success"
		}else if !trans.RefundStatus && trans.Type == "Credit"{
			transactions[i].Status = "Success"
			transactions[i].BadgeClass = "success"
		}else {
			transactions[i].Status = "Failed"
			transactions[i].BadgeClass = "danger"
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	var pages []map[string]int
	for i := 1; i <= totalPages; i++{
		pages = append(pages, map[string]int{"Number":i})
	}

	c.HTML(http.StatusOK,"walletTransaction.html",gin.H{
		"user":"done",
		"Transactions":transactions,
		"CurrentPage": page,
		"Pages": pages,
		"HasPrev": page > 1,
		"HasNext": page < totalPages,
		"PrevPage": page - 1,
		"NextPage": page + 1,
	})

}