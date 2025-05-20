package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct{
	ID 						uint   						`gorm:"primarykey;autoIncrement"`
	Username 				string 						`gorm:"not null"`
	Email 					string 						`gorm:"not null; unique; index"`
	Password 				string 						`gorm:"not null"`
	Phone 					string 		
	Status 					string 						`gorm:"check(status IN('Active', 'Inactive', 'Blocked'))"`
	Addresses 				[]Address 					`gorm:"constraint:OnDelete:CASCADE; foreignKey:UserID"` 
	Created_at 				time.Time
	Updated_at 				time.Time
	Reviews					[]Review					`gorm:"constraint:OnDelete:CASCADE; foreignKey:UserID"`
	Orders					[]Order						`gorm:"constraint:ONDELETE:CASCADE; foreignKey:UserID"`
	ProfileImages			[]ProfileImage				`gorm:"constraint:ONDELETE:CASCADE; foreignKey:UserID"`
	CartItems				[]CartItem					`gorm:"constraint:ONDELETE:CASCADE; foreignkey:UserID"`
	WishLists				[]WishList						`gorm:"constraint:ONDELETE:CASCADE; foreignkey:UserID"`

	DeletedAt 				gorm.DeletedAt	
}

type ProfileImage struct{
	ID				uint				`gorm:"primarykey;autoIncrement"`
	UserID			uint				`gorm:"index"`
	ImageUrl		string	
	CreateAt		time.Time
	User			User				`gorm:"constraint:ONDELETE:CASCADE"`
	DeletedAt 		gorm.DeletedAt
}

type Address struct {
	AddressID 		uint 						`gorm:"primarykey;autoIncrement"`
	UserID 			uint 						`gorm:"not null; index;"`
	User 			User 						`gorm:"constraint:OnDelete:CASCADE;"`
	AddressLine1 	string
	AddressLine2 	string
	Country 		string 
	City 			string
	State 			string  
	PostalCode 		string
	Orders			[]Order						`gorm:"constraint:ONDELETE:CASCADE;foreignKey:AddressID"`
	DeletedAt		gorm.DeletedAt
}


type CartItem struct {
	ID 				uint		`gorm:"primarykey;autoIncrement"`
	UserID			uint 		`gorm:"index"`
	ProductID 		uint 		`gorm:"index"`
	Quantity		int 		`gorm:"not null"`
	Price 			float64 	`gorm:"not null"`
	AddAt 			time.Time 
	User			User		`gorm:"constraint:ONDELETE:CASCADE"`
	Product 		Product_Variant		`gorm:"constraint:ONDELETE:CASCADE"`
}

type Order struct{
	ID 						uint		`gorm:"primarykey;autoIncrement"`
	UserID					uint		`gorm:"not null; index"`
	AddressID				uint		`gorm:"not null"`
	PaymentID				string
	OrderDate				time.Time   `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	TotalTax				float64		`gorm:"not null"`
	DiscountTotal			float64
	ShippingFee				float64
	TotalAmount				float64		`gorm:"not null"`
	Status					string
	PaymentStatus			string 		
	CreateAt				time.Time
	BadgeClass 				string
	Reason 					string 	
	OrderItems				[]OrderItem  `gorm:"constraint:ONDELETE:CASCADE;foreignKey:OrderID"`
	Address 				Address 	`gorm:"constraint:ONDELETE:CASCADE"`
}

type OrderItem struct {
	ID			uint 	`gorm:"primarykey;autoIncrement"`
	OrderID		uint	`gorm:"index"`
	UserID 		uint 	`gorm:"index"`
	ProductID	uint 	`gorm:"index"`
	Quantity	int 
	Price 		float64	`gorm:"not null"`
	Discount 	float64	
	Order 		Order	`gorm:"constraint:ONDELETE:CASCADE"`

}

type WishList struct {
	ID			uint 		`gorm:"primarykey;autoIncrement"`
	UserID 		uint 		`gorm:"not null"`
	ProductID 	uint 		`gorm:"not null"`
	CreatedAt 	time.Time
}

type Admin struct {
	ID 			uint   		`gorm:"primarykey;autoIncrement"`
	Username 	string 		`gorm:"not null"`
	Email 		string 		`gorm:"not null; unique; index"`
	Password 	string 		`gorm:"not null"`
	Phone 		string 		`gorm:"not null; unique"`
	Status 		string 		`gorm:"check(status IN('Active', 'Inactive', 'Blocked'))"`
	Created_at 	time.Time
	Updated_at 	time.Time
}

type Category struct {
	CategoryID 			uint 			`gorm:"primarykey;autoIncrement" json:"category_id"`
	CategoryName 		string 			`json:"name"`
	CreateAt 			time.Time
	IsBlocked			bool			`gorm:"default:false"`
	SubCategories  		[]SubCategory 	`gorm:"constraint:OnDelete:CASCADE; foreignKey:CategoryID"` 
	DeletedAt 			gorm.DeletedAt 	`gorm:"index"`
}

type SubCategory struct{
	SubCategoryID 		uint 				`gorm:"primarykey;autoIncrement"`
	SubCategoryName 	string 				`gorm:"not null"`
	CategoryID 			uint  				`gorm:"not null;index"`
	IsBlocked			bool				`gorm:"default:false"`
	Category 			Category 			`gorm:"constraint:OnDelete:CASCADE;"`
	Products			[]Product 			`gorm:"foreignkey:SubCategoryID"`
	Deleted_at 			gorm.DeletedAt 		`gorm:"index"`
}

type Product struct {
	ProductID				uint				`gorm:"primarykey;autoIncrement"`
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
	ID					uint				`gorm:"primarykey;autoIncrement"`
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
	CartItems				[]CartItem			`gorm:"constraint:OnDelelte:CASCADE;foreignkey:ProductID"`
	WishLists				[]WishList					`gorm:"constraint:OnDelelte:CASCADE;foreignkey:ProductID"`
	DeletedAt			gorm.DeletedAt 		`gorm:"index"`
}

type Product_image struct {
	ProductImageID					uint				`gorm:"primarykey;autoIncrement"`
	ProductVariantID				uint				`gorm:"index"`
	Image_url						string				`gorm:"index"`
	Is_primary						bool
	Order_no						int
	CreatedAt						time.Time
	Product_Variant					Product_Variant		`gorm:"constraint: OnDelete:CASCADE"`
	DeleteAt						gorm.DeletedAt		`gorm:"index"`
}

type Review struct {
	ID 				uint		`gorm:"primarykey;autoIncrement"`
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
	ID				uint		`gorm:"primarykey;autoIncrement"`
	Title			string
	ImageUrl		string		`gorm:"not null"`
	RedirectURL		string
	Active			bool
	CreatedAt		time.Time
	UpdateAt		time.Time 
}

type OTPVerification struct {
	ID 					uint		`gorm:"primarykey;autoIncrement"`
	Email				string		`gorm:"not null"`
	OTP 				string		`gorm:"not null"`
	ExpiresAt			time.Time	`gorm:"not null"`
	IsUsed				bool		`gorm:"default:false"`		
}