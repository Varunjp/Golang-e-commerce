package helper

import (
	db "first-project/DB"
	"first-project/models"
	"log"
	"os"

	"gorm.io/gorm"
)

func UpdateUserImage(userID int, imagePath string) error {

	var ProfileImage models.ProfileImage

	if err := db.Db.Where("user_id = ?",userID).First(&ProfileImage).Error; err != nil{

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

	if ProfileImage.ImageUrl != ""{
		err := os.Remove(ProfileImage.ImageUrl)

		if err != nil{
			log.Println("Error removing old image:", err)
		}
	}

	ProfileImage.ImageUrl = imagePath

	return db.Db.Save(&ProfileImage).Error

}