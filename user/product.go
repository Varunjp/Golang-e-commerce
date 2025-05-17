package user

import (
	db "first-project/DB"
	"first-project/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Product(c *gin.Context){

	productID := c.Param("id")

	var product models.Product
	var product_variant models.Product_Variant
	var images []models.Product_image

	if err := db.Db.Where("deleted_at IS NULL").First(&product_variant,productID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product not found"})
		return 
	}

	if err := db.Db.Where("deleted_at IS NULL AND product_id = ?",product_variant.ProductID).First(&product).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product not found"})
		return 
	}

	if err := db.Db.Where("product_variant_id = ?",productID).Order("order_no ASC").Find(&images).Error; err != nil{
		log.Println("No images found :",err.Error())
	}

	c.HTML(http.StatusOK,"product.html",gin.H{
		"user":"done",
		"pagetitle":product_variant.Variant_name,
		"Product": product,
		"variant": product_variant,
		"Images": images,
	})

}