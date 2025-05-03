package routes

import (
	"first-project/admin"
	"first-project/middleware"
	"first-project/user"

	"github.com/gin-gonic/gin"
)

func GetUrl(router *gin.Engine){
	
	// user
	router.POST("/login",user.Login)

	//Admin
	router.POST("/admin/login",admin.Login)
	router.GET("/admin/users-list",middleware.AuthMiddlerware("admin"),admin.ListUsers)

}