package admin

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListCoupons(c *gin.Context){
	
	tokenStr,_ := c.Cookie("JWT-Admin")
	_,userId,_ := helper.DecodeJWT(tokenStr)
	var AdminUser models.Admin

	if err := db.Db.Where("id = ?",userId).First(&AdminUser).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Please login again"})
		return 
	}

	name := AdminUser.Username

	var coupons []models.Coupons

	if err := db.Db.Find(&models.Coupons{}).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Failed to load coupons details, please try again later"})
			return 
		}
	}

	pageStr := c.DefaultQuery("page","1")
	limitStr := c.DefaultQuery("limit","10")

	page , err := strconv.Atoi(pageStr)

	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var total int64

	keyword := c.Query("search")

	if keyword == ""{

		adDb := db.Db.Model(&models.Coupons{})
		adDb.Count(&total)

		adDb.Order("id DESC").Limit(limit).Offset(offset).Find(&coupons)

		totaPage := int(math.Ceil(float64(total)/float64(limit)))

		c.HTML(http.StatusOK,"coupons.html",gin.H{
			"user":name,
			"couponsList": coupons,
			"page":page,
			"limit":limit,
			"totalPages":totaPage,
		})

	}else{

		adDb := db.Db.Where("LOWER(code) LIKE ?","%"+strings.ToLower(keyword)+"%").Order("id DESC").Limit(limit).Offset(offset).Find(&coupons)

		if err := adDb.Error; err != nil {
			c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Unable retreive info please try again later"})
			return 
		}

		adDb.Count(&total)

		if total < 1 {
			c.HTML(http.StatusNotFound,"coupons.html",gin.H{"error":"Coupon not found"})
			return 
		}

		totalPages := int(math.Ceil(float64(total)/float64(limit)))

		c.HTML(http.StatusOK,"coupons.html",gin.H{
			"user":name,
			"couponsList": coupons,
			"page":page,
			"limit":limit,
			"totalPages":totalPages,
		})

	}

	c.HTML(http.StatusOK,"coupons.html",gin.H{"couponsList":coupons})

}

func AddCoupon(c *gin.Context){

	type couponInput struct {
		Code 			string 		`form:"code" binding:"required"`
		Description 	string 		`form:"code" binding:"required"`
		Discount 		float64 	`form:"discount" binding:"required"`
		Active 			string 		`form:"active" binding:"required"`
	}


	var input couponInput

	if err := c.ShouldBind(&input); err != nil{
		
		c.HTML(http.StatusBadRequest,"coupons.html",gin.H{"error":"All fields are required"})
		return 
	}

	isActive := input.Active == "true"

	coupon := models.Coupons{
		Code: input.Code,
		Description: input.Description,
		Discount: input.Discount,
		IsActive: isActive,
		CreatedAt: time.Now(),
	}

	if err := db.Db.Create(&coupon).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Failed to add coupon: "+err.Error()})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/admin/coupons")

}

func ToggleCoupon(c *gin.Context){
	
	id := c.Param("id")
	var coupon models.Coupons

	if err := db.Db.Where("id = ?",id).First(&coupon).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Failed to get details about coupons"})
		return 
	}

	if coupon.IsActive {
		coupon.IsActive = false
		if err := db.Db.Save(&coupon).Error; err != nil{
			c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Couldn't update coupon status, please try again later"})
			return 
		}

		c.Redirect(http.StatusSeeOther,"/admin/coupons")
	}else{

		coupon.IsActive = true
		if err := db.Db.Save(&coupon).Error; err != nil{
			c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Couldn't update coupon status, please try again later"})
			return 
		}

		c.Redirect(http.StatusSeeOther,"/admin/coupons")


	}
}