package helper

import (
	db "first-project/DB"
	"first-project/models"
	"first-project/models/responsemodels"
	"fmt"
)

func GetHomePage() ([]responsemodels.HomePage, string, error){
	
	var Products []models.Product_Variant
	var Banner models.Banner

	if err := db.Db.Where("active").Order("created_at DESC").First(&Banner).Error; err != nil{
		Banner.ImageUrl = ""
	}

	if err := db.Db.Where("deleted_at IS NULL AND stock > 0").Order("id DESC").Limit(10).Find(&Products).Error; err != nil{
		return []responsemodels.HomePage{},Banner.ImageUrl,fmt.Errorf("no products found")
	}

	result := make([]responsemodels.HomePage,len(Products))

	for i, product := range Products{

		var Image models.Product_image

		var Rating struct {
			ID		uint	`gorm:"column:id"`
			Rating	int		`gorm:"column:avg"`
		}

		if err := db.Db.Where("order_no = 1 AND product_variant_id = ?",product.ID).First(&Image).Error; err != nil{
			Image.Image_url = ""
		}

		if err := db.Db.Select("id,AVG(rating)").Where("product_id = ?",product.ProductID).Table("reviews").Group("id,product_id").Find(&Rating).Error; err != nil{
			Rating.Rating = 0
		}

		result[i] = responsemodels.HomePage{
			ID: product.ID,
			ImageURL: Image.Image_url,
			Name: product.Variant_name,
			Rating: Rating.Rating,
			Price: int(product.Price),
		}
	}

	return result,Banner.ImageUrl,nil
}