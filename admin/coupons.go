package admin

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
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

	session := sessions.Default(c)
	flash := session.Get("flash")
	name := AdminUser.Username

	var coupons []models.Coupons

	if err := db.Db.Find(&models.Coupons{}).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Failed to load coupons details, please try again later"})
			return 
		}
	}


	// loading categories

	var subcat []models.SubCategory
	
	type Response struct {
		SubCategoryID 		int
		SubCategoryName		string 
		CategoryName		string 
	}


	if err := db.Db.Where("is_blocked = ?",false).Find(&subcat).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_addProduct.html",gin.H{"error":"Failed to load subcategory,please try again later"})
		return 
	}

	response := make([]Response,len(subcat))

	for i,subitem := range subcat {

		var Category models.Category

		if err := db.Db.Where("category_id = ?",subitem.CategoryID).First(&Category).Error; err != nil{
			log.Println("Failed to load category details")
			c.Redirect(http.StatusTemporaryRedirect,"/admin")
			return 
		}

		response[i] = Response{
			SubCategoryID: int(subitem.SubCategoryID),
			SubCategoryName: subitem.SubCategoryName,
			CategoryName: Category.CategoryName,
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

		if flash != nil{
			session.Delete("flash")
			session.Save()
			c.HTML(http.StatusOK,"coupons.html",gin.H{
				"user":name,
				"couponsList": coupons,
				"page":page,
				"limit":limit,
				"totalPages":totaPage,
				"Subcategories":response,
				"error":flash,
			})
			return
		}

		c.HTML(http.StatusOK,"coupons.html",gin.H{
			"user":name,
			"couponsList": coupons,
			"page":page,
			"limit":limit,
			"totalPages":totaPage,
			"Subcategories":response,
		})

		return 

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

		if flash != nil{
			session.Delete("flash")
			session.Save()
			c.HTML(http.StatusOK,"coupons.html",gin.H{
				"user":name,
				"couponsList": coupons,
				"page":page,
				"limit":limit,
				"totalPages":totalPages,
				"Subcategories":response,
				"error":flash,
			})
			return 
		}

		c.HTML(http.StatusOK,"coupons.html",gin.H{
			"user":name,
			"couponsList": coupons,
			"page":page,
			"limit":limit,
			"totalPages":totalPages,
			"Subcategories":response,
		})

		return

	}

	//c.HTML(http.StatusOK,"coupons.html",gin.H{"couponsList":coupons})

}

func AddCoupon(c *gin.Context){

	type couponInput struct {
		Code 			string 		`form:"code" binding:"required"`
		Description 	string 		`form:"description" binding:"required"`
		Discount 		float64 	`form:"discount" binding:"required"`
		MinAmount		float64		`form:"min_amount" binding:"required"`
		MaxAmount 		float64 	`form:"max_amount" binding:"required"`
		Type 			string 		`form:"type" binding:"required"`
		Active 			string 		`form:"active" binding:"required"`
	}

	var Existcoupon models.Coupons

	categoryID := c.PostForm("subcategory_id")
	var input couponInput

	if err := c.ShouldBind(&input); err != nil{
		
		c.HTML(http.StatusBadRequest,"coupons.html",gin.H{"error":"All fields are required"})
		return 
	}
	session := sessions.Default(c)
	if strings.TrimSpace(input.Code) == ""{
		session.Set("flash","Please provide proper code")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/admin/coupons")
		return 
	}

	if input.Discount > 75{
		session.Set("flash","Maximum discount is set to 75")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/admin/coupons")
		return 
	}

	if input.MinAmount < 100 {
		session.Set("flash","Provide minimum amount higher than 100")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/admin/coupons")
		return 
	}

	if strings.TrimSpace(input.Description) == "" {
		session.Set("flash","Please provide proper description")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/admin/coupons")
		return 
	}

	checkCode := strings.TrimSpace(input.Code)
	checkCode = strings.ReplaceAll(checkCode," ","")

	if err := db.Db.Where("code ILIKE ?",checkCode).First(&Existcoupon).Error; err == nil{
		session.Set("flash","Coupon code already exist")
		session.Save()
		c.Redirect(http.StatusSeeOther,"/admin/coupons")
		return 
	}

	isActive := input.Active == "true"

	var coupon models.Coupons

	newCode := strings.ToUpper(strings.TrimSpace(input.Code))
	newCode = strings.ReplaceAll(newCode," ","")

	if categoryID != ""{
		catId,_ := strconv.Atoi(categoryID)

		coupon = models.Coupons{
		Code: newCode,
		Description: input.Description,
		Discount: input.Discount,
		MinAmount: input.MinAmount,
		MaxAmount: input.MaxAmount,
		Type: input.Type,
		IsActive: isActive,
		CreatedAt: time.Now(),
		CategoryID: uint(catId),
		}
	}else{

		coupon = models.Coupons{
		Code: newCode,
		Description: input.Description,
		Discount: input.Discount,
		MinAmount: input.MinAmount,
		MaxAmount: input.MaxAmount,
		Type: input.Type,
		IsActive: isActive,
		CreatedAt: time.Now(),
		CategoryID: 0,
		}
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

func DeleteCoupon(c *gin.Context){
	id := c.Param("id")

	if err := db.Db.Delete(&models.Coupons{},id).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Failed to delete coupon"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/admin/coupons")
}

func EditCouponPage(c *gin.Context){

	id := c.Param("id")
	var coupon models.Coupons
	if err := db.Db.Where("id = ?",id).First(&coupon).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"coupons.html",gin.H{"error":"Failed to load coupon details"})
		return 
	}

	var subcat []models.SubCategory
	
	type Response struct {
		SubCategoryID 		int
		SubCategoryName		string 
		CategoryName		string 
	}


	if err := db.Db.Where("is_blocked = ?",false).Find(&subcat).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_addProduct.html",gin.H{"error":"Failed to load subcategory,please try again later"})
		return 
	}

	response := make([]Response,len(subcat))

	for i,subitem := range subcat {

		var Category models.Category

		if err := db.Db.Where("category_id = ?",subitem.CategoryID).First(&Category).Error; err != nil{
			log.Println("Failed to load category details")
			c.Redirect(http.StatusTemporaryRedirect,"/admin")
			return 
		}

		response[i] = Response{
			SubCategoryID: int(subitem.SubCategoryID),
			SubCategoryName: subitem.SubCategoryName,
			CategoryName: Category.CategoryName,
		}

	}

	c.HTML(http.StatusOK,"admin_editCoupon.html",gin.H{
		"Coupon":coupon,
		"Subcategories":response,
	})
}

func EditCoupon(c *gin.Context){
	// var couponExist models.Coupons
	Errors := make(map[string]string)
	idParam := c.Param("id")
	id,err := strconv.Atoi(idParam)
	if err != nil{
		c.HTML(http.StatusInternalServerError,"coupon.html",gin.H{"error":"Invaild id"})
		return 
	}

	var coupon models.Coupons

	if err := db.Db.Where("id = ?",id).First(&coupon).Error; err != nil{
		c.String(http.StatusInternalServerError,"Coupon not found")
		return 
	}

	newCode := c.PostForm("code")
	newDescription := c.PostForm("description")
	discount,_ := strconv.ParseFloat(c.PostForm("discount"),64)
	minAmount,_ := strconv.ParseFloat(c.PostForm("min_amount"),64)
	maxAmount,_ := strconv.ParseFloat(c.PostForm("max_amount"),64)

	// if err := db.Db.Where("code ILIKE ?",newCode).First(&couponExist).Error; err == nil{
	// 	Errors["code"] = "Already exist"
	// }

	if strings.TrimSpace(newCode) == ""{
		Errors["code"] = "Coupon code doesn't meet requirment" 
	}

	if minAmount <= 0 {
		Errors["min_amount"] = "Coupon minimum amount should be greater than zero"
	}

	if discount > 75{
		Errors["discount"] = "Coupon disount should be less than 75%"
	}

	if strings.TrimSpace(newDescription) == ""{
		Errors["description"] = "Coupon description doesn't meet requirment"
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest,gin.H{
			"status":"error",
			"errors":Errors,
		})
		return 
	}


	coupon.Code = strings.ToUpper(newCode)
	coupon.Description = newDescription

	

	coupon.Discount = discount
	coupon.MinAmount = minAmount
	coupon.MaxAmount = maxAmount
	coupon.Type = c.PostForm("type")

	

	categoryIDstr := c.PostForm("category_id")
	if categoryIDstr != ""{
		categoryID,_ := strconv.Atoi(categoryIDstr)
		coupon.CategoryID = uint(categoryID)
	}else{
		coupon.CategoryID = 0
	}

	if err := db.Db.Save(&coupon).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error","message":"Failed to update coupon"})
		return 
	}

	c.JSON(http.StatusOK,gin.H{"redirect":"/admin/coupons"})
}