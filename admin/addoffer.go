package admin

import (
	db "first-project/DB"
	"first-project/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AddProductOffer(c *gin.Context) {
	Errors := make(map[string]string)
	var products []models.Product
	var existingOffers []models.ProductOffer

	if err := db.Db.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Could not retrieve products"})
		return
	}
	
	productIDStr := c.PostForm("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil || productID <= 0 {
		Errors["product_id"] = "Invalid product ID"
	}
	offerName := c.PostForm("offer_name")	
	discountPercentageStr := c.PostForm("discount")
	discountPercentage, err := strconv.ParseFloat(discountPercentageStr, 64)
	if err != nil || discountPercentage < 0 || discountPercentage > 50 {
		Errors["discount"] = "Invalid discount percentage"
	}

	startDateStr := c.PostForm("start_date")
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {	
		Errors["start_date"] = "Invalid start date format"
	}
	endDateStr := c.PostForm("end_date")
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil || endDate.Before(startDate) {
		Errors["end_date"] = "Invalid end date format or end date is before start date"
	}

	if err := db.Db.Where("product_id = ? AND active = true", productID).Find(&existingOffers).Error; err != nil {
		log.Println("Error checking existing offers:", err)
	}
	
	if len(existingOffers) > 0 {
		Errors["product_id"] = "This product already has an active offer"
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status":"error","errors": Errors})
		return
	}

	offer := models.ProductOffer{
		ProductID:          uint(productID),
		OfferName:          offerName,
		DiscountPercentage: discountPercentage,
		CreatedAt:          startDate,
		EndAt:            endDate,
		Active:             true,
	}

	if err := db.Db.Create(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status":"error","message": "Could not create offer, please try again later"})
		return
	}	

	c.JSON(http.StatusOK, gin.H{"redirect":"/admin/offers"})
}

func AddCategoryOffer(c *gin.Context){
	Errors := make(map[string]string)
	var subcategories []models.SubCategory
	var subcate models.SubCategory
	var existingOffers []models.CategoryOffer

	if err := db.Db.Find(&subcategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Could not retrieve subcategories"})
		return
	}

	subcategoryIDStr := c.PostForm("subcategory_id")
	subcategoryID, err := strconv.Atoi(subcategoryIDStr)

	if err := db.Db.Where("sub_category_id = ?", subcategoryID).First(&subcate).Error; err != nil {
		log.Println("Error retrieving subcategory:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Could not retrieve subcategory"})
		return
	}

	if err != nil || subcategoryID <= 0 {
		Errors["subcategory_id"] = "Invalid subcategory ID"
	}
	offerName := c.PostForm("offer_name")	
	if strings.TrimSpace(offerName) == "" {
		Errors["offer_name"] = "Offer name is required"
	}
	discountPercentageStr := c.PostForm("discount")
	discountPercentage, err := strconv.ParseFloat(discountPercentageStr, 64)
	if err != nil || discountPercentage < 0 || discountPercentage > 50 {
		Errors["discount"] = "Invalid discount percentage"
	}

	startDateStr := c.PostForm("start_date")
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {	
		Errors["start_date"] = "Invalid start date format"
	}
	endDateStr := c.PostForm("end_date")
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil || endDate.Before(startDate) {
		Errors["end_date"] = "Invalid end date format or end date is before start date"
	}

	if err := db.Db.Where("category_id = ? AND active = true", subcategoryID).Find(&existingOffers).Error; err != nil {
		log.Println("Error checking existing offers:", err)
	}

	if len(existingOffers) > 0 {
		Errors["subcategory_id"] = "This subcategory already has an active offer"
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status":"error","errors": Errors})
		return
	}

	offer := models.CategoryOffer{
		CategoryID:      uint(subcategoryID),
		OfferName:          offerName,
		CategorryName: subcate.SubCategoryName,
		DiscountPercentage: discountPercentage,
		CreatedAt:          startDate,
		EndAt:              endDate,
		Active:             true,
	}

	if err := db.Db.Create(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status":"error","message": "Could not create offer, please try again later"})
		return
	}	

	c.JSON(http.StatusOK, gin.H{"redirect":"/admin/offers"})
}