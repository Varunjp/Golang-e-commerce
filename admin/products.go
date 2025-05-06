package admin

import (
	"encoding/base64"
	db "first-project/DB"
	"first-project/models"
	"first-project/models/responsemodels"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ViewProducts(c *gin.Context){

	var dbProducts []struct {
		ProductID int `gorm:"column:id"`
		ProductName string `gorm:"column:variant_name"`
		ProductSize string `gorm:"column:size"`
		ProductImage string	`gorm:"column:image_url"`
		ProductPrice float64 `gorm:"column:price"`
		ProductStock int	`gorm:"column:stock"`
		CreatedAt	time.Time	`gorm:"column:created_at"`
	}

	pageStr := c.DefaultQuery("page","1")
	limitStr := c.DefaultQuery("limit","10")

	page, err := strconv.Atoi(pageStr)

	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1{
		limit = 10
	}

	offset := (page - 1) * limit

	var total int

	keyword := c.Query("search")

	if keyword == ""{
			
		err := db.Db.Table("product_variants").Select("product_variants.id,product_variants.variant_name,product_variants.size,product_images.image_url,product_variants.price,product_variants.stock").Joins("LEFT JOIN product_images ON product_images.product_variant_id = product_variants.id").Where("product_variants.deleted_at IS NULL").Group("product_variants.id,product_variants.variant_name,product_images.image_url").Order("product_variants.id Desc").Offset(offset).Find(&dbProducts).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

			
		//result.Count(&total)
		total = len(dbProducts)

		if len(dbProducts) == 0{
			c.HTML(http.StatusOK,"admin_product_list.html",gin.H{"message":"No products listed"})
			return
		}

		responseProducts := make([]responsemodels.Products,len(dbProducts))

		for i, dbProduct := range dbProducts{
			status := true 
			if dbProduct.ProductStock == 0{
				status = false 
			}
			responseProducts[i] = responsemodels.Products{
				ID: uint(dbProduct.ProductID),
				Name: dbProduct.ProductName,
				Size: dbProduct.ProductSize,
				ImageURl: dbProduct.ProductImage,
				Price: dbProduct.ProductPrice,
				Quantity: dbProduct.ProductStock,
				CreatedAt: dbProduct.CreatedAt,
				InStock: status,
			}
		}

		totalPages := int(math.Ceil(float64(total)/ float64(limit)))

		c.HTML(http.StatusOK,"admin_product_list.html",gin.H{"products":responseProducts,"page":page,
		"limit":limit,
		"totalPages":totalPages,})

	}else{

		var dbProducts []struct {
			ProductID int `gorm:"column:id"`
			ProductName string `gorm:"column:variant_name"`
			ProductSize string `gorm:"column:size"`
			ProductImage string	`gorm:"column:image_url"`
			ProductPrice float64 `gorm:"column:price"`
			ProductStock int	`gorm:"column:stock"`
			CreatedAt	time.Time	`gorm:"column:created_at"`
		}

		err := db.Db.Table("product_variants").Select("product_variants.id,product_variants.variant_name,product_variants.size,product_images.image_url,product_variants.price,product_variants.stock").Joins("LEFT JOIN product_images ON product_images.product_variant_id = product_variants.id").Where("product_variants.variant_name ILIKE ?","%"+keyword+"%").Group("product_variants.id,product_variants.variant_name,product_images.image_url").Order("product_variants.id Desc").Offset(offset).Find(&dbProducts).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		total = len(dbProducts)

		if total == 0 {
			c.HTML(http.StatusOK,"admin_product_list.html",gin.H{"error":"Product not found"})
			return
		}

		responseProducts := make([]responsemodels.Products,len(dbProducts))

		for i, dbProduct := range dbProducts{
			status := true 
			if dbProduct.ProductStock == 0{
				status = false 
			}
			responseProducts[i] = responsemodels.Products{
				ID: uint(dbProduct.ProductID),
				Name: dbProduct.ProductName,
				Size: dbProduct.ProductSize,
				ImageURl: dbProduct.ProductImage,
				Price: dbProduct.ProductPrice,
				Quantity: dbProduct.ProductStock,
				CreatedAt: dbProduct.CreatedAt,
				InStock: status,
			}
		}

		totalPages := int(math.Ceil(float64(total)/ float64(limit)))

		c.HTML(http.StatusOK,"admin_product_list.html",gin.H{"products":responseProducts,"page":page,
		"limit":limit,
		"totalPages":totalPages,})

	}

}

func AddProductPage(c *gin.Context){
	c.HTML(http.StatusOK,"admin_addProduct.html",nil)
}

func AddProduct(c *gin.Context){


	ProductName := c.PostForm("name")
	ProductSubCat,_ := strconv.Atoi(c.PostForm("subcategory"))
	ProductDescription := c.PostForm("description")

	// product variant details
	ProductVariantName := c.PostForm("variant_name")
	ProductSize := c.PostForm("size")
	ProductStock,_ := strconv.Atoi(c.PostForm("stock"))
	ProductPrice,_ := strconv.ParseFloat(c.PostForm("price"),64) 
	ProductTax,_ := strconv.ParseFloat(c.PostForm("tax"),64)

	var subCat models.SubCategory
	

	if err := db.Db.Find(&subCat, ProductSubCat).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"category id does not exist"})
		return
	}

	if subCat.CategoryID == 0 {
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"category id does not exist"})
		return
	}

	product := models.Product{
		ProductName: ProductName,
		Description: ProductDescription,
		SubCategoryID: uint(ProductSubCat),
	}

	if err := db.Db.Create(&product).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":err.Error()})
		return
	}

	variant := models.Product_Variant{
		ProductID: product.ProductID,
		Variant_name: ProductVariantName,
		Size: ProductSize,
		Stock: ProductStock,
		Price: ProductPrice,
		Tax: ProductTax,
	}
	
	if err := db.Db.Create(&variant).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":err.Error()})
		return 
	}

	for i := 0; i < 3;i++{

		base64Str := c.PostForm(fmt.Sprintf("cropped_image%d",i))

		if base64Str != ""{
			
			data := strings.Split(base64Str,",")[1]
			decoded, err := base64.StdEncoding.DecodeString(data)
			
			if err != nil{
				continue
			}

			filename := fmt.Sprintf("uploads/cropped_%d_%d.jpg",time.Now().UnixNano(),i)

			if err := ioutil.WriteFile(filename,decoded,0644); err != nil{
				continue
			}

			order,_ := strconv.Atoi(c.PostForm(fmt.Sprintf("order%d",i)))
			isPrimary := c.PostForm(fmt.Sprintf("is_primary%d",i))=="true"

			image := models.Product_image{
				ProductVariantID: variant.ID,
				Image_url: filename,
				Order_no: order,
				Is_primary: isPrimary,
				CreatedAt: time.Now(),
			}

			if err := db.Db.Create(&image).Error; err != nil{
				c.String(http.StatusInternalServerError,"Error saving image to DB: %v", err)
				return
			}


		}

	}

	c.Redirect(http.StatusSeeOther,"/admin/products")

}

func UpdateProductPage(c *gin.Context){
	
	productID,_ := strconv.Atoi(c.Param("id"))

	var Product_Variant models.Product_Variant
	var Images []models.Product_image
	var Product models.Product

	if err := db.Db.Where("deleted_at IS NULL").First(&Product_Variant,productID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product not found"})
		return 
	}

	if err := db.Db.Where("deleted_at IS NULL AND product_id = ?",Product_Variant.ProductID).First(&Product).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product not found"})
		return
	}

	if err := db.Db.Find(&Images,productID).Error; err != nil{
		c.String(http.StatusNotFound,"Error loading image from DB: %v",err)
		return
	}

	c.HTML(http.StatusFound,"edit_Product.html",gin.H{
		"Product": Product,
		"Variant": Product_Variant,
		"Images": Images,
	})
	
}

func UpdateProduct(c *gin.Context){
	
	productID,_ := strconv.Atoi(c.Param("id"))
	ProductName := c.PostForm("name")
	ProductSubCat,_ := strconv.Atoi(c.PostForm("subcategory"))
	ProductDescription := c.PostForm("description")

	// product variant details
	ProductVariantName := c.PostForm("variant_name")
	ProductSize := c.PostForm("size")
	ProductStock,_ := strconv.Atoi(c.PostForm("stock"))
	ProductPrice,_ := strconv.ParseFloat(c.PostForm("price"),64) 
	ProductTax,_ := strconv.ParseFloat(c.PostForm("tax"),64)

	var Product models.Product
	var Product_variant models.Product_Variant

	if err := db.Db.First(&Product_variant,productID).Error;err != nil{
		c.String(http.StatusNotFound,"Error loading product detail from DB: %v",err)
		return
	}

	if err := db.Db.Where("product_id = ?", Product_variant.ProductID).First(&Product).Error;err != nil{
		c.String(http.StatusNotFound,"Error loading product detail from DB: %v",err)
		return
	}

	Product.ProductName = ProductName
	Product.SubCategoryID = uint(ProductSubCat)
	Product.Description = ProductDescription

	if err := db.Db.Save(Product).Error;err != nil{
		c.String(http.StatusInternalServerError, "Failed to update product")
        return
	}
	
	Product_variant.Variant_name = ProductVariantName
	Product_variant.Size = ProductSize
	Product_variant.Stock = ProductStock
	Product_variant.Price = ProductPrice
	Product_variant.Tax = ProductTax

	if err := db.Db.Save(Product_variant).Error; err!= nil{
		c.String(http.StatusInternalServerError, "Failed to update product")
        return
	}

	for i := 0;i < 3;i++{
		
		file,err := c.FormFile(fmt.Sprintf("image%d",i))
		if err == nil && file.Size > 0 {
			
			filename := fmt.Sprintf("upload/%d_%s",time.Now().UnixNano(),file.Filename)
			
			if err := c.SaveUploadedFile(file,filename); err != nil{
				c.String(http.StatusInternalServerError, "Failed to save image")
                return
			}

			orderNo,_ := strconv.Atoi(c.PostForm(fmt.Sprintf("order%d",i)))
			isPrimary := c.PostForm(fmt.Sprintf("is_primary%d",i)) == "time"

			image := models.Product_image{
				ProductVariantID: Product_variant.ID,
				Image_url: filename,
				Order_no: orderNo,
				Is_primary: isPrimary,
				CreatedAt: time.Now(),
			}

			if err:= db.Db.Create(&image).Error;err != nil{
				c.String(http.StatusInternalServerError, "Failed to save image in DB")
                return
			}
		} 
	}

	c.Redirect(http.StatusSeeOther,"/admin/products")
}

func DeleteProduct(c *gin.Context){
	
	id := c.Param("id")
	var Product_variant models.Product_Variant
	
	if err := db.Db.First(&Product_variant,id).Error; err!=nil{
		c.String(http.StatusNotFound, "Product not found")
        return
	}

	if err := db.Db.Delete(&Product_variant).Error; err != nil{
		c.String(http.StatusInternalServerError, "Failed to delete product")
        return
	}

	c.Redirect(http.StatusSeeOther,"/admin/products")
}