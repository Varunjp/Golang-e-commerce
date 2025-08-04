package user

import (
	"first-project/helper"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func HomePage(c *gin.Context){

	session := sessions.Default(c)
	username,usererr := session.Get("name").(string)

	products, imageUrl,redirctUrl, err := helper.GetHomePage()

	offers := helper.GetSpecialOffer()

	if usererr{
			if err != nil{
				c.HTML(http.StatusBadRequest,"home.html",gin.H{
					"user": username,
					"Image_url": imageUrl,
					"RedirectURL": redirctUrl,
					"Products": products,
					"error" : err.Error(),
					"Announcements":offers,
				})
				return 
			}

			c.HTML(http.StatusOK,"home.html",gin.H{
				"user": username,
				"Image_url": imageUrl,
				"RedirectURL": redirctUrl,
				"Products": products,
				"Announcements":offers,
			})
	}else{

		if err != nil{
			c.HTML(http.StatusBadRequest,"home.html",gin.H{
				"Image_url": imageUrl,
				"Products": products,
				"RedirectURL": redirctUrl,
				"error" : err.Error(),
				"Announcements":offers,
				})
			return 
		}

		c.HTML(http.StatusOK,"home.html",gin.H{
			"Image_url": imageUrl,
			"Products": products,
			"RedirectURL": redirctUrl,
			"Announcements":offers,
		})
	}

	
}

func DemoPage(c *gin.Context){

	products, imageUrl,redirctUrl, err := helper.GetHomePage()

	if err != nil{
		c.HTML(http.StatusBadRequest,"home.html",gin.H{
			"Image_url": imageUrl,
			"Products": products,
			"RedirectURL": redirctUrl,
			"error" : err.Error(),
		})
	}

	c.HTML(http.StatusOK,"home.html",gin.H{
		"Image_url": imageUrl,
		"Products": products,
		"RedirectURL": redirctUrl,
	})

}