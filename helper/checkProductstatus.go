package helper

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
)

func CheckProduct(orderid uint) error {
	var order models.Order
	var orderItems []models.OrderItem
	
	if err := db.Db.Preload("OrderItems").Where("id = ?",orderid).First(&order).Error; err != nil{
		return err 
	}

	if err := db.Db.Where("user_id = ? AND status = ?",order.UserID,"Delivered").Find(&orderItems).Error; err != nil{
		return err 
	}

	for _,item := range order.OrderItems{
		itemcount := 0
		for _,delItem := range orderItems {
			if item.ProductID == delItem.ProductID{
				itemcount += delItem.Quantity
			}
		}

		if item.Quantity+itemcount > 5{
			return fmt.Errorf("item exceeded product purchase limit")
		}

		var product models.Product_Variant
		db.Db.Where("id = ?",item.ProductID).First(&product)

		if product.Stock < item.Quantity {
			return  fmt.Errorf("item out of stock")
		}
	}

	return nil 
}