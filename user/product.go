package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Product(c *gin.Context){

	productID := c.Param("id")
	
	tokenStr,_ := c.Cookie("JWT-User")
	_,userId,_ := helper.DecodeJWT(tokenStr)

	session := sessions.Default(c)
	flash := session.Get("flash")
	errmsg := session.Get("error")
	var product models.Product
	var product_variant models.Product_Variant
	var images []models.Product_image
	type ReviewResponse struct{
		UserName 	string
		Rating 		int 
		CreatedAt 	time.Time
		Comment 	string
	}

	if err := db.Db.Where("deleted_at IS NULL").First(&product_variant,productID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product not found"})
		return 
	}

	if err := db.Db.Preload("Product_variants").Where("deleted_at IS NULL AND product_id = ?",product_variant.ProductID).First(&product).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product not found"})
		return 
	}

	var reviews []models.Review
	var respReview []ReviewResponse
    db.Db.Preload("User").Where("product_id = ?", product.ProductID).Order("created_at desc").Find(&reviews)

	for _,rev := range reviews{
		respReview =  append(respReview, ReviewResponse{
			UserName: rev.User.Username,
			Rating: rev.Rating,
			CreatedAt: rev.CreatedAt,
			Comment: rev.Comment,
		})
	}

    // Calculate average rating
    var total int64
    var sum float64
    db.Db.Model(&models.Review{}).
        Where("product_id = ?", product.ProductID).
        Count(&total).
        Select("AVG(rating)").Row().Scan(&sum)

    averageRating := 0.0
    if total > 0 {
        averageRating = sum
    }

	type responseVariant struct {
		ID		uint 
		Size 	string 
	}

	availableVariants := make([]responseVariant,len(product.Product_variants))

	for i, vari := range product.Product_variants {
		availableVariants[i] = responseVariant{
			ID: vari.ID,
			Size: vari.Size,
		}
	}

	if err := db.Db.Where("product_variant_id = ?",productID).Order("order_no ASC").Find(&images).Error; err != nil{
		log.Println("No images found :",err.Error())
	}

	var wishlist models.WishList
	isWishlist := false

	if userId != 0 {
		if err := db.Db.Where("user_id = ? AND product_id = ?",userId,product_variant.ID).First(&wishlist).Error; err == nil{
			isWishlist = true
		}
		if flash != nil{
			session.Delete("flash")
			session.Save()
			c.HTML(http.StatusOK,"product.html",gin.H{
				"user":"done",
				"pagetitle":product_variant.Variant_name,
				"Product": product,
				"variant": product_variant,
				"AllVariants":availableVariants,
				"Images": images,
				"Wishlist":isWishlist,
				"Reviews": respReview,
				"AverageRating": averageRating,
				"TotalReviews":  total,
				"message":flash,
			})
			return 
		}else if errmsg != nil{
			session.Delete("error")
			session.Save()
			c.HTML(http.StatusOK,"product.html",gin.H{
				"user":"done",
				"pagetitle":product_variant.Variant_name,
				"Product": product,
				"variant": product_variant,
				"AllVariants":availableVariants,
				"Images": images,
				"Reviews": respReview,
				"AverageRating": averageRating,
				"TotalReviews":  total,
				"Wishlist":isWishlist,
				"error":errmsg,
			})
			return 
		}
		c.HTML(http.StatusOK,"product.html",gin.H{
			"user":"done",
			"pagetitle":product_variant.Variant_name,
			"Product": product,
			"variant": product_variant,
			"AllVariants":availableVariants,
			"Reviews": respReview,
			"AverageRating": averageRating,
			"TotalReviews":  total,
			"Images": images,
			"Wishlist":isWishlist,
		})
	}else{

		if flash != nil{
			session.Delete("flash")
			session.Save()
			c.HTML(http.StatusOK,"product.html",gin.H{
				"pagetitle":product_variant.Variant_name,
				"Product": product,
				"variant": product_variant,
				"AllVariants":availableVariants,
				"Images": images,
				"Reviews": respReview,
				"AverageRating": averageRating,
				"TotalReviews":  total,
				"Wishlist":false,
				"message":flash,
			})
			return 
		}else if errmsg != nil{
			session.Delete("error")
			session.Save()
			c.HTML(http.StatusOK,"product.html",gin.H{
				"pagetitle":product_variant.Variant_name,
				"Product": product,
				"variant": product_variant,
				"AllVariants":availableVariants,
				"Images": images,
				"Reviews": respReview,
				"AverageRating": averageRating,
				"TotalReviews":  total,
				"Wishlist":false,
				"error":errmsg,
			})
			return
		}
		c.HTML(http.StatusOK,"product.html",gin.H{
			"pagetitle":product_variant.Variant_name,
			"Product": product,
			"variant": product_variant,
			"AllVariants":availableVariants,
			"Images": images,
			"Reviews": respReview,
			"AverageRating":  averageRating,
			"TotalReviews":  total,
			"Wishlist":false,
		})
	}

}