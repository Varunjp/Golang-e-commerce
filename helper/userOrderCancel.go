package helper

import (
	db "first-project/DB"
	"first-project/models"
	"log"
	"math"

	"gorm.io/gorm"
)

func UserOrderCancelItem(itemId string) error {

	var orderItem models.OrderItem
	var order models.Order

	if err := db.Db.Where("id = ?", itemId).First(&orderItem).Error; err != nil {
		return err
	}

	if err := db.Db.Preload("OrderItems").Where("id = ?", orderItem.OrderID).First(&order).Error; err != nil {
		return err 
	}

	if err := db.Db.Model(&models.Product_Variant{}).Where("id = ?", orderItem.ProductID).Update("stock", gorm.Expr("stock + ?", orderItem.Quantity)).Error; err != nil {
		return err
	}

	checkRemaing := 0

	for _, item := range order.OrderItems {
		if item.Status == "Pending" || item.Status == "Processing" || item.Status == "Delivered" || item.Status == "Delivered non returnable" {
			checkRemaing++
		}
	}

	var product models.Product_Variant
	db.Db.Where("id = ?", orderItem.ProductID).First(&product)

	retrunAmount := orderItem.Price*float64(orderItem.Quantity) + product.Tax*float64(orderItem.Quantity)

	newTotal := order.SubTotal - retrunAmount

	valueCheck, usedCouponId, errVal := GetOrderValue(order.ID, order.UserID, newTotal)
	var walletTransaction models.WalletTransaction
	db.Db.Unscoped().Where("order_id = ? AND user_id = ? AND type = ?", order.ID, order.UserID, "Debit").First(&walletTransaction)

	if valueCheck && errVal == nil {

		order.SubTotal = order.SubTotal - retrunAmount
		order.TotalAmount = order.TotalAmount - retrunAmount
	} else if !valueCheck && errVal == nil {

		if walletTransaction.ID != 0 {

			order.SubTotal = order.SubTotal - retrunAmount
			updateTotal := order.SubTotal + walletTransaction.Amount

			if updateTotal <= 0 {
				order.TotalAmount = 0
				order.DiscountTotal = 0
			} else {
				order.TotalAmount = updateTotal
				order.DiscountTotal = math.Abs(walletTransaction.Amount)
			}

		} else {

			order.SubTotal = order.SubTotal - retrunAmount
			order.TotalAmount = order.SubTotal
			order.DiscountTotal = 0.0
		}

		if usedCouponId != 0 {
			db.Db.Delete(&models.UsedCoupon{}, usedCouponId)
		}

	} else if errVal != nil {

		log.Println(errVal)
	}

	if checkRemaing == 0 {

		if order.PaymentMethod != "cod" || order.Status == "Delivered" || order.PaymentStatus == "Refund is being processed" || order.Status == "Returned" {
			order.PaymentStatus = "Refunded"
		} else {
			order.PaymentStatus = "Failed"
		}
		order.Status = "Returned"
		order.Reason = orderItem.Reason
	}

	orderItem.Status = "Returned"
	db.Db.Save(&order)
	db.Db.Save(&orderItem)
	db.Db.Delete(&orderItem)

	return nil 
}