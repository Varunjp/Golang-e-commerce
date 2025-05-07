package user

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func UserLogout(c *gin.Context){
	
	session := sessions.Default(c)
	session.Delete("name")
	session.Save()

	c.SetCookie("JWT","",-1,"/","",false,true)

	c.Redirect(http.StatusFound,"/")
}