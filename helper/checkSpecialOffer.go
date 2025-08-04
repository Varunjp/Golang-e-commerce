package helper

import (
	db "first-project/DB"
	"first-project/models"
	"log"
	"time"
)

func CheckSpecialOffer(id uint) float64 {

	var productOffer models.ProductOffer
	var categoryOffer models.CategoryOffer
	var product models.Product_Variant

	db.Db.Preload("Product").Where("id = ?",id).First(&product)

	today := time.Now().Format("2006-01-02")

	if err := db.Db.Where("created_at <= ? AND product_id = ? AND active = true ", today+" 23:00:00", product.ProductID).First(&productOffer).Error; err != nil {
		log.Println("No offer available ", err)
	}

	if err := db.Db.Where("created_at <= ? AND category_id = ? AND active = true", today+" 23:00:00", product.Product.SubCategoryID).First(&categoryOffer).Error; err != nil {
		log.Println("No offer available as category ", err)
	}

	specialDiscountPercent := 0.0

	if productOffer.DiscountPercentage > categoryOffer.DiscountPercentage {
		specialDiscountPercent = productOffer.DiscountPercentage
	} else if categoryOffer.DiscountPercentage > productOffer.DiscountPercentage {
		specialDiscountPercent = categoryOffer.DiscountPercentage
	} else {
		specialDiscountPercent = 0.0
	}

	return specialDiscountPercent
}