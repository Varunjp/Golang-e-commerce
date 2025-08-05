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

func EditProductOfferPage(c *gin.Context) {
	var offer models.ProductOffer
	var products []models.Product
	offerIDStr := c.Param("id")

	if err := db.Db.Where("id = ?", offerIDStr).First(&offer).Error; err != nil {
		c.HTML(http.StatusInternalServerError,"edit_productOffer.html", gin.H{ "error": "Offer not found,please try again later"})
		return
	}

	if err := db.Db.Find(&products).Error; err != nil {
		c.HTML(http.StatusInternalServerError,"edit_productOffer.html", gin.H{ "error": "Could not retrieve products,please try again later"})
		return
	}

	c.HTML(http.StatusOK, "edit_productOffer.html", gin.H{
		"offer":    offer,
		"Products": products,
	})
}

func EditProductOffer(c *gin.Context) {
	var offer models.ProductOffer
	offerIDStr := c.Param("id")
	Errors := make(map[string]string)
	var existingOffers []models.ProductOffer

	if err := db.Db.Where("id = ?", offerIDStr).First(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Offer not found"})
		return
	}

	productIDStr := c.PostForm("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil || productID <= 0 {
		Errors["product_id"] = "Invalid product ID"
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
	if err != nil || endDate.Before(startDate) || endDate.Equal(startDate){
		Errors["end_date"] = "Invalid end date format or end date is before or equal to start date"
	}

	if err := db.Db.Where("product_id = ? AND id != ? AND active = true", productID, offerIDStr).Find(&existingOffers).Error; err != nil {
		log.Println("Error checking existing offers:", err)
	}

	if len(existingOffers) > 0 {
		Errors["product_id"] = "This product already has an active offer"
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "errors": Errors})
		return
	}

	offer.ProductID = uint(productID)
	offer.OfferName = offerName
	offer.DiscountPercentage = discountPercentage
	offer.CreatedAt = startDate
	offer.EndAt = endDate
	offer.Active = true

	if err := db.Db.Save(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update offer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"redirect":"/admin/offers"})
}

func ToggleProductOffer(c *gin.Context) {
	offerIDStr := c.Param("id")
	var offer models.ProductOffer

	if err := db.Db.Where("id = ?", offerIDStr).First(&offer).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "edit_productOffer.html", gin.H{"error": "Offer not found, please try again later"})
		return
	}

	// Toggle the Active status
	offer.Active = !offer.Active

	if err := db.Db.Save(&offer).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "edit_productOffer.html", gin.H{"error": "Failed to update offer status, please try again later"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/offers")
}

func EditCategoryOfferPage(c *gin.Context) {
	var subcategories []models.SubCategory
	offerIDStr := c.Param("id")
	var offer models.CategoryOffer

	if err := db.Db.Where("id = ?", offerIDStr).First(&offer).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "edit_categoryOffer.html", gin.H{"error": "Offer not found, please try again later"})
		return
	}

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

	c.HTML(http.StatusOK, "edit_categoryOffer.html", gin.H{
		"Subcategories": response,
		"offer": offer,
	})
}

func EditCategoryOffer(c *gin.Context) {

	offerId := c.Param("id")
	var offer models.CategoryOffer
	var subcategory models.SubCategory
	Errors := make(map[string]string)
	var existingOffers []models.CategoryOffer

	if err := db.Db.Where("id = ?", offerId).First(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Offer not found"})
		return
	}

	subcategoryIDStr := c.PostForm("subcategory_id")
	subcategoryID, err := strconv.Atoi(subcategoryIDStr)
	if err != nil || subcategoryID <= 0 {
		Errors["subcategory_id"] = "Invalid subcategory ID"
	}

	if err := db.Db.Where("sub_category_id = ?", subcategoryID).First(&subcategory).Error; err != nil {
		log.Println("Error retrieving subcategory:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Could not retrieve subcategory"})
		return
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
	if err != nil || endDate.Before(startDate) || endDate.Equal(startDate) {
		Errors["end_date"] = "Invalid end date format or end date is before or equal to start date"
	}

	if err := db.Db.Where("category_id = ? AND id != ? AND active = true", subcategoryID,offerId).Find(&existingOffers).Error; err != nil {
		log.Println("Error checking existing offers:", err)	
	}

	if len(existingOffers) > 0 {
		Errors["subcategory_id"] = "This subcategory already has an active offer"
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "errors": Errors})
		return
	}

	offer.CategoryID = uint(subcategoryID)
	offer.OfferName = offerName
	offer.CategorryName = subcategory.SubCategoryName
	offer.DiscountPercentage = discountPercentage
	offer.CreatedAt = startDate
	offer.EndAt = endDate
	offer.Active = true

	if err := db.Db.Save(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update offer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"redirect": "/admin/offers"})

}

func ToggleCategoryOffer(c *gin.Context) {	
	offerIDStr := c.Param("id")
	var offer models.CategoryOffer

	if err := db.Db.Where("id = ?", offerIDStr).First(&offer).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "edit_categoryOffer.html", gin.H{"error": "Offer not found, please try again later"})
		return
	}

	// Toggle the Active status
	offer.Active = !offer.Active

	if err := db.Db.Save(&offer).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "edit_categoryOffer.html", gin.H{"error": "Failed to update offer status, please try again later"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/offers")
}