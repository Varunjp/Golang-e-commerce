package responsemodels

type User struct {
	ID       uint   `gorm:"primarykey"`
	Username string `gorm:"not null"`
	Email    string `gorm:"not null; unique; index"`
	Phone    string `gorm:"not null; unique"`
	Status   string `gorm:"check(status IN('Active', 'Inactive', 'Blocked'))"`
}