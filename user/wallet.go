package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"net/http"
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

