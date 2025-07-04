package middleware

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OrderUpdate() gin.HandlerFunc{
	return func(c *gin.Context){
		var Orders []models.Order
		tokenstr,_ := c.Cookie("JWT-User")
		_,userId,_ := helper.DecodeJWT(tokenstr)

		if err := db.Db.Preload("OrderItems").Where("user_id = ? ",userId).Find(&Orders).Error; err != nil{
			if err != gorm.ErrRecordNotFound{
				c.Redirect(http.StatusSeeOther,"/user/orders")
				c.Abort()
				return 
			}
		}

		for _,order := range Orders {
			items := len(order.OrderItems)
			cancelItem := 0
			if order.Status == "Processing" {
				for _, item := range order.OrderItems{
					if strings.TrimSpace(item.Status) == "Cancelled"{
						cancelItem++
					}
				}
				if items == cancelItem{
					order.Status = "Cancelled"
					if order.PaymentMethod != "cod"{
						order.PaymentStatus = "Refunded"
					}else{
						order.PaymentStatus = "Not valid"
					}
					db.Db.Save(&order)
				}
			}
		}
		c.Next()
	}
}