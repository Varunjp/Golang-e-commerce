package user

import (
	db "first-project/DB"
	"first-project/models"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ShowProductList(c *gin.Context){

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit","9")

	page,_ := strconv.Atoi(pageStr)
	limit,_ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit
	
	var total int64
	var products []models.Product_Variant

	db.Db.Model(&models.Product_Variant{}).Count(&total)

	err := db.Db.
    Preload("Product").
    Preload("Product.SubCategory").
    Preload("Product_images", func(db *gorm.DB) *gorm.DB {
        return db.Order("order_no ASC")
    }).
    Offset(offset).
    Limit(limit).
    Find(&products).Error

	
	if err != nil{
		c.HTML(http.StatusInternalServerError, "product_list.html", gin.H{"error": "Error fetching products"})
        return
	}

	formatted := make([]struct{
		ID		uint
		Name	string
		Price	float64
		DiscountedPrice	string 
		ImageURL 	string
	},len(products))

	for i,p := range products{

		formatted[i] = struct{ID uint; Name string; Price float64; DiscountedPrice string; ImageURL string}{
			ID: p.ID,
			Name: p.Variant_name,
			Price: p.Price,
			DiscountedPrice: "",
		}

		if len(p.Product_images) > 0{
			formatted[i].ImageURL = p.Product_images[0].Image_url
		}else{
			formatted[i].ImageURL = ""
		}
	}

	totalPage := int((total + int64(limit) -1 )/ int64(limit))

	session := sessions.Default(c)
	Name,_ := session.Get("name").(string)

	if Name != ""{
		c.HTML(http.StatusOK,"product_list.html",gin.H{
			"user":Name,
			"Products": formatted,
			"Page": page,
			"TotalPages": totalPage,
		})
		return
	}

	c.HTML(http.StatusOK,"product_list.html",gin.H{
		"Products": formatted,
		"Page": page,
		"TotalPages": totalPage,
	})
}