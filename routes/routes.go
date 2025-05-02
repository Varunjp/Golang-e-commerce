package routes

import (
	"github.com/gin-gonic/gin"
)

func GetUrl(router *gin.Engine){
	
	// user
	router.POST("/login")

	//Admin
	router.POST("/admin/login")

}