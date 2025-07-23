package helper

import (
	db "first-project/DB"
	"first-project/models"
	"log"
)

type Response struct {
	SubCategoryID   int
	SubCategoryName string
	CategoryName    string
}

func GetCategories()([]Response,error) {
	var subcat []models.SubCategory
	
	

	if err := db.Db.Where("is_blocked = ?", false).Find(&subcat).Error; err != nil {
		return []Response{},err
	}

	response := make([]Response, len(subcat))
	for i, subitem := range subcat {

		var Category models.Category

		if err := db.Db.Where("category_id = ?", subitem.CategoryID).First(&Category).Error; err != nil {
			log.Println("Failed to load category details")
			return []Response{},err
		}

		response[i] = Response{
			SubCategoryID:   int(subitem.SubCategoryID),
			SubCategoryName: subitem.SubCategoryName,
			CategoryName:    Category.CategoryName,
		}

	}

	return response,nil 
}