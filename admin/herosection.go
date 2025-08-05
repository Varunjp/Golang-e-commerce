package admin

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func BannerPage(c *gin.Context){

	session := sessions.Default(c)
	name,_ := session.Get("admin-name").(string)
	var Banners []models.Banner

	berr := session.Get("banner-error")

	if berr != nil {
		session.Delete("banner-error")
		session.Save()
	}

	if err := db.Db.Order("created_at DESC").Find(&Banners).Error; err != nil && err != gorm.ErrRecordNotFound{
		c.HTML(http.StatusInternalServerError,"banner.html",gin.H{"error":"Could not retrive banners, please try again later"})
		return 
	}

	c.HTML(http.StatusOK,"banner.html",gin.H{
		"user":name,
		"banners":Banners,
		"error": berr,
	})

}

func AddBanner(c *gin.Context){

	title := c.PostForm("title")
	redirectUrl := c.PostForm("redirect")
	active := c.PostForm("active") == "on"
	session := sessions.Default(c)

	if strings.TrimSpace(title) == "" || strings.TrimSpace(redirectUrl) == "" {
		session.Set("banner-error", "Title and Redirect URL are required")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/banners")
		return
	}

	// getting file 

	file, err := c.FormFile("image")

	if err != nil {
		session.Set("banner-error", "Image is required")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/banners")
		return
	}

	// Check MIME type to ensure it's an image
	openedFile, err := file.Open()
	if err != nil {
		session.Set("banner-error", "Failed to open uploaded file")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/banners")
		return
	}
	defer openedFile.Close()
	buffer := make([]byte, 512)
	_, err = openedFile.Read(buffer)
	if err != nil {
		session.Set("banner-error", "Failed to read uploaded file")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/banners")
		return
	}
	contentType := http.DetectContentType(buffer)
	if !(strings.HasPrefix(contentType, "image/")) {
		session.Set("banner-error", "Only image files are allowed")
		session.Save()	
		c.Redirect(http.StatusSeeOther, "/admin/banners")
		return
	}

	uploadpath := "./upload"

	if err := os.MkdirAll(uploadpath,os.ModePerm); err != nil{
		session.Set("banner-error", "Unable to access upload path")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/banners")
		return 
	}

	filename := fmt.Sprintf("%d_%s",time.Now().Unix(),filepath.Base(file.Filename))

	path := "upload/"+filename

	if err := c.SaveUploadedFile(file,path); err != nil{
		session.Set("banner-error", "Failed to save image")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/banners")
		return 
	}

	banner := models.Banner{
		Title: title,
		ImageUrl: "/"+path,
		RedirectURL: redirectUrl,
		Active: active,
	}

	if err := db.Db.Create(&banner).Error; err != nil{
		session.Set("banner-error", "Failed to save banner, please try again later")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/banners")
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

func EditBanner(c *gin.Context){
	id, _ := strconv.Atoi(c.Param("id"))
	var banner models.Banner
	session := sessions.Default(c)

	if err := db.Db.First(&banner, id).Error; err != nil {
		session.Set("banner-error", "Banner not found")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/banners")
		return
	}

	if banner.Active {
		banner.Active = false
	} else {
		banner.Active = true
	}

	db.Db.Save(&banner)

	c.Redirect(http.StatusSeeOther, "/admin/banners")
}