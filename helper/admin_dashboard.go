package helper

import (
	db "first-project/DB"
	"first-project/models"
)

type Product struct {
	Name      string
	TotalSold int
}

type Category struct {
	CategoryName 	string 
	Name      		string
	TotalSold 		int
}

func TopProductCategory() ([]Product, []Category) {
	var products []Product
	var categories []Category

	db.Db.Table("order_items").
		Select("products.product_name as name, SUM(order_items.quantity) as total_sold").
		Joins("JOIN product_variants ON product_variants.id = order_items.product_id").
		Joins("JOIN products ON products.product_id = product_variants.product_id").
		Where("order_items.status = ?", "Delivered").
		Group("product_variants.id,products.product_name").
		Order("total_sold DESC").
		Limit(10).
		Scan(&products)
	
	db.Db.Table("order_items").
    Select("categories.category_name, sub_categories.sub_category_name as name, SUM(order_items.quantity) as total_sold").
    Joins("JOIN product_variants ON product_variants.id = order_items.product_id").
	Joins("JOIN products ON products.product_id = product_variants.product_id").
    Joins("JOIN sub_categories ON sub_categories.sub_category_id = products.sub_category_id").
	Joins("JOIN categories ON sub_categories.category_id = categories.category_id").
    Where("order_items.status = ?", "Delivered").
    Group("sub_categories.sub_category_id,categories.category_name").
    Order("total_sold DESC").
    Limit(10).
    Scan(&categories)

	return products,categories
}

func SalesReport() float64{
	var totalorders []models.Order

	dbFullordes := db.Db.Model(&models.Order{})

	dbFullordes.Preload("OrderItems").Find(&totalorders)
	totalSales := 0.0
	for _,item := range totalorders{
		if item.Status == "Delivered"{
			totalSales += item.TotalAmount
		}
	}

	return totalSales
}