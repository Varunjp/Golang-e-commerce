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
	var isuser uint 
	
	var productOffer models.ProductOffer
	var categoryOffer models.CategoryOffer
	today := time.Now().Format("2006-01-02")

	type ReviewResponse struct{
		ID			uint
		UserName 	string
		UserID 		uint
		Rating 		int 
		CreatedAt 	time.Time
		Comment 	string
	}

	if userId > 0 {
		isuser = uint(userId)
	}

	type relatedItem struct{
		ID 				uint 
		ImageURL 		string
		Name 			string 
		AverageRating 	int 
		Price 			float64
	}
	var relatedProduct []relatedItem

	if err := db.Db.Where("deleted_at IS NULL").First(&product_variant,productID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product not found"})
		return 
	}

	if err := db.Db.Preload("Product_variants").Where("deleted_at IS NULL AND product_id = ?",product_variant.ProductID).First(&product).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product not found"})
		return 
	}

	if err := db.Db.Where("created_at <= ? AND product_id = ? AND active = true ",today+" 23:00:00",product.ProductID).First(&productOffer).Error; err != nil{
		log.Println("No offer available ",err)
	}

	if err := db.Db.Where("created_at <= ? AND category_id = ? AND active = true",today+" 23:00:00",product.SubCategoryID).First(&categoryOffer).Error; err != nil{
		log.Println("No offer available as category ",err)
	}

	//related products
	var ProductRelated []models.Product
	if err := db.Db.Preload("Product_variants").Where("sub_category_id = ?",product.SubCategoryID).Limit(4).Find(&ProductRelated).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product not found"})
		return 
	}

	for _,pro := range ProductRelated {
		variCount := 0
		for _, vari := range pro.Product_variants{
			
			if vari.ID != product_variant.ID{
					variCount++
				if variCount > 1{
					break
				}

				var total int64
				var sum float64
				var image models.Product_image
				db.Db.Model(&models.Review{}).
					Where("product_id = ?", pro.ProductID).
					Count(&total).
					Select("AVG(rating)").Row().Scan(&sum)

				averageRating := 0.0
				if total > 0 {
					averageRating = sum
				}

				db.Db.Where("product_variant_id = ?",vari.ID).Order("order_no ASC").First(&image)

				relatedProduct = append(relatedProduct, relatedItem{
					ID: vari.ID,
					Name: pro.ProductName,
					Price: vari.Price,
					ImageURL: image.Image_url,
					AverageRating: int(averageRating),
				})
			}

		}

	}

	var reviews []models.Review
	var respReview []ReviewResponse
    db.Db.Preload("User").Where("product_id = ?", product.ProductID).Order("created_at desc").Find(&reviews)

	for _,rev := range reviews{
		respReview =  append(respReview, ReviewResponse{
			ID: rev.ID,
			UserName: rev.User.Username,
			UserID: rev.UserID,
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
		}else if errmsg != nil{
			session.Delete("error")
			session.Save()
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
			"RelatedProducts":relatedProduct,
			"CurrentUserID":isuser,
			"message":flash,
			"error":errmsg,
			"ProductOffer":productOffer,
			"CategoryOffer":categoryOffer,
		})
	}else{

		if flash != nil{
			session.Delete("flash")
			session.Save()
		}else if errmsg != nil{
			session.Delete("error")
			session.Save()
			
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
			"RelatedProducts":relatedProduct,
			"message":flash,
			"error":errmsg,
			"ProductOffer":productOffer,
			"CategoryOffer":categoryOffer,
		})
	}

}