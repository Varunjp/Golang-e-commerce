package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WishlistPage(c *gin.Context) {

	var wishlists []models.WishList

	tokenStr,_ := c.Cookie("JWT-User")
	_,userId,_ := helper.DecodeJWT(tokenStr)

	if err := db.Db.Where("user_id = ?",userId).Find(&wishlists).Error; err != nil{
		
		if err != gorm.ErrRecordNotFound{

			log.Println(err)
			c.HTML(http.StatusInternalServerError,"wishlist.html",gin.H{"error":"Failed to load wishlist, please try again later"})
			return 
		}
	}

	if len(wishlists) == 0{
		c.HTML(http.StatusOK,"wishlist.html",gin.H{"message":"Wishlist is empty!!"})
		return 
	}

	pageStr := c.DefaultQuery("page","1")
	limiStr := c.DefaultQuery("limit","10")

	page, err := strconv.Atoi(pageStr)

	if err != nil || page < 1{
		page = 1
	}

	limit , err := strconv.Atoi(limiStr)

	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1)* limit 

	productIDs := []uint{}

	for _,item := range wishlists{
		productIDs = append(productIDs, item.ProductID)
	}

	var products []models.Product_Variant

	if err := db.Db.Preload("Product_images", func(db *gorm.DB) *gorm.DB {
        return db.Order("order_no ASC")
    	}).Where("id IN ?",productIDs).Offset(offset).Limit(limit).Find(&products).Error; err != nil{
		log.Println(err)
		c.HTML(http.StatusInternalServerError,"wishlist.html",gin.H{"error":"Matching products not found, please try again later"})
		return 
	}

	total := len(products)

	type Response struct {
		ID					uint
		Name				string
		Price				float64
		DiscountedPrice		string 
		ImageURL 			string
	}

	response := make([]Response,len(products))

	for i,p := range products{

		response[i] = Response{
			ID: p.ID,
			Name: p.Variant_name,
			Price: p.Price,
			DiscountedPrice: "",
		}

		if len(p.Product_images) > 0 {
			response[i].ImageURL = p.Product_images[0].Image_url
		}else{
			response[i].ImageURL = ""
		}

	}

	totalPage := int(math.Ceil(float64(total)/float64(limit)))

	session := sessions.Default(c)
	name,_ := session.Get("name").(string)

	c.HTML(http.StatusOK,"wishlist.html",gin.H{
		"user": name,
		"Products": response,
		"pagetitle": "wishlist",
		"Page": page,
		"TotalPages":totalPage,
	})

}

func AddToWishlist(c *gin.Context){

	productIDstr := c.Param("id")
	productID,_ := strconv.Atoi(productIDstr)
	tokenStr,_ := c.Cookie("JWT-User")
	_,userId,_ := helper.DecodeJWT(tokenStr)

	if err := db.Db.Where("user_id = ? AND product_id = ?",userId,productID).First(&models.WishList{}).Error ; err != nil{
		if err != gorm.ErrRecordNotFound {
			c.Redirect(http.StatusSeeOther,"/user/shop")
			return 
		}
	}

	wishlist := models.WishList{
		UserID: uint(userId),
		ProductID: uint(productID),
		CreatedAt: time.Now(),
	}

	if err := db.Db.Create(&wishlist).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"wishlist.html",gin.H{"error":"failed to add to wishlist"})
		return 
	}

	referer := c.Request.Referer()

	if referer != ""{
		c.Redirect(http.StatusSeeOther,referer)
	}else{
		c.Redirect(http.StatusSeeOther,"/user/shop")
	}

	
}

func RemoveWishlist(c *gin.Context){

	productIDstr := c.Param("id")
	productID,_ := strconv.Atoi(productIDstr)
	tokenStr,_ := c.Cookie("JWT-User")
	_,userId,_ := helper.DecodeJWT(tokenStr)
	var wishlist models.WishList

	if err := db.Db.Where("user_id = ? AND product_id = ?",userId,productID).First(&wishlist).Error ; err != nil{
		if err != gorm.ErrRecordNotFound {
			c.Redirect(http.StatusSeeOther,"/user/shop")
			return 
		}
	}

	if err := db.Db.Delete(&wishlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to remove from wishlist"})
		return 
	}

	referer := c.Request.Referer()

	if referer != ""{
		c.Redirect(http.StatusSeeOther,referer)
	}else{
		c.Redirect(http.StatusSeeOther,"/user/shop")
	}

}