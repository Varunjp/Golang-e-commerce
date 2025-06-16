package helper

import (
	db "first-project/DB"
	"first-project/models"

	"gorm.io/gorm"
)

func ItemCancelOnline(orderId, itemId, reason string) error{
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
			refundAmount := order.TotalAmount - newTotal

			// amount refunded
			walletTranscation := models.WalletTransaction{
				UserID: order.UserID,
				OrderID: order.ID,
				OrderItemID: orderItem.ID,
				Amount: refundAmount,
				Type: "Credit",
				Description: reason,
				RefundStatus: true,
			}

			err := db.Db.Create(&walletTranscation).Error
			if err != nil{
				return err 
			}
	

		}else{
			
			// amount refund
			walletTransaction := models.WalletTransaction{
				UserID: order.UserID,
				OrderID: order.ID,
				OrderItemID: orderItem.ID,
				Amount: itemTotal,
				Type: "Credit",
				Description: reason,
				RefundStatus: true,
			}
			err := db.Db.Create(&walletTransaction).Error
			if err != nil{
				return err 
			}

		}

	}else{

		walletTransaction := models.WalletTransaction{
			UserID: order.UserID,
			OrderID: order.ID,
			OrderItemID: orderItem.ID,
			Amount: itemTotal,
			Type: "Credit",
			Description: reason,
			RefundStatus: true,
		}



		err := db.Db.Create(&walletTransaction).Error
		if err != nil{
			return err 
		}

	}

	// update item status
	orderItem.Status = "Return requested"
	orderItem.Reason = reason

	db.Db.Save(&order)
	db.Db.Save(&orderItem)

	return nil
}

func ItemCancelCod(orderId, itemId, reason string) error{
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

	if order.Status == "Delivered" {

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
					OrderItemID: orderItem.ID,
					Amount: refundAmount,
					Type: "Credit",
					Description: reason,
					RefundStatus: true,
				}

				err := db.Db.Create(&walletTranscation).Error
				if err != nil{
					return err 
				}
	

				// // update order status
				// order.TotalAmount = newTotal
				// order.DiscountTotal = 0.0

				// // remove item from orderitem
				// db.Db.Model(&models.Product_Variant{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))

				orderItem.Status = "Return requested"
				orderItem.Reason = reason

			}else{

				walletTransaction := models.WalletTransaction{
					UserID: order.UserID,
					OrderID: order.ID,
					OrderItemID: orderItem.ID,
					Amount: itemTotal,
					Type: "Credit",
					Description: reason,
					RefundStatus: true,
				}
				err := db.Db.Create(&walletTransaction).Error
				if err != nil{
					return err 
				}


				// update order amount
				// order.TotalAmount = newTotal
				// // remove item from orderitem
				// db.Db.Model(&models.Product_Variant{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
				orderItem.Status = "Return requested"
				orderItem.Reason = reason
			}

		}else{

			walletTransaction := models.WalletTransaction{
				UserID: order.UserID,
				OrderID: order.ID,
				OrderItemID: orderItem.ID,
				Amount: itemTotal,
				Type: "Credit",
				Description: reason,
				RefundStatus: true,
			}



			err := db.Db.Create(&walletTransaction).Error
			if err != nil{
				return err 
			}

			// newTotal := order.TotalAmount - itemTotal

			// // update order amount
			// order.TotalAmount = newTotal
			// // remove item from orderitem
			// db.Db.Model(&models.Product_Variant{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
			// orderItem.Status = "Return requested"
			// orderItem.Reason = reason


		}

	}else{

		if usedCoupon.ID != 0 {

			var coupon models.Coupons
			db.Db.Where("id = ?",usedCoupon.CouponID).First(&coupon)

			adjustedTotal := order.TotalAmount - itemTotal
			// less than minmum amount in coupon
			if adjustedTotal < coupon.MinAmount{
				newTotal := orignalTotal - itemTotal
				// update order status
				order.TotalAmount = newTotal
				order.SubTotal = newTotal
				order.DiscountTotal = 0.0

				// remove item from orderitem
				// db.Db.Model(&models.Product_Variant{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
				orderItem.Status = "Return requested"
				orderItem.Reason = reason

			}else{
				newTotal := order.TotalAmount - itemTotal

				// update order amount
				order.TotalAmount = newTotal
				order.SubTotal = newTotal
				// remove item from orderitem
				// db.Db.Model(&models.Product_Variant{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
				orderItem.Status = "Return requested"
				orderItem.Reason = reason
			}
		}else{

			newTotal := order.TotalAmount - itemTotal

			// update order amount
			order.TotalAmount = newTotal
			order.SubTotal = newTotal
			// remove item from orderitem
			//db.Db.Model(&models.Product_Variant{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
			orderItem.Status = "Return requested"
			orderItem.Reason = reason

		}

	}

	db.Db.Save(&order)
	db.Db.Save(&orderItem)

	return nil 
}

func AdminOrderCancel(orderId uint)error{
	var order models.Order

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderId).First(&order).Error; err != nil{
		return err 
	}

	if order.PaymentMethod != "cod" {
		err := CreditWallet(order.UserID,order.TotalAmount,"Order canceled")
		if err != nil{
			return err 
		}
	}

	for _, item := range order.OrderItems{

		db.Db.Model(&models.Product_Variant{}).Where("id = ?",item.ProductID).Update("stoke",gorm.Expr("stoke + ?",item.Quantity))
		item.Status = "Canceled"
		db.Db.Save(&item)
	}

	return nil 
}