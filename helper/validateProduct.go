package helper

import (
	db "first-project/DB"
	"first-project/models"
)

func ValidateProduct(ID uint, quantity int) bool {

	var Product models.Product_Variant

	if err := db.Db.First(&Product,ID).Error; err != nil{

		return false 

	}

	if Product.Stock == 0 || Product.Stock < quantity {
		return false 
	}

	return true

}