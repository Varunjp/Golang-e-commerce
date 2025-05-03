package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct{
	ID 			uint   		`gorm:"primarykey"`
	Username 	string 		`gorm:"not null"`
	Email 		string 		`gorm:"not null; unique; index"`
	Password 	string 		`gorm:"not null"`
	Phone 		string 		`gorm:"not null; unique"`
	Status 		string 		`gorm:"check(status IN('Active', 'Inactive', 'Blocked'))"`
	Created_at 	time.Time
	Updated_at 	time.Time
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
	CategoryID uint `gorm:"primarykey" json:"category_id"`
	CategoryName string `json:"name"`
	CreateAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}