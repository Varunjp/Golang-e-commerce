package admin

import (
	db "first-project/DB"
	"first-project/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type OfferView struct {
	ID                 uint
	OfferName          string
	OfferType          string // "Product" or "Category"
	TargetName         string // Product.Name or Category.Name
	DiscountPercentage float64
	CreatedAt          time.Time
	EndAt              time.Time
	Active             bool
}

func OffersPage(c *gin.Context) {
	session := sessions.Default(c)
	name, _ := session.Get("admin-name").(string)

	pageStr := c.DefaultQuery("page","1")
	limitStr := c.DefaultQuery("limit","10")

	page, err := strconv.Atoi(pageStr)

	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1{
		limit = 10
	}

	offset := (page - 1) * limit

	var total int64

	keyword := c.Query("search")

	var productOffers []models.ProductOffer
	var categoryOffers []models.CategoryOffer

	prodoff := db.Db.Preload("Product").Model(&models.ProductOffer{})
	cateoff := db.Db.Model(&models.CategoryOffer{})

	if keyword != "" {
		var (
			productCount int64
			categoryCount int64
		)	
		prodoff = prodoff.Where("offer_name LIKE ?", "%"+keyword+"%").Count(&productCount)
		cateoff = cateoff.Where("offer_name LIKE ?", "%"+keyword+"%").Count(&categoryCount)
		total = productCount + categoryCount
	}else{
		var (
			productCount int64
			categoryCount int64
		)
		prodoff.Count(&productCount)
		cateoff.Count(&categoryCount)
		total = productCount + categoryCount
	}

	prodoff = prodoff.Offset(offset).Limit(limit).Find(&productOffers)
	cateoff = cateoff.Offset(offset).Limit(limit).Find(&categoryOffers)

	if prodoff.Error != nil || cateoff.Error != nil {
		c.HTML(http.StatusInternalServerError, "admin_offers.html", gin.H{"error": "Could not retrieve offers, please try again later"})
		return
	}
	
	var offers []OfferView
	for _, offer := range productOffers {
		offers = append(offers, OfferView{
			ID:                 offer.ID,
			OfferName:          offer.OfferName,
			OfferType:          "Product",
			TargetName:         offer.Product.ProductName,
			DiscountPercentage: offer.DiscountPercentage,
			CreatedAt:          offer.CreatedAt,
			EndAt:              offer.EndAt,
			Active:             offer.Active,
		})
	}

	for _, offer := range categoryOffers {
		var category models.Category
		var subcategory models.SubCategory
		if err := db.Db.Where("sub_category_id = ?",offer.CategoryID).First(&subcategory).Error; err != nil {
			log.Println("Failed to load subcategory details")
			c.Redirect(http.StatusTemporaryRedirect, "/admin")
			return
		}

		if err := db.Db.Where("category_id = ?",subcategory.CategoryID).First(&category).Error; err != nil {
			log.Println("Failed to load category details")	
			c.Redirect(http.StatusTemporaryRedirect, "/admin")
			return
		}
		offers = append(offers, OfferView{
			ID:                 offer.ID,
			OfferName:          offer.OfferName,
			OfferType:          "Category",
			TargetName:         offer.CategorryName + "("+category.CategoryName+")",
			DiscountPercentage: offer.DiscountPercentage,
			CreatedAt:          offer.CreatedAt,
			EndAt:              offer.EndAt,
			Active:             offer.Active,
		})
	}

	totalPages := int(math.Ceil(float64(total)/ float64(limit)))

	c.HTML(http.StatusOK, "admin_offers.html", gin.H{
		"user":    name,
		"Offers":  offers,
		"page":    page,	
		"limit":   limit,
		"totalPages":   totalPages,
		"search": keyword,
	})
}

func AddOfferProductPage(c *gin.Context){
	var products []models.Product

	if err := db.Db.Find(&products).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "admin_productoffer.html", gin.H{"error": "Could not retrieve products, please try again later"})
		return
	}

	c.HTML(http.StatusOK, "admin_productoffer.html", gin.H{
		"Products": products,	
	})
}

func AddOfferCategoryPage(c *gin.Context) {
	
	var subcategories []models.SubCategory

	type Response struct {
		SubCategoryID 		int
		SubCategoryName		string 
		CategoryName		string 
	}


	if err := db.Db.Find(&subcategories).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "admin_categoryoffer.html", gin.H{"error": "Could not retrieve subcategories, please try again later"})
		return
	}

	response := make([]Response,len(subcategories))

	for i,subitem := range subcategories {

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

	c.HTML(http.StatusOK, "admin_categoryoffer.html", gin.H{
		"Subcategories": response,
	})
}