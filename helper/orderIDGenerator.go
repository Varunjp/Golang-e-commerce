package helper

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"math/rand"
	"time"
)

func getOrderID() string {
	now := time.Now()
	year := now.Year() % 100
	dayOfYear := now.YearDay()
	random := rand.Intn(100)

	return fmt.Sprintf("ORD%02d%03d%02d",year,dayOfYear,random)[0:9]
}

func GenerateOrderID() string {
	for {
		orderid := getOrderID()
		var count int64
		db.Db.Model(models.Order{}).Where("order_id = ?",orderid).Count(&count)
		if count == 0 {
			return orderid
		}
	}
}