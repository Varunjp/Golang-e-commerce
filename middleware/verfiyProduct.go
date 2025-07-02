package middleware

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func VerifyProduct() gin.HandlerFunc{
	return func(c *gin.Context) {
		tokenStr,err := c.Cookie("JWT-User")

		if err != nil{
			c.HTML(http.StatusUnauthorized,"userLogin.html",gin.H{"error":"Login required"})
			c.Abort()
			return 
		}
		_,userId,_ := helper.DecodeJWT(tokenStr)
		
		var CartItems []models.CartItem
		referer := c.Request.Referer()

		if err := db.Db.Where("user_id = ?",userId).Find(&CartItems).Error; err != nil {
			if err != gorm.ErrRecordNotFound{
				if referer != ""{
					c.Redirect(http.StatusSeeOther,referer)
				}else{
					c.HTML(http.StatusNotFound,"cart.html",gin.H{"error":"Not able to load cart items"})
				}
				c.Abort()
				return 
			}
			
		}

		for _,item := range CartItems{
			var product models.Product_Variant
			if err := db.Db.Where("id = ?",item.ProductID).First(&product).Error; err != nil{
				if referer != ""{
					db.Db.Delete(&item)
					c.Redirect(http.StatusSeeOther,referer)
				}else{
					c.HTML(http.StatusNotFound,"cart.html",gin.H{"error":"Product has been removed."})
				}
				c.Abort()
				return 
			}

			if product.Stock < item.Quantity{
				if referer != ""{
					db.Db.Delete(&item)
					c.Redirect(http.StatusSeeOther,referer)
				}else{
					c.HTML(http.StatusNotFound,"cart.html",gin.H{"error":"Product has been removed."})
				}
				c.Abort()
				return 
			}
		}

		c.Next()
	}
}