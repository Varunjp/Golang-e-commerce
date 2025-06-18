package helper

import (
	db "first-project/DB"
	"first-project/models"

	"gorm.io/gorm"
)

func GetOrderValue(orderId, userId uint, returnamount float64) (bool,int,error) {
	var usedCoupon models.UsedCoupon
	var coupon models.Coupons

	if err := db.Db.Where("order_id = ? AND user_id = ?",orderId,userId).First(&usedCoupon).Error; err != nil{
		if err != gorm.ErrRecordNotFound{
			return false,0,err
		}
		
	}

	if err := db.Db.Where("id = ?",usedCoupon.CouponID).First(&coupon).Error; err != nil{
		
		if err != gorm.ErrRecordNotFound{
			return false,0,err
		}
		 
	}

	if coupon.ID != 0 && coupon.MinAmount <= returnamount {
		return true,int(usedCoupon.ID),nil
	}else {
		return false,int(usedCoupon.ID),nil
	}
}