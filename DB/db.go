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

func DbInit() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(err)
	}

	dsn := os.Getenv("dns")

	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	sqlDB, err := Db.DB()
	if err != nil {
		log.Fatalf("Error getting SQL DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	log.Println("Database connected successfully")

	autoerr := Db.AutoMigrate(
		&models.User{},
		&models.Admin{},
		&models.ProfileImage{},
		&models.Category{},
		&models.Address{},
		&models.WishList{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.SubCategory{},
		&models.Product{},
		&models.Product_Variant{},
		&models.Product_image{},
		&models.Banner{},
		&models.OTPVerification{},
		&models.Coupons{},
		&models.UsedCoupon{},
		&models.Wallet{},
		&models.WalletTransaction{},
		&models.OrderAddress{},
		&models.Review{},
		&models.ProductOffer{},
		&models.CategoryOffer{},
	)

	if autoerr != nil {
		log.Fatal("Migration failed", autoerr)
	}
}
