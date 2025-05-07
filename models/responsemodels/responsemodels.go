package responsemodels

import "time"

type User struct {
	ID       uint   `gorm:"primarykey"`
	Username string `gorm:"not null"`
	Email    string `gorm:"not null; unique; index"`
	Phone    string `gorm:"not null; unique"`
	Status   string `gorm:"check(status IN('Active', 'Inactive', 'Blocked'))"`
}

type Products struct {
	ID        uint
	Name      string
	Size 	  string
	ImageURl  string 
	Price     float64
	Quantity  int
	CreatedAt time.Time
	InStock	  bool
}

type HomePage struct {
	ID			uint
	ImageURL	string
	Name		string
	Rating		int
	OldPrice	int
	Price		int
	Discount	int
}