package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct{
	ID 			uint   		`gorm:"primarykey;autoIncrement"`
	Username 	string 		`gorm:"not null"`
	Email 		string 		`gorm:"not null; unique; index"`
	Password 	string 		`gorm:"not null"`
	Phone 		string 		`gorm:"not null; unique"`
	Status 		string 		`gorm:"check(status IN('Active', 'Inactive', 'Blocked'))"`
	Addresses 	[]Address 	`gorm:"constraint:OnDelete:CASCADE; foreignKey:UserID"` 
	Created_at 	time.Time
	Updated_at 	time.Time
	Reviews		[]Review	`gorm:"constraint:OnDelete:CASCADE; foreignKey:UserID"`
}

type Address struct {
	AddressID 		int 			`gorm:"primarykey;autoIncrement"`
	UserID 			uint 			`gorm:"not null; index;"`
	User 			User 			`gorm:"constraint:OnDelete:CASCADE;"`
	AddressLine1 	string
	AddressLine2 	string
	Country 		string 
	City 			string
	State 			string  
	PostalCode 		string
	Deleted_at 		gorm.DeletedAt 	`gorm:"index"` 
}

type Admin struct {
	ID 			uint   		`gorm:"primarykey"`
	Username 	string 		`gorm:"not null"`
	Email 		string 		`gorm:"not null; unique; index"`
	Password 	string 		`gorm:"not null"`
	Phone 		string 		`gorm:"not null; unique"`
	Status 		string 		`gorm:"check(status IN('Active', 'Inactive', 'Blocked'))"`
	Created_at 	time.Time
	Updated_at 	time.Time
}

type Category struct {
	CategoryID 			uint 			`gorm:"primarykey" json:"category_id"`
	CategoryName 		string 			`json:"name"`
	CreateAt 			time.Time
	SubCategories  		[]SubCategory 	`gorm:"constraint:OnDelete:CASCADE; foreignKey:CategoryID"` 
	DeletedAt 			gorm.DeletedAt 	`gorm:"index"`
}

type SubCategory struct{
	SubCategoryID 		uint 				`gorm:"primarykey"`
	SubCategoryName 	string 				`gorm:"not null"`
	CategoryID 			uint  				`gorm:"not null;index"`
	Category 			Category 			`gorm:"constraint:OnDelete:CASCADE;"`
	Products			[]Product 			`gorm:"foreignkey:SubCategoryID"`
	Deleted_at 			gorm.DeletedAt 		`gorm:"index"`
}

type Product struct {
	ProductID				uint				`gorm:"primarykey"`
	ProductName				string				`gorm:"not null" json:"name"`
	Description				string				`json:"description"`
	SubCategoryID			uint				`gorm:"not null"`
	SubCategory				SubCategory			`gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt 				time.Time
	Product_variants 		[]Product_Variant 	`gorm:"constraint:OnDelete:CASCADE;foreignkey:ProductID"`
	Reviews					[]Review 			`gorm:"constraint:OnDelete:CASCADE;foreignkey:ProductID"`
	DeletedAt				gorm.DeletedAt 		`gorm:"index"`
}

type Product_Variant struct {
	ID	uint				`gorm:"primarykey"`
	ProductID			uint				`gorm:"index"`
	Variant_name		string				`gorm:"not null"`
	Size				string
	Stock				int
	Price				float64				`gorm:"not null"`
	Tax					float64				`gorm:"not null"`
	CreatedAt			time.Time
	UpdatedAt			time.Time
	Product 			Product				`gorm:"constraint: OnDelete:CASCADE"`
	Product_images		[]Product_image		`gorm:"constraint:OnDelete:CASCADE;foreignkey:ProductVariantID"`
	DeletedAt			gorm.DeletedAt 		`gorm:"index"`
}

type Product_image struct {
	ProductImageID					uint				`gorm:"primarykey"`
	ProductVariantID				uint				`gorm:"index"`
	Image_url						string				`gorm:"index"`
	Is_primary						bool
	Order_no						int
	CreatedAt						time.Time
	Product_Variant					Product_Variant		`gorm:"constraint: OnDelete:CASCADE"`
}

type Review struct {
	ID 				uint		`gorm:"primarykey"`
	UserID			uint
	ProductID		uint
	Rating			int
	Comment			string
	CreatedAt		time.Time 
	UpdatedAt		time.Time 
	User			User		`gorm:"constraint: OnDelete:CASCADE"`
	Product			Product		`gorm:"constraint: OnDelete:CASCADE"`
}

type Banner struct {
	ID				uint		`gorm:"primarykey"`
	Title			string
	ImageUrl		string		`gorm:"not null"`
	RedirectURL		string
	Active			bool
	CreatedAt		time.Time
	UpdateAt		time.Time 
}

type OTPVerification struct {
	ID 		uint		`gorm:"primarykey"`
	Email	string		`gorm:"not null"`
	OTP 	string		`gorm:"not null"`
	ExpiresAt	time.Time	`gorm:"not null"`
	IsUsed		bool		`gorm:"default:false"`		
}