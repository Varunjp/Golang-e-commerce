package admin

import (
	"encoding/base64"
	db "first-project/DB"
	"first-project/models"
	"first-project/models/responsemodels"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ViewProducts(c *gin.Context){

	var Products []models.Product_Variant

	session := sessions.Default(c)
	name := session.Get("name").(string)
	// var dbProducts []struct {
	// 	ProductID int `gorm:"column:id"`
	// 	ProductName string `gorm:"column:variant_name"`
	// 	ProductSize string `gorm:"column:size"`
	// 	ProductImage string	`gorm:"column:image_url"`
	// 	ProductPrice float64 `gorm:"column:price"`
	// 	ProductStock int	`gorm:"column:stock"`
	// 	CreatedAt	time.Time	`gorm:"column:created_at"`
	// }

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

	var total int64

	keyword := c.Query("search")

	if keyword == ""{
			
		// err := db.Db.Table("product_variants").Select("product_variants.id,product_variants.variant_name,product_variants.size,product_images.image_url,product_variants.price,product_variants.stock").Joins("LEFT JOIN product_images ON product_images.product_variant_id = product_variants.id").Where("product_variants.deleted_at IS NULL").Group("product_variants.id,product_variants.variant_name,product_images.image_url").Order("product_variants.id Desc").Offset(offset).Find(&dbProducts).Error

		db.Db.Model(&models.Product_Variant{}).Count(&total)

		err := db.Db.Model(&models.Product_Variant{}).Preload("Product_images",func(db *gorm.DB)*gorm.DB{
			return db.Where("order_no = ?",1)
		}).Order("id DESC").Offset(offset).Find(&Products).Error
		

		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		if len(Products) == 0{
			c.HTML(http.StatusOK,"admin_product_list.html",gin.H{"message":"No products listed"})
			return
		}

		responseProducts := make([]responsemodels.Products,len(Products))


		for i, dbProduct := range Products{
			status := true 
			if dbProduct.Stock == 0{
				status = false 
			}
			responseProducts[i] = responsemodels.Products{
				ID: dbProduct.ID,
				Name: dbProduct.Variant_name,
				Size: dbProduct.Size,
				Price: dbProduct.Price,
				Quantity: dbProduct.Stock,
				CreatedAt: dbProduct.CreatedAt,
				InStock: status,
			}
			
			if len(dbProduct.Product_images) > 0{
				responseProducts[i].ImageURl = dbProduct.Product_images[0].Image_url

			}else{
				responseProducts[i].ImageURl = ""
			}

		}


		totalPages := int(math.Ceil(float64(total)/ float64(limit)))


		c.HTML(http.StatusOK,"admin_product_list.html",gin.H{"products":responseProducts,"page":page,
		"limit":limit,
		"totalPages":totalPages,"user":name})

	}else{


		// err := db.Db.Table("product_variants").Select("product_variants.id,product_variants.variant_name,product_variants.size,product_images.image_url,product_variants.price,product_variants.stock").Joins("LEFT JOIN product_images ON product_images.product_variant_id = product_variants.id").Where("product_variants.variant_name ILIKE ?","%"+keyword+"%").Group("product_variants.id,product_variants.variant_name,product_images.image_url").Order("product_variants.id Desc").Offset(offset).Find(&dbProducts).Error

		db.Db.Model(&models.Product_Variant{}).Where("product_variants.variant_name ILIKE ?","%"+keyword+"%").Count(&total)

		err := db.Db.Model(&models.Product_Variant{}).Preload("Product_images",func(db *gorm.DB)*gorm.DB{
			return db.Where("order_no = ?",1)
		}).Where("product_variants.variant_name ILIKE ?","%"+keyword+"%").Offset(offset).Find(&Products).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		if total == 0 {
			c.HTML(http.StatusOK,"admin_product_list.html",gin.H{"error":"Product not found"})
			return
		}

		responseProducts := make([]responsemodels.Products,len(Products))

		for i, dbProduct := range Products{
			status := true 
			if dbProduct.Stock == 0{
				status = false 
			}
			responseProducts[i] = responsemodels.Products{
				ID: uint(dbProduct.ID),
				Name: dbProduct.Variant_name,
				Size: dbProduct.Size,
				Price: dbProduct.Price,
				Quantity: dbProduct.Stock,
				CreatedAt: dbProduct.CreatedAt,
				InStock: status,
			}

			if len(dbProduct.Product_images) > 0{
				responseProducts[i].ImageURl = dbProduct.Product_images[0].Image_url	
			}else{
				responseProducts[i].ImageURl = ""
			}
			
		}

		totalPages := int(math.Ceil(float64(total)/ float64(limit)))

		c.HTML(http.StatusOK,"admin_product_list.html",gin.H{"products":responseProducts,"page":page,
		"limit":limit,
		"totalPages":totalPages,"user":name})

	}

}

func AddProductPage(c *gin.Context){

	var subcat []models.SubCategory
	
	type Response struct {
		SubCategoryID 		int
		SubCategoryName		string 
		CategoryName		string 
	}


	if err := db.Db.Where("is_blocked = ?",false).Find(&subcat).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_addProduct.html",gin.H{"error":"Failed to load subcategory,please try again later"})
		return 
	}

	response := make([]Response,len(subcat))

	for i,subitem := range subcat {

		var Category models.Category

		if err := db.Db.Where("category_id = ?",subitem.CategoryID).First(&Category).Error; err != nil{
			log.Println("Failed to load category details")
			c.Redirect(http.StatusTemporaryRedirect,"/admin")
			return 
		}

		response[i] = Response{
			SubCategoryID: int(subitem.SubCategoryID),
			SubCategoryName: subitem.SubCategoryName,
			CategoryName: Category.CategoryName,
		}

	}

	c.HTML(http.StatusOK,"admin_addProduct.html",gin.H{"Subcategories":response})
}

func AddProduct(c *gin.Context){


	ProductName := c.PostForm("name")
	ProductSubCat := c.PostForm("subcategory_id")
	ProductDescription := c.PostForm("description")


	// product variant details
	ProductVariantName := c.PostForm("variant_name")
	ProductSize := c.PostForm("size")
	ProductStock,_ := strconv.Atoi(c.PostForm("stock"))
	ProductPrice,_ := strconv.ParseFloat(c.PostForm("price"),64) 
	ProductTax,_ := strconv.ParseFloat(c.PostForm("tax"),64)

	var subCat models.SubCategory
	var ProductCheck models.Product
	
	if ProductPrice <= 0 {
		c.HTML(http.StatusBadRequest,"admin_product_list.html",gin.H{"error":"Price cannot be 0 or less"})
		return
	}

	if err := db.Db.Where("product_name ILIKE ?",ProductName).First(&ProductCheck).Error; err == nil{
		c.HTML(http.StatusConflict,"admin_product_list.html",gin.H{"error":"Product already exist"})
		return 
	}

	if ProductStock < 1 {
		c.HTML(http.StatusBadRequest,"admin_product_list.html",gin.H{"error":"Quanity cannot be less than 1"})
		return 
	}

	if err := db.Db.Where("sub_category_id = ?",ProductSubCat).First(&subCat).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"category id does not exist"})
		return
	}

	
	if subCat.CategoryID == 0 {
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"category id does not exist"})
		return
	}

	productSubcatInt,_ := strconv.Atoi(ProductSubCat)

	product := models.Product{
		ProductName: ProductName,
		Description: ProductDescription,
		SubCategoryID: uint(productSubcatInt),
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

	imageCount := 0

	for i := 0; i < 3; i++ {
		
		base64Str := c.PostForm(fmt.Sprintf("cropped_image%d", i))
	
		if base64Str != "" {
			
			var base64Data string
			if strings.Contains(base64Str, ",") {
				// Format: data:image/jpeg;base64,<data>
				parts := strings.SplitN(base64Str, ",", 2)
				base64Data = parts[1]
			} else {
				// Raw base64 only
				base64Data = base64Str
			}
		
			decoded, err := base64.StdEncoding.DecodeString(base64Data)
			if err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("Failed to decode image %d: %v", i+1, err))
				return
			}
		
			// Ensure upload folder exists
			if _, err := os.Stat("upload"); os.IsNotExist(err) {
				os.Mkdir("upload", 0755)
			}
		
			filename := fmt.Sprintf("upload/cropped_%d_%d.jpg", time.Now().UnixNano(), i)
			if err := os.WriteFile(filename, decoded, 0644); err != nil {
				continue
			}
		
			order, _ := strconv.Atoi(c.PostForm(fmt.Sprintf("order%d", i)))
			checkPrimary := c.PostForm("is_primary")
			var isPrimary bool

			if checkPrimary != ""{
				isPrimary = true 
			}else{
				isPrimary = false
			}
			
		
			image := models.Product_image{
				ProductVariantID: variant.ID,
				Image_url:        filename,
				Order_no:         order,
				Is_primary:       isPrimary,
				CreatedAt:        time.Now(),
			}
		
			if err := db.Db.Create(&image).Error; err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving image %d to DB: %v", i+1, err))
				return
			}

			imageCount++
		}
	
		
	}

	if imageCount < 1 {
		db.Db.Delete(&models.Product{},product.ProductID)
		db.Db.Delete(&models.Product_Variant{},variant.ID)
		c.HTML(http.StatusBadRequest,"admin_product_list.html",gin.H{"error":"Provide atleast 1 image"})
		return
	}

	c.Redirect(http.StatusSeeOther,"/admin/products")

}

func AddProductVariantPage(c *gin.Context){
	var Products []models.Product
	type response struct {
		ProductID  		uint
		ProductName 	string
		Type 			string 
	}

	if err := db.Db.Find(&Products).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"Could not load product details."})
		return 
	}

	responseProduct := make([]response,len(Products))

	for i, pro := range Products{
		var subCategory models.SubCategory
		db.Db.Where("sub_category_id = ?",pro.SubCategoryID).First(&subCategory)
		if strings.Contains(subCategory.SubCategoryName,"shoe"){
			responseProduct[i] = response{
				ProductID: pro.ProductID,
				ProductName: pro.ProductName,
				Type: "shoes",
			}
		}else{
			responseProduct[i] = response{
				ProductID: pro.ProductID,
				ProductName: pro.ProductName,
				Type: "clothing",
			}
		}
	}

	c.HTML(http.StatusOK,"admin_addProductVariant.html",gin.H{"Products":responseProduct})

}

func AddProductVariant (c *gin.Context){
	// form data
	productID := c.PostForm("product_id")
	variant_name := c.PostForm("variant_name")
	size := c.PostForm("size")
	ProductStock,_ := strconv.Atoi(c.PostForm("stock"))
	ProductPrice,_ := strconv.ParseFloat(c.PostForm("price"),64) 

	if ProductStock < 1 || ProductPrice < 1 {
		c.HTML(http.StatusBadRequest,"admin_addProductVariant.html",gin.H{"error":"Stock or price could not be less than 1"})
		return 
	}

	var Product models.Product
	var ProductImage []models.Product_image
	var ProductVariant models.Product_Variant

	if err := db.Db.Preload("Product_variants").Where("product_id = ?",productID).First(&Product).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"Could not load product details"})
		return 
	}

	productVariantID := Product.Product_variants[0].ID

	if err := db.Db.Where("product_variant_id = ?",productVariantID).Find(&ProductImage).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"Could not load product images."})
		return 
	}

	if err := db.Db.Where("product_id = ? AND size ILIKE ?",productID,size).First(&ProductVariant).Error; err == nil{
		ProductVariant.Stock = ProductVariant.Stock + ProductStock
		ProductVariant.Price = ProductPrice

		db.Db.Save(&ProductVariant)

		c.Redirect(http.StatusSeeOther,"/admin/products")
		return 
	}

	newProductVariant := models.Product_Variant{
		Variant_name: variant_name,
		ProductID: Product.ProductID,
		Size: size,
		Stock: ProductStock,
		Price: ProductPrice,
		Tax: Product.Product_variants[0].Tax,
	}

	if err := db.Db.Create(&newProductVariant).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"Error while creating new variant, please try again later"})
		return 
	}

	for _, image := range ProductImage{
		newProductImage := models.Product_image{
			ProductVariantID: newProductVariant.ID,
			Image_url: image.Image_url,
			Is_primary: image.Is_primary,
			Order_no: image.Order_no,
		}

		db.Db.Create(&newProductImage)
	}

	c.Redirect(http.StatusSeeOther,"/admin/products")
}

func UpdateProductPage(c *gin.Context){
	
	productID,_ := strconv.Atoi(c.Param("id"))

	var subcat []models.SubCategory
	
	type Response struct {
		SubCategoryID 		int
		SubCategoryName		string 
		CategoryName		string 
	}


	if err := db.Db.Where("is_blocked = ?",false).Find(&subcat).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_addProduct.html",gin.H{"error":"Failed to load subcategory,please try again later"})
		return 
	}

	response := make([]Response,len(subcat))

	for i,subitem := range subcat {

		var Category models.Category

		if err := db.Db.Where("category_id = ?",subitem.CategoryID).First(&Category).Error; err != nil{
			log.Println("Failed to load category details")
			c.Redirect(http.StatusTemporaryRedirect,"/admin")
			return 
		}

		response[i] = Response{
			SubCategoryID: int(subitem.SubCategoryID),
			SubCategoryName: subitem.SubCategoryName,
			CategoryName: Category.CategoryName,
		}

	}

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

	if err := db.Db.Where("product_variant_id = ?",productID).Find(&Images).Error; err != nil{
		log.Println("No images found:",err.Error())
	}


	c.HTML(http.StatusFound,"edit_Product.html",gin.H{
		"Product": Product,
		"Variant": Product_Variant,
		"Images": Images,
		"Subcategories":response,
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

	if ProductStock < 1 {
		c.HTML(http.StatusBadRequest,"admin_product_list.html",gin.H{"error":"Product stock cannot be updated below 1"})
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

	// for i := 0;i < 3;i++{
		
		
	// 	base64Str := c.PostForm(fmt.Sprintf("cropped_image%d",i))

	// 	if base64Str != ""{
			
	// 		data := strings.Split(base64Str,",")[1]
	// 		decoded, err := base64.StdEncoding.DecodeString(data)
			
	// 		if err != nil{
	// 			continue
	// 		}

	// 		filename := fmt.Sprintf("uploads/cropped_%d_%d.jpg",time.Now().UnixNano(),i)

	// 		if err := os.WriteFile(filename,decoded,0644); err != nil{
	// 			continue
	// 		}

	// 		order,_ := strconv.Atoi(c.PostForm(fmt.Sprintf("order%d",i)))
	// 		isPrimary := c.PostForm(fmt.Sprintf("is_primary%d",i))=="true"

	// 		image := models.Product_image{
	// 			ProductVariantID: uint(productID),
	// 			Image_url: filename,
	// 			Order_no: order,
	// 			Is_primary: isPrimary,
	// 			CreatedAt: time.Now(),
	// 		}


	// 		if err := db.Db.Create(&image).Error; err != nil{
	// 			c.String(http.StatusInternalServerError,"Error saving image to DB: %v", err)
	// 			return
	// 		}

			

	// 	} 
	// }


	for i := 0; i < 3; i++ {
		
		base64Str := c.PostForm(fmt.Sprintf("cropped_image%d", i))
	
		if base64Str != "" {
			
			var base64Data string
			if strings.Contains(base64Str, ",") {
				// Format: data:image/jpeg;base64,<data>
				parts := strings.SplitN(base64Str, ",", 2)
				base64Data = parts[1]
			} else {
				// Raw base64 only
				base64Data = base64Str
			}
		
			decoded, err := base64.StdEncoding.DecodeString(base64Data)
			if err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("Failed to decode image %d: %v", i+1, err))
				return
			}
		
			// Ensure upload folder exists
			if _, err := os.Stat("upload"); os.IsNotExist(err) {
				os.Mkdir("upload", 0755)
			}
		
			filename := fmt.Sprintf("upload/cropped_%d_%d.jpg", time.Now().UnixNano(), i)
			if err := os.WriteFile(filename, decoded, 0644); err != nil {
				continue
			}
		
			order, _ := strconv.Atoi(c.PostForm(fmt.Sprintf("order%d", i)))
			isPrimary := c.PostForm(fmt.Sprintf("is_primary%d", i)) == "true"
		
			image := models.Product_image{
				ProductVariantID: uint(productID),
				Image_url:        filename,
				Order_no:         order,
				Is_primary:       isPrimary,
				CreatedAt:        time.Now(),
			}
		
			if err := db.Db.Create(&image).Error; err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving image %d to DB: %v", i+1, err))
				return
			}
		}
	
		
	}

	

	c.Redirect(http.StatusSeeOther,"/admin/products")
}

func DeleteImage(c *gin.Context){
	
	ID := c.Param("id")
	var Image models.Product_image

	if err := db.Db.Where("product_image_id = ?",ID).First(&Image).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"edit_Product.html",gin.H{
			"error":"Image not found "+err.Error(),
		})
		return
	}

	if err := db.Db.Delete(&Image).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"edit_Product.html",gin.H{
			"error":"Failed to delete image "+err.Error(),
		})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect,"/admin/products")
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