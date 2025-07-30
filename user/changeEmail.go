package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ChangeEmail(c *gin.Context) {
	email := c.Query("email")
	name := c.Query("name")
	phone := c.Query("phone")

	if strings.TrimSpace(email) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(phone) == ""{
		c.Redirect(http.StatusSeeOther,c.Request.Referer())
		return 
	}

	c.HTML(http.StatusOK,"changeEmail.html",gin.H{
		"Email":email,
		"name":name,
		"phone":phone,
	})
}