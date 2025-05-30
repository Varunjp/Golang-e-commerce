package helper

import (
	db "first-project/DB"
	"first-project/models"
)

func DeleteAllUnderCategory(categoryID string) error {

	var subCategories []models.SubCategory

	if err := db.Db.Where("category_id = ?", categoryID).Find(&subCategories).Error; err != nil {
		return err
	}

	for _, subCat := range subCategories{

		var products []models.Product
		
		if err := db.Db.Where("sub_category_id = ?",subCat.SubCategoryID).Find(&products).Error; err != nil{
			return err 
		}

		for _,product := range products{

			var variants []models.Product_Variant
			
			if err := db.Db.Where("product_id = ?",product.ProductID).Find(&variants).Error; err != nil {
				return err 
			}

			ids := []uint{}

			for _,data := range variants{
				ids = append(ids, data.ID)
			}

			// for _, variant := range variants{

			// 	if err := db.Db.Where("product_variant_id = ?",variant.ID).Delete(&models.Product_image{}).Error; err != nil{
			// 		return err
			// 	}
			// }

			if err := db.Db.Model(&models.Product_Variant{}).Where("id IN ?",ids).Update("is_active",false).Error; err != nil{
				return err 
			}
		}

		if err := db.Db.Model(&models.SubCategory{}).Where("sub_category_id = ?",subCat.SubCategoryID).Update("is_blocked",true).Error; err != nil{
			return err 
		}

	}

	// if err := db.Db.Where("category_id = ?",categoryID).Delete(&models.SubCategory{}).Error; err != nil{
	// 	return err 
	// }

	return db.Db.Model(&models.Category{}).Where("category_id = ?",categoryID).Update("is_blocked",true).Error

}


func DeleteAllUnderSubCategory(subCategoryID string)error{

	var products []models.Product
		
	if err := db.Db.Where("sub_category_id = ?",subCategoryID).Find(&products).Error; err != nil{
		return err 
	}

	for _,product := range products{

		var variants []models.Product_Variant
			
		if err := db.Db.Where("product_id = ?",product.ProductID).Find(&variants).Error; err != nil {
			return err 
		}

		ids := []uint{}

		for _,data := range variants{
			ids = append(ids, data.ID)
		}

		// for _, variant := range variants{

		// 	if err := db.Db.Where("product_variant_id = ?",variant.ID).Delete(&models.Product_image{}).Error; err != nil{
		// 		return err
		// 	}
		// }

		if err := db.Db.Model(&models.Product_Variant{}).Where("id IN ?",ids).Update("is_active",false).Error; err != nil{
			return err 
		}

		// for _, variant := range variants{

		// 	if err := db.Db.Where("product_variant_id = ?",variant.ID).Delete(&models.Product_image{}).Error; err != nil{
		// 		return err
		// 	}

		// }

		// if err := db.Db.Where("product_id = ?",product.ProductID).Delete(&models.Product_Variant{}).Error; err != nil{
		// 	return err 
		// }
	}	
	
	return db.Db.Model(&models.SubCategory{}).Where("sub_category_id = ?",subCategoryID).Update("is_blocked",true).Error
}