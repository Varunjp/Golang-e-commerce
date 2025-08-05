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
	ReferralCode			string						`gorm:"unique"`
	ReferredBy 				string 
	DeletedAt 				gorm.DeletedAt
	Created_at 				time.Time
	Updated_at 				time.Time
	
	// Associations
	Orders					[]Order						`gorm:"constraint:ONDELETE:CASCADE;"`
	ProfileImages			[]ProfileImage				`gorm:"constraint:ONDELETE:CASCADE;"`
	CartItems				[]CartItem					`gorm:"constraint:ONDELETE:CASCADE;"`
	WishLists				[]WishList					`gorm:"constraint:ONDELETE:CASCADE;"`
	Reviews 				[]Review					`gorm:"constraint:ONDELETE:CASCADE;"`
	Addresses 				[]Address 					`gorm:"constraint:OnDelete:CASCADE;"` 
}

type ProfileImage struct{
	ID				uint				`gorm:"primarykey;autoIncrement"`
	UserID			uint				`gorm:"index"`
	ImageUrl		string	
	CreateAt		time.Time
	DeletedAt 		gorm.DeletedAt

	// Associations
	User			User				`gorm:"constraint:ONDELETE:CASCADE"`
}

type Address struct {
	AddressID 		uint 						`gorm:"primarykey;autoIncrement"`
	UserID 			uint 						`gorm:"not null; index;"`
	AddressLine1 	string
	AddressLine2 	string
	Country 		string 
	City 			string
	State 			string  
	PostalCode 		string
	IsDefault		bool
	DeletedAt		gorm.DeletedAt
	
	// Associations
	User 			User 						`gorm:"constraint:ONDELETE:CASCADE;"`
}

type OrderAddress struct {
	ID 				uint 						`gorm:"primarykey;autoIncrement"`
	OrderID 		uint 						`gorm:"not null; index"`
	AddressLine1 	string
	AddressLine2 	string
	Country 		string 
	City 			string
	State 			string  
	PostalCode 		string
}

type CartItem struct {
	ID 				uint					`gorm:"primarykey;autoIncrement"`
	UserID			uint 					`gorm:"index"`
	ProductID 		uint 					`gorm:"index"`
	Quantity		int 					`gorm:"not null"`
	Price 			float64 				`gorm:"not null"`
	AddAt 			time.Time 
	User			User					`gorm:"constraint:ONDELETE:CASCADE"`
	Product 		Product_Variant			`gorm:"constraint:ONDELETE:CASCADE"`
}

type Order struct{
	ID 						uint		`gorm:"primarykey;autoIncrement"`
	OrderID					string 		`gorm:"unique"`
	UserID					uint		`gorm:"not null; index"`
	AddressID				uint		`gorm:"not null"`
	PaymentID				string
	OrderDate				time.Time   `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	TotalTax				float64		`gorm:"not null"`
	SubTotal 				float64
	DiscountTotal			float64
	ShippingFee				float64
	TotalAmount				float64		`gorm:"not null"`
	Status					string
	PaymentMethod 			string 		
	PaymentStatus			string 		
	CreateAt				time.Time
	BadgeClass 				string
	Reason 					string 	
	DeletedAt 				gorm.DeletedAt
	
	// Associations
	OrderItems				[]OrderItem  `gorm:"constraint:ONDELETE:CASCADE;foreignKey:OrderID"`
	Reviews 				[]Review	`gorm:"constraint:ONDELETE:CASCADE;foreignKey:OrderID"`
}

type OrderItem struct {
	ID					uint 	`gorm:"primarykey;autoIncrement"`
	OrderID				uint	`gorm:"index"`
	UserID 				uint 	`gorm:"index"`
	ProductID			uint 	`gorm:"index"`
	Quantity			int 
	Price 				float64	`gorm:"not null"`
	Tax					float64
	Discount 			float64	
	Status 				string
	PaymentStatus 		string 
	Reason 				string 
	DeletedAt 			gorm.DeletedAt

	// Associations
	Order 				Order	`gorm:"constraint:ONDELETE:CASCADE"`
}

type WishList struct {
	ID			uint 		`gorm:"primarykey;autoIncrement"`
	UserID 		uint 		`gorm:"not null"`
	ProductID 	uint 		`gorm:"not null"`
	CreatedAt 	time.Time
}


type Coupons struct {
	ID					uint 		`gorm:"primarykey;autoIncrement"`
	Code 				string 		`gorm:"not null"`
	Type 				string 
	UserID				uint
	CouponID			uint
	Description			string 	
	Discount			float64
	MinAmount 			float64
	MaxAmount			float64
	IsActive			bool 
	CreatedAt			time.Time
	CategoryID 			uint 		
}

type UsedCoupon struct {
	ID 				uint 		`gorm:"primarykey;autoIncrement"`
	UserID 			uint 		`gorm:"not null"`
	CouponID		uint 		`gorm:"not null"`
	OrderID			uint 		`gorm:"not null"`
}

type Wallet struct {
	ID 				uint 		`gorm:"primarykey;autoIncrement"`
	UserID 			uint 		`gorm:"uniqueIndex"`
	Balance			float64     
	UpdatedAt 		time.Time 
}

type WalletTransaction struct {
	ID 				uint 		`gorm:"primarykey;autoIncrement"`
	UserID 			uint 
	OrderID 		uint 
	OrderItemID 	uint
	Amount 			float64 	
	Type 			string 
	Description 	string 
	RefundStatus	bool 		`gorm:"default:false"`
	Status 			bool 		
	CreatedAt 		time.Time
	DeletedAt 		gorm.DeletedAt
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
	DeletedAt 			gorm.DeletedAt 	`gorm:"index"`
	
	// Associations
	SubCategories  		[]SubCategory 	`gorm:"constraint:OnDelete:CASCADE; foreignKey:CategoryID"` 
}

type SubCategory struct{
	SubCategoryID 		uint 				`gorm:"primarykey;autoIncrement"`
	SubCategoryName 	string 				`gorm:"not null"`
	CategoryID 			uint  				`gorm:"not null;index"`
	IsBlocked			bool				`gorm:"default:false"`
	CategoryDiscount 	uint 				
	Deleted_at 			gorm.DeletedAt 		`gorm:"index"`

	// Associations
	Products			[]Product 			`gorm:"foreignkey:SubCategoryID"`
	Category 			Category 			`gorm:"constraint:OnDelete:CASCADE;"`
}

type Product struct {
	ProductID				uint				`gorm:"primarykey;autoIncrement"`
	ProductName				string				`gorm:"not null" json:"name"`
	Description				string				`json:"description"`
	SubCategoryID			uint				`gorm:"not null"`
	CreatedAt 				time.Time
	DeletedAt				gorm.DeletedAt 		`gorm:"index"`

	// Associations
	SubCategory				SubCategory			`gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Product_variants 		[]Product_Variant 	`gorm:"constraint:OnDelete:CASCADE;foreignkey:ProductID"`
	Reviews          		[]Review          	`gorm:"constraint:OnDelete:CASCADE;foreignkey:ProductID"`
}


type Product_Variant struct {
	ID					uint				`gorm:"primarykey;autoIncrement"`
	ProductID			uint				`gorm:"index"`
	Variant_name		string				`gorm:"not null"`
	Size				string
	Stock				int
	Price				float64				`gorm:"not null"`
	Tax					float64				`gorm:"not null"`
	DiscountedPrice		float64
	CreatedAt			time.Time
	UpdatedAt			time.Time
	IsActive			bool      			`gorm:"default:true"`
	DeletedAt			gorm.DeletedAt 		`gorm:"index"`

	// Associations
	Product 			Product				`gorm:"constraint: OnDelete:CASCADE"`
	Product_images		[]Product_image		`gorm:"constraint:OnDelete:CASCADE;foreignkey:ProductVariantID"`
	CartItems			[]CartItem			`gorm:"constraint:OnDelelte:CASCADE;foreignkey:ProductID"`
	WishLists			[]WishList			`gorm:"constraint:OnDelelte:CASCADE;foreignkey:ProductID"`
}

type Product_image struct {
	ProductImageID					uint				`gorm:"primarykey;autoIncrement"`
	ProductVariantID				uint				`gorm:"index"`
	Image_url						string				`gorm:"index"`
	Is_primary						bool
	Order_no						int
	CreatedAt						time.Time
	DeleteAt						gorm.DeletedAt		`gorm:"index"`

	// Associations
	Product_Variant					Product_Variant		`gorm:"constraint: OnDelete:CASCADE"`
}

type Review struct {
    ID        				uint      			`gorm:"primarykey;autoIncrement"`
    ProductID 				uint				`gorm:"not null;index"`
    UserID    				uint				`gorm:"not null;index"`
	OrderID 				uint
    Rating    				int       
    Comment   				string
    CreatedAt 				time.Time

	// Associations
	Product    				Product   			`gorm:"constraint: OnDelete:CASCADE"`
	User      				User      			`gorm:"constraint:ONDELETE:CASCADE; foreignkey:UserID"`
	Order 					Order				`gorm:"constraint:ONDELETE:CASCADE"`
}

type Banner struct {
	ID				uint		`gorm:"primarykey;autoIncrement"`
	Title			string
	ImageUrl		string		`gorm:"not null"`
	RedirectURL		string
	Active			bool
	CreatedAt		time.Time
	UpdateAt		time.Time 
	DeletedAt 		gorm.DeletedAt
}

type OTPVerification struct {
	ID 					uint		`gorm:"primarykey;autoIncrement"`
	Email				string		`gorm:"not null"`
	OTP 				string		`gorm:"not null"`
	ExpiresAt			time.Time	`gorm:"not null"`
	IsUsed				bool		`gorm:"default:false"`		
}

type ProductOffer struct {
	ID 					uint		`gorm:"primarykey;autoIncrement"`
	ProductID			uint		`gorm:"not null;index"`
	OfferName			string		`gorm:"not null"`
	DiscountPercentage	float64		`gorm:"not null"`
	CreatedAt			time.Time
	EndAt				time.Time
	Active				bool		`gorm:"default:true"`

	// Associations
	Product				Product		`gorm:"constraint:OnDelete:CASCADE;"`
}

type CategoryOffer struct {
	ID 					uint		`gorm:"primarykey;autoIncrement"`
	CategoryID			uint		`gorm:"not null;index"`
	CategorryName 		string		`gorm:"not null"`
	OfferName			string		`gorm:"not null"`
	DiscountPercentage	float64		`gorm:"not null"`
	CreatedAt			time.Time
	EndAt				time.Time
	Active				bool		`gorm:"default:true"`

}