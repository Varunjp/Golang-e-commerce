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
	router.GET("/admin",middleware.AuthMiddlerware("admin"),admin.LoginPage)
	router.GET("/admin/login",admin.LoginPage)
	router.POST("/admin/login",admin.Login)
	// Admin users
	router.GET("/admin/users-list",middleware.AuthMiddlerware("admin"),admin.ListUsers)
	router.GET("/admin/users",middleware.AuthMiddlerware("admin"),admin.FindUser)
	router.GET("/admin/users/block/:id",middleware.AuthMiddlerware("admin"),admin.BlockUser)
	router.GET("/admin/users/unblock/:id",middleware.AuthMiddlerware("admin"),admin.UnblockUser)

	// Admin categories
	router.GET("/admin/categories",middleware.AuthMiddlerware("admin"),admin.ViewCategory)
	

}