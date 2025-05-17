package helper

import (
	db "first-project/DB"
	"first-project/models"

	"gorm.io/gorm"
)

func UpdateUserImage(userID int, imagePath string) error {

	var ProfileImage models.ProfileImage

	if err := db.Db.First(&ProfileImage,userID).Error; err != nil{

		if err == gorm.ErrRecordNotFound{
			newImage := models.ProfileImage{
				UserID: uint(userID),
				ImageUrl: imagePath,
			}

			if err := db.Db.Create(&newImage).Error; err != nil{
				return err 
			}
		}else{
			return err 
		}

	}

	ProfileImage.ImageUrl = imagePath

	return db.Db.Save(&ProfileImage).Error

}