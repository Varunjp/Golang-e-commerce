package helper

import (
	db "first-project/DB"
	"first-project/models"
	"strconv"

	"gorm.io/gorm"
)

func CancelOrderForProduct(productId string) error{
	var Product models.Product_Variant
	var Orders []models.Order

	if err := db.Db.Where("id = ?",productId).First(&Product).Error; err != nil{
		return err 
	}

	if err := db.Db.Preload("OrderItems").Where("status = ?","Processing").Find(&Orders).Error; err != nil{
		if err != gorm.ErrRecordNotFound{
			return err 
		}
	}

	for _,order := range Orders{
		itemCount := 0
		for _,item := range order.OrderItems{
			if item.ProductID == Product.ID && item.Status == "Processing"{
				orderId := strconv.FormatUint(uint64(order.ID),10)
				itemId := strconv.FormatUint(uint64(item.ID),10)
				reason := "Product not available"
				if order.PaymentMethod != "cod"{
					err := ItemCancelOnline(orderId,itemId,reason)
					if err != nil{
						return err
					}
				}else{
					err := ItemCancelCod(orderId,itemId,reason)
					if err != nil{
						return err
					}
				}
				item.Status = "Cancelled"
				itemCount++
			}
			if len(order.OrderItems) == itemCount{
				order.Status = "Cancelled"
				if order.PaymentMethod != "cod" {
					order.PaymentStatus = "Refunded"
				}else{
					order.PaymentMethod = "Not valid"
				}
			}
			db.Db.Save(&item)
		}
		db.Db.Save(order)
	}

	return nil 
}