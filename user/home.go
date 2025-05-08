package user

import (
	"first-project/helper"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func HomePage(c *gin.Context){

	session := sessions.Default(c)
	username := session.Get("name").(string)

	products, imageUrl, err := helper.GetHomePage()

	if err != nil{
		c.HTML(http.StatusBadRequest,"home.html",gin.H{
			"user": username,
			"Image_url": imageUrl,
			"Products": products,
			"error" : err.Error(),
		})
	}

	c.HTML(http.StatusOK,"home.html",gin.H{
		"user": username,
		"Image_url": imageUrl,
		"Products": products,
	})
	
}

func DemoPage(c *gin.Context){

	products, imageUrl, err := helper.GetHomePage()

	if err != nil{
		c.HTML(http.StatusBadRequest,"home.html",gin.H{
			"Image_url": imageUrl,
			"Products": products,
			"error" : err.Error(),
		})
	}

	c.HTML(http.StatusOK,"home.html",gin.H{
		"Image_url": imageUrl,
		"Products": products,
	})

}