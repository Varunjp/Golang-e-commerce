package helper

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"log"
	"time"
)

func GetSpecialOffer() []string {
	var productoffer []models.ProductOffer
	var categoryOffer []models.CategoryOffer
	today := time.Now().Format("2006-01-02")
	
	if err := db.Db.Where("created_at <= ? AND active = true ", today+" 23:00:00").Find(&productoffer).Error; err != nil{
		log.Println("error in product offers :",err)
	}

	if err := db.Db.Where("created_at <= ? AND active = true ", today+" 23:00:00").Find(&categoryOffer).Error; err != nil{
		log.Println("error in category offers :",err)
	}

	offersSlice := make([]string,0)

	for _,poffer := range productoffer{
		var product models.Product
		db.Db.Where("product_id = ?",poffer.ProductID).First(&product)
		offer := fmt.Sprintf("%s sale %v%% off on %s", poffer.OfferName, poffer.DiscountPercentage, product.ProductName)
		offersSlice = append(offersSlice, offer)
	}

	for _,coffer := range categoryOffer{
		var subcategory models.SubCategory
		db.Db.Preload("Category").Where("sub_category_id = ?",coffer.CategoryID).First(&subcategory)
		offer := fmt.Sprintf("%s sale %v%% off on %s (%s)",coffer.OfferName,coffer.DiscountPercentage,coffer.CategorryName,subcategory.Category.CategoryName)
		offersSlice = append(offersSlice, offer)
	}

	return offersSlice
}