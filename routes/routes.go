package routes

import (
	"first-project/user"

	"github.com/gin-gonic/gin"
)

func GetUrl(router *gin.Engine){
	
	// user
	router.POST("/login",user.Login)

	//Admin
	router.POST("/admin/login")

}