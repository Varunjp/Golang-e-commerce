package helper

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"math"

	"gorm.io/gorm"
)

func ItemCancelOnline(orderId, itemId, reason string) error{
	var order models.Order
	var orderItem models.OrderItem
	var usedCoupon models.UsedCoupon
	var Product models.Product_Variant
	var WalletTransaction models.WalletTransaction

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

	if err := db.Db.Where("order_id = ? AND user_id = ? AND type = ?",orderId,order.UserID,"Debit").First(&WalletTransaction).Error; err != nil{
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

	walletAmount := 0.0


	if WalletTransaction.ID != 0 {
		walletAmount = math.Abs(WalletTransaction.Amount)
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
				Amount: refundAmount+walletAmount,
				Type: "Credit",
				Description: reason,
				RefundStatus: true,
			}

			if WalletTransaction.ID != 0 {
				db.Db.Delete(&WalletTransaction)
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
				Amount: itemTotal+walletAmount,
				Type: "Credit",
				Description: reason,
				RefundStatus: true,
			}
			err := db.Db.Create(&walletTransaction).Error
			if err != nil{
				return err 
			}

			if WalletTransaction.ID != 0 {
				db.Db.Delete(&WalletTransaction)
			}

		}

		//db.Db.Delete(&usedCoupon)

	}else{

		walletTransaction := models.WalletTransaction{
			UserID: order.UserID,
			OrderID: order.ID,
			OrderItemID: orderItem.ID,
			Amount: itemTotal+walletAmount,
			Type: "Credit",
			Description: reason,
			RefundStatus: true,
		}

		if WalletTransaction.ID != 0 {
			db.Db.Delete(&WalletTransaction)
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
	var WalletTransaction models.WalletTransaction

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

	if err := db.Db.Where("order_id = ? AND user_id = ? AND type = ?",orderId,order.UserID,"Debit").First(&WalletTransaction).Error; err != nil{
		if err != gorm.ErrRecordNotFound {
			return err 
		}
	}

	walletAmount := 0.0

	if WalletTransaction.ID != 0 {
		walletAmount = math.Abs(WalletTransaction.Amount)
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
					Amount: refundAmount+walletAmount,
					Type: "Credit",
					Description: reason,
					RefundStatus: true,
				}

				if WalletTransaction.ID != 0 {
				db.Db.Delete(&WalletTransaction)
				}

				err := db.Db.Create(&walletTranscation).Error
				if err != nil{
					return err 
				}
	
				orderItem.Status = "Return requested"
				orderItem.Reason = reason

			}else{

				walletTransaction := models.WalletTransaction{
					UserID: order.UserID,
					OrderID: order.ID,
					OrderItemID: orderItem.ID,
					Amount: itemTotal+walletAmount,
					Type: "Credit",
					Description: reason,
					RefundStatus: true,
				}
				err := db.Db.Create(&walletTransaction).Error
				if err != nil{
					return err 
				}

				if WalletTransaction.ID != 0 {
				db.Db.Delete(&WalletTransaction)
				}

				orderItem.Status = "Return requested"
				orderItem.Reason = reason
			}

			//db.Db.Delete(&usedCoupon)

		}else{

			walletTransaction := models.WalletTransaction{
				UserID: order.UserID,
				OrderID: order.ID,
				OrderItemID: orderItem.ID,
				Amount: itemTotal+walletAmount,
				Type: "Credit",
				Description: reason,
				RefundStatus: true,
			}



			err := db.Db.Create(&walletTransaction).Error
			if err != nil{
				return err 
			}
			if WalletTransaction.ID != 0 {
				db.Db.Delete(&WalletTransaction)
			}
			

		}

	}else{

		ItemCheck := 0

		for _, item := range order.OrderItems {
			if item.Status == "Delivered" || item.Status == "Processing"{
				ItemCheck ++
			}
		}

		if usedCoupon.ID != 0 {

			var coupon models.Coupons
			db.Db.Where("id = ?",usedCoupon.CouponID).First(&coupon)

			adjustedTotal := order.TotalAmount - itemTotal

			// less than minmum amount in coupon
			if adjustedTotal < coupon.MinAmount{
				
				orderItem.Status = "Return requested"
				orderItem.Reason = reason

			}else{
				//newTotal := order.TotalAmount - itemTotal

				// update order amount
				// order.TotalAmount = newTotal
				// order.SubTotal = newTotal
				// remove item from orderitem
				// db.Db.Model(&models.Product_Variant{}).Where("id = ?",orderItem.ProductID).Update("stock",gorm.Expr("stock + ?",orderItem.Quantity))
				orderItem.Status = "Return requested"
				orderItem.Reason = reason
			}

			if (ItemCheck == 0 && len(order.OrderItems) > 1) || (ItemCheck == 1 && len(order.OrderItems) == 1){
				if WalletTransaction.ID != 0{
					newTransaction := models.WalletTransaction{
						UserID: order.UserID,
						OrderID: order.ID,
						Amount: math.Abs(WalletTransaction.Amount),
						OrderItemID: orderItem.ID,
						Type: "Credit",
						Description: fmt.Sprintf("Refund for order : %d",order.ID),
						RefundStatus: true,
					}

					db.Db.Create(&newTransaction)
					db.Db.Delete(&WalletTransaction)
				}
			}


		}else if WalletTransaction.ID != 0{

			if (ItemCheck == 0 && len(order.OrderItems) > 1) || (ItemCheck == 1 && len(order.OrderItems) == 1){
				
				newTransaction := models.WalletTransaction{
					UserID: order.UserID,
					OrderID: order.ID,
					Amount: math.Abs(WalletTransaction.Amount),
					OrderItemID: orderItem.ID,
					Type: "Credit",
					Description: fmt.Sprintf("Refund for order : %d",order.ID),
					RefundStatus: true,
				}

				db.Db.Create(&newTransaction)
				db.Db.Delete(&WalletTransaction)

			}
			orderItem.Status = "Return requested"
			orderItem.Reason = reason

			
		}else{
			
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
	var walletTransaction models.WalletTransaction

	if err := db.Db.Preload("OrderItems").Where("id = ?",orderId).First(&order).Error; err != nil{
		return err 
	}

	if err := db.Db.Where("order_id = ? AND user_id = ? AND type = ?",order.ID,order.UserID,"Debit").First(&walletTransaction).Error; err != nil{
		if err != gorm.ErrRecordNotFound{
			return err
		}
	}


	if order.PaymentMethod != "cod" {
		totalAmount := order.TotalAmount
		if walletTransaction.ID != 0{
			totalAmount += math.Abs(walletTransaction.Amount)
		}
		err := CreditWallet(order.UserID,order.TotalAmount,"Order cancelled")
		if err != nil{
			return err 
		}
	}else if walletTransaction.ID != 0 {
		walletAmount := math.Abs(walletTransaction.Amount)
		err := CreditWallet(order.UserID,walletAmount,"Order cancelled")
		if err != nil{
			return err 
		}
	}

	for _, item := range order.OrderItems{

		db.Db.Model(&models.Product_Variant{}).Where("id = ?",item.ProductID).Update("stock",gorm.Expr("stock + ?",item.Quantity))
		item.Status = "Cancelled"
		db.Db.Save(&item)
	}

	return nil 
}