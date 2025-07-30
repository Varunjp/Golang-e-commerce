package helper

import (
	db "first-project/DB"
	"first-project/models"
)

func CheckLeftItems(orderID uint) bool {

	var order models.Order

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderID).First(&order).Error; err != nil{
		return false 
	}

	checkCount := 0

	for _,item := range order.OrderItems{
		if item.Status == "Pending" || item.Status == "Processing"{
			checkCount++
		}
	}

	if checkCount == 1{
		return true
	}else{
		return false 
	}
}