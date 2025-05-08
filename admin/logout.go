package admin

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Logout(c *gin.Context) {

	session := sessions.Default(c)
	session.Delete("name")
	session.Save()

	c.SetCookie("JWT-Admin","",-1,"/","",false,true)

	c.Redirect(http.StatusTemporaryRedirect,"/admin/login")
}