package helper

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"math"

	"gorm.io/gorm"
)

func ItemReturnOnline(orderId, itemId, reason string) error{
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
	itemTotal := orderItem.Price * float64(orderItem.Quantity) + ptax * float64(orderItem.Quantity) - orderItem.Discount
	orignalTotal := 0.0


	// fetching orignal amount
	for _, item := range order.OrderItems{
		var tempP models.Product_Variant
		db.Db.Where("id = ?",item.ProductID).First(&tempP)
		tempTax := tempP.Tax * float64(item.Quantity)
		orignalTotal += item.Price * float64(item.Quantity) + tempTax - item.Discount
	}

	if usedCoupon.ID != 0 {

		var coupon models.Coupons
		db.Db.Where("id = ?",usedCoupon.CouponID).First(&coupon)

		adjustedTotal := order.TotalAmount - itemTotal

	
		// less than minmum amount in coupon
		if adjustedTotal < coupon.MinAmount && adjustedTotal > 0{

			newTotal := orignalTotal - itemTotal
			refundAmount := order.TotalAmount - newTotal

			if refundAmount < 0{
				refundAmount = 0
			}

			if refundAmount < 1{
				desc := newTotal - order.TotalAmount
				order.DiscountTotal = desc 
			}else{
				
				order.DiscountTotal = 0.0
				db.Db.Delete(&usedCoupon)
			}
			
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

			if WalletTransaction.ID != 0 {
				db.Db.Delete(&WalletTransaction)
			}


			err := db.Db.Create(&walletTranscation).Error
			if err != nil{
				return err 
			}
	

		}else if adjustedTotal < coupon.MinAmount && adjustedTotal < 0{

			var refundAmount float64
			var newTotal float64

			if CheckLeftItems(order.ID){
				refundAmount = order.TotalAmount
			}else{
				newTotal = orignalTotal - itemTotal
				refundAmount = order.TotalAmount - newTotal
			}

			if refundAmount < 0 {
				refundAmount = 0
			}
			
			// amount refund
			walletTransaction := models.WalletTransaction{
				UserID: order.UserID,
				OrderID: order.ID,
				OrderItemID: orderItem.ID,
				Amount: refundAmount,
				Type: "Credit",
				Description: reason,
				RefundStatus: true,
			}
			err := db.Db.Create(&walletTransaction).Error
			if err != nil{
				return err 
			}

			if refundAmount < 1{
				desc := newTotal - order.TotalAmount
				order.DiscountTotal = desc 
			}else{
				order.DiscountTotal = 0
			}

		}else{

			returnamount := itemTotal
			newTotal := orignalTotal - itemTotal

			if adjustedTotal < newTotal{
				returnamount = order.TotalAmount - newTotal
				if returnamount < 0 {
					returnamount = 0
				}

				if order.TotalAmount < newTotal {
					order.DiscountTotal = newTotal - order.TotalAmount
				}else{
					order.DiscountTotal = 0.0
					db.Db.Delete(&usedCoupon)
				}
				
			}

			if returnamount > order.TotalAmount{
				returnamount = 0
			}

			// amount refund
			walletTransaction := models.WalletTransaction{
				UserID: order.UserID,
				OrderID: order.ID,
				OrderItemID: orderItem.ID,
				Amount: returnamount,
				Type: "Credit",
				Description: reason,
				RefundStatus: true,
			}

			err := db.Db.Create(&walletTransaction).Error
			if err != nil{
				return err 
			}

		}

		//db.Db.Delete(&usedCoupon)

	}else{

		var returnamount float64

		if itemTotal > order.TotalAmount{
			returnamount = order.TotalAmount
		}else{
			returnamount = itemTotal
		}

		newalletTransaction := models.WalletTransaction{
			UserID: order.UserID,
			OrderID: order.ID,
			OrderItemID: orderItem.ID,
			Amount: returnamount,
			Type: "Credit",
			Description: reason,
			RefundStatus: true,
		}

		err := db.Db.Create(&newalletTransaction).Error
		if err != nil{
			return err 
		}

	}

	// update item status
	orderItem.Status = "Return requested"
	orderItem.Reason = reason

	db.Db.Save(&order)
	db.Db.Save(&orderItem)

	if order.DiscountTotal < 1{
		if err := OrderDiscountFix(order.ID); err != nil{
			return err 
		}
	}

	return nil
}

func ItemReturnCod(orderId, itemId, reason string) error{
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


	if err := db.Db.Where("id = ?",orderItem.ProductID).First(&Product).Error; err != nil{
		return err 
	}

	ptax := Product.Tax
	itemTotal := orderItem.Price * float64(orderItem.Quantity) + ptax * float64(orderItem.Quantity) - orderItem.Discount
	orignalTotal := 0.0

	// fetching orignal amount
	for _, item := range order.OrderItems{
		var tempP models.Product_Variant
		db.Db.Where("id = ?",item.ProductID).First(&tempP)
		tempTax := tempP.Tax * float64(item.Quantity)
		orignalTotal += item.Price * float64(item.Quantity) + tempTax - item.Discount
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

				if refundAmount < 0{
					refundAmount = 0
				}

				if refundAmount < 1{
					desc := newTotal - order.TotalAmount
					order.DiscountTotal = desc 
				}else{
					
					order.DiscountTotal = 0.0
					db.Db.Delete(&usedCoupon)
				}

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

				returnamount := itemTotal
				newTotal := orignalTotal - itemTotal

				if adjustedTotal < newTotal{
				returnamount = order.TotalAmount - newTotal
					if returnamount < 0 {
						returnamount = 0
					}

					if order.TotalAmount < newTotal {
						order.DiscountTotal = newTotal - order.TotalAmount
					}else{
						order.DiscountTotal = 0.0
						db.Db.Delete(&usedCoupon)
					}
					
				}

				if returnamount > order.TotalAmount{
					returnamount = 0
				}

				walletTransaction := models.WalletTransaction{
					UserID: order.UserID,
					OrderID: order.ID,
					OrderItemID: orderItem.ID,
					Amount: returnamount,
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

			var returnamount float64

			if itemTotal > order.TotalAmount{
				returnamount = order.TotalAmount
			}else{
				returnamount = itemTotal
			}

			walletTransaction := models.WalletTransaction{
				UserID: order.UserID,
				OrderID: order.ID,
				OrderItemID: orderItem.ID,
				Amount: returnamount,
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


			if adjustedTotal < coupon.MinAmount{
				
				orderItem.Status = "Return requested"
				orderItem.Reason = reason

			}else{

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

