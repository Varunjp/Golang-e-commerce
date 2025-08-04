package helper

import (
	db "first-project/DB"
	"first-project/models"
)

func OrderDiscountFix(orderId uint) error {
	var order models.Order

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderId).First(&order).Error; err != nil{
		return err 
	}

	checkDiscount := 0.0

	for _,item := range order.OrderItems{
		if item.Status == "Delivered" || item.Status == "Processing" || item.Status == ""{
			checkDiscount += item.Discount
		} 
	}

	order.DiscountTotal = order.DiscountTotal + checkDiscount

	if checkDiscount > 0{
		order.TotalAmount = order.SubTotal - checkDiscount
	}

	if err := db.Db.Save(&order).Error; err != nil{
		return err 
	}

	return nil 
}