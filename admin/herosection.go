package admin

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func BannerPage(c *gin.Context){

	session := sessions.Default(c)
	name,_ := session.Get("admin-name").(string)
	var Banners []models.Banner

	if err := db.Db.Order("created_at DESC").Find(&Banners).Error; err != nil && err != gorm.ErrRecordNotFound{
		c.HTML(http.StatusInternalServerError,"banner.html",gin.H{"error":"Could not retrive banners, please try again later"})
		return 
	}

	c.HTML(http.StatusOK,"banner.html",gin.H{
		"user":name,
		"banners":Banners,
	})

}

func AddBanner(c *gin.Context){

	title := c.PostForm("title")
	redirectUrl := c.PostForm("redirect")
	active := c.PostForm("active") == "on"

	// getting file 

	file, err := c.FormFile("image")

	if err != nil {
		c.HTML(http.StatusBadRequest,"banner.html",gin.H{"error":"Image is required"})
		return 
	}

	uploadpath := "./upload"

	if err := os.MkdirAll(uploadpath,os.ModePerm); err != nil{
		c.HTML(http.StatusInternalServerError,"banner.html",gin.H{"error":"Unable to access path"})
		return 
	}

	filename := fmt.Sprintf("%d_%s",time.Now().Unix(),filepath.Base(file.Filename))

	path := "upload/"+filename

	if err := c.SaveUploadedFile(file,path); err != nil{
		c.HTML(http.StatusInternalServerError,"banner.html",gin.H{"error":"Failed to save"})
		return 
	}

	banner := models.Banner{
		Title: title,
		ImageUrl: "/"+path,
		RedirectURL: redirectUrl,
		Active: active,
	}

	if err := db.Db.Create(&banner).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"banner.html",gin.H{"error":"Failed to save image"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/admin/banners")

}

func DeleteBanner(c *gin.Context){
	
	id,_ := strconv.Atoi(c.Param("id"))

	if err := db.Db.Delete(&models.Banner{},id).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"banner.html",gin.H{"error":"Failed to delete, please try again later"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/admin/banners")
}