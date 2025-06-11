package helper

import (
	db "first-project/DB"
	"first-project/models"

	"gorm.io/gorm"
)

func ItemCancelOnline(orderId, itemId string) error{
	var order models.Order
	var orderItem models.OrderItem
	var usedCoupon models.UsedCoupon
	var Product models.Product_Variant

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderId).First(&order).Error; err != nil{
		return err 
	}

	if err := db.Db.Where("id = ?",itemId).First(&orderItem).Error; err != nil{
		return err 
	}

	if err := db.Db.Where("order_id = ?",orderId).First(&usedCoupon).Error; err != nil{
		return err
	}

	if err := db.Db.Where("id = ?",orderItem.ProductID).First(&Product).Error; err != nil{
		return err 
	}

	ptax := Product.Tax
	itemTotal := orderItem.Price * float64(orderItem.Quantity) + ptax * float64(orderItem.Quantity)
	orignalTotal := 0.0


	// fetching orignal amount
	for _, item := range order.OrderItems{
		var tempP models.Product_Variant
		db.Db.Where("id = ?",item.ProductID).First(&tempP)
		tempTax := tempP.Tax * float64(item.Quantity)
		orignalTotal += item.Price * float64(item.Quantity) + tempTax
	}


	if usedCoupon.ID != 0 {

		var coupon models.Coupons
		db.Db.Where("id = ?",usedCoupon.CouponID).First(&coupon)

		adjustedTotal := order.TotalAmount - itemTotal

		// less than minmum amount in coupon
		if adjustedTotal < coupon.MinAmount{
			newTotal := orignalTotal - itemTotal
			refundAmount := order.TotalAmount - newTotal

			// amount refunded
			walletTranscation := models.WalletTransaction{
				UserID: order.UserID,
				OrderID: order.ID,
				Amount: refundAmount,
				Type: "Credit",
				Description: "Item cancel",
				RefundStatus: true,
			}

			err := db.Db.Create(&walletTranscation).Error
			if err != nil{
				return err 
			}
			// update order status
			order.TotalAmount = newTotal
			order.DiscountTotal = 0.0
			order.SubTotal = newTotal
			db.Db.Save(&order)
			// remove item from orderitem
			db.Db.Model(&models.WalletTransaction{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
			orderItem.Status = "Returned"
			db.Db.Save(&orderItem)
			db.Db.Delete(&orderItem)
		}else{
			newTotal := order.TotalAmount - itemTotal
			
			// amount refund
			walletTransaction := models.WalletTransaction{
				UserID: order.UserID,
				OrderID: order.ID,
				Amount: itemTotal,
				Type: "Credit",
				Description: "Item canceled",
				RefundStatus: true,
			}
			err := db.Db.Create(&walletTransaction).Error
			if err != nil{
				return err 
			}
			// update order amount
			order.TotalAmount = newTotal
			order.SubTotal = newTotal
			db.Db.Save(&order)
			// remove item from orderitem
			db.Db.Model(&models.WalletTransaction{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
			orderItem.Status = "Returned"
			db.Db.Save(&orderItem)
			db.Db.Delete(&orderItem)
		}

	}else{

		newTotal := order.TotalAmount - itemTotal
		// amount refund
		walletTransaction := models.WalletTransaction{
			UserID: order.UserID,
			OrderID: order.ID,
			Amount: itemTotal,
			Type: "Credit",
			Description: "Item canceled",
			RefundStatus: true,
		}



		err := db.Db.Create(&walletTransaction).Error
		if err != nil{
			return err 
		}
		// update order amount
		order.TotalAmount = newTotal
		order.SubTotal = newTotal
		db.Db.Save(&order)
		// remove item from orderitem
		db.Db.Model(&models.WalletTransaction{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
		orderItem.Status = "Returned"
		db.Db.Save(&orderItem)
		db.Db.Delete(&orderItem)

	}

	return nil
}

func ItemCancelCod(orderId, itemId string) error{
	var order models.Order
	var orderItem models.OrderItem
	var usedCoupon models.UsedCoupon
	var Product models.Product_Variant

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderId).First(&order).Error; err != nil{
		return err 
	}

	if err := db.Db.Where("id = ?",itemId).First(&orderItem).Error; err != nil{
		return err 
	}

	if err := db.Db.Where("order_id = ?",orderId).First(&usedCoupon).Error; err != nil{
		if err != gorm.ErrRecordNotFound{
			return err
		}
		
	}


	if err := db.Db.Where("id = ?",orderItem.ProductID).First(&Product).Error; err != nil{
		return err 
	}

	ptax := Product.Tax
	itemTotal := orderItem.Price * float64(orderItem.Quantity) + ptax * float64(orderItem.Quantity)
	orignalTotal := 0.0

	// fetching orignal amount
	for _, item := range order.OrderItems{
		var tempP models.Product_Variant
		db.Db.Where("id = ?",item.ProductID).First(&tempP)
		tempTax := tempP.Tax * float64(item.Quantity)
		orignalTotal += item.Price * float64(item.Quantity) + tempTax
	}

	if usedCoupon.ID != 0 {

		var coupon models.Coupons
		db.Db.Where("id = ?",usedCoupon.CouponID).First(&coupon)

		adjustedTotal := order.TotalAmount - itemTotal
		// less than minmum amount in coupon
		if adjustedTotal < coupon.MinAmount{
			newTotal := orignalTotal - itemTotal
			// update order status
			order.TotalAmount = newTotal
			order.DiscountTotal = 0.0
			db.Db.Save(&order)
			// remove item from orderitem
			db.Db.Model(&models.WalletTransaction{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
			orderItem.Status = "Returned"
			db.Db.Save(&orderItem)
			db.Db.Delete(&orderItem)
		}else{
			newTotal := order.TotalAmount - itemTotal

			// update order amount
			order.TotalAmount = newTotal
			db.Db.Save(&order)
			// remove item from orderitem
			db.Db.Model(&models.WalletTransaction{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
			orderItem.Status = "Returned"
			db.Db.Save(&orderItem)
			db.Db.Delete(&orderItem)
		}

	}else{

		newTotal := order.TotalAmount - itemTotal

		// update order amount
		order.TotalAmount = newTotal
		db.Db.Save(&order)
		// remove item from orderitem
		db.Db.Model(&models.WalletTransaction{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
		orderItem.Status = "Returned"
		db.Db.Save(&orderItem)
		db.Db.Delete(&orderItem)

	}
	return nil 
}