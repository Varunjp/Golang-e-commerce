package db

import (
	"log"
	"os"

	"first-project/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB

func DbInit(){
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading env file:",err)
	}

	Db, err = gorm.Open(postgres.Open(os.Getenv("dns")), &gorm.Config{})
	if err != nil{
		log.Fatal("Error loading database",err)
	}

	autoerr := Db.AutoMigrate(
		&models.User{},
		&models.Admin{},
		&models.Category{},
		&models.Address{},
		&models.SubCategory{},
		&models.Product{},
		&models.Product_Variant{},
		&models.Product_image{},
		&models.Review{},
		&models.Banner{},
		&models.OTPVerification{},
	)

	if autoerr != nil{
		log.Fatal("Migration failed",autoerr)
	}
}