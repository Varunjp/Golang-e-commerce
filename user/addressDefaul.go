package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func MakeAddressDefault(c *gin.Context) {

	tokenstr,_ := c.Cookie("JWT-User")
	_,userid,_ := helper.DecodEJWT(tokenstr)
	addressIdstr := c.PostForm("address_id")
	session := sessions.Default(c)
	var addresses []models.Address 

	if err := db.Db.Where("user_id = ?",userid).Find(&addresses).Error; err != nil{
		session.Set("flash","Could not get user address details")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/user/profile")
		return
	}

	addressId,_ := strconv.Atoi(addressIdstr)

	for _,address := range addresses{

		if address.AddressID == uint(addressId) {
			address.IsDefault = true
		}else{
			address.IsDefault = false 
		}

		if err := db.Db.Save(&address).Error; err != nil{
			session.Set("flash","Could not update address, please try again later")
			session.Save()
			c.Redirect(http.StatusSeeOther,"/user/profile")
			return
		}
	}

	c.Redirect(http.StatusSeeOther,"/user/profile")
}