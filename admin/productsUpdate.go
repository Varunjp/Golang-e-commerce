package admin

import (
	"encoding/base64"
	db "first-project/DB"
	"first-project/models"
	"first-project/utils"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)


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

func UpdateProductHandler(c *gin.Context){
	Errors := make(map[string]string)
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
	Productdiscount,_ := strconv.ParseFloat(c.PostForm("discount_price"),64)

	if strings.TrimSpace(ProductName) == "" {
		Errors["name"] = "Name should be properly defined"	
	}

	if strings.TrimSpace(ProductVariantName) == ""{
		Errors["varian_name"] = "Variant name should be properly defined"
	}

	if strings.TrimSpace(ProductDescription) == ""{
		Errors["description"] = "Provide proper description"
	}

	if ProductSize == ""{
		Errors["size"] = "Size should be properly defined"
	}
	var Product models.Product
	var Product_variant models.Product_Variant

	if err := db.Db.Preload("Product_images").First(&Product_variant,productID).Error;err != nil{
		c.JSON(http.StatusNotFound,gin.H{"status":"error","message":"Error loading product detail from DB"})
		return
	}

	if err := db.Db.Where("product_id = ?", Product_variant.ProductID).First(&Product).Error;err != nil{
		c.JSON(http.StatusNotFound,gin.H{"status":"error","message":"Error loading product detail from DB"})
		return
	}

	var pcount int64
	var existProduct models.Product

	if err := db.Db.Model(models.Product{}).Where("product_name ILIKE ? AND product_id != ?",ProductName,Product.ProductID).First(&existProduct).Count(&pcount).Error; err == nil{
		if pcount > 0 {
			Errors["name"] = "Product already exist"
			
		}
	}

	ProductSize,pserr := utils.SizeAdjust(ProductSize)

	if pserr != nil {
		Errors["size"] = pserr.Error()
	}

	// image check
	for i := 0; i < 3; i++{
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
				Errors[fmt.Sprintf("cropped_image%d", i)] = "Image is invalid"
				continue
			}

			if !isValidImage(decoded) {
				Errors[fmt.Sprintf("cropped_image%d", i)] = "Image is invalid"
				continue
			}
			
		}
	}

	if ProductPrice <= 0 {
		Errors["price"] = "Price must be greater than 0"
		
	}

	if ProductStock < 1 {
		Errors["stock"] = "Stock must be greater than 0"
		
	}

	if Productdiscount > 0{
		if Productdiscount < ProductPrice {
			Errors["discount_price"] = "Price need to be more than sale price"
		}
	}


	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest,gin.H{"status":"error","errors":Errors})
		return
	}

	

	Product.ProductName = ProductName
	Product.SubCategoryID = uint(ProductSubCat)
	Product.Description = ProductDescription

	if err := db.Db.Save(Product).Error;err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"status":"error","message":"Failed to update product"})
        return
	}
	
	Product_variant.Variant_name = ProductVariantName
	Product_variant.Size = ProductSize
	Product_variant.Stock = ProductStock
	Product_variant.Price = ProductPrice
	Product_variant.DiscountedPrice = Productdiscount
	Product_variant.Tax = ProductTax

	if err := db.Db.Save(Product_variant).Error; err!= nil{
		c.JSON(http.StatusInternalServerError, gin.H{"status":"error","message":"Failed to update product"})
        return
	}

	for i := 0; i < 3; i++ {
		
		base64Str := c.PostForm(fmt.Sprintf("cropped_image%d", i))
	
		if base64Str != "" {
			
			var ProductImage []models.Product_image
			if err := db.Db.Where("product_variant_id = ?", productID).Find(&ProductImage).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to load existing images"})
				return	
			}

			if len(ProductImage) >= 4 {
				continue
			}

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
				c.JSON(http.StatusBadRequest,gin.H{"status":"error"})
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
			if order < 1 {
				order = len(Product_variant.Product_images) + 1
			}
			isPrimary := c.PostForm(fmt.Sprintf("is_primary%d", i)) == "true"
		
			image := models.Product_image{
				ProductVariantID: uint(productID),
				Image_url:        filename,
				Order_no:         order,
				Is_primary:       isPrimary,
				CreatedAt:        time.Now(),
			}
		
			if err := db.Db.Create(&image).Error; err != nil {
				c.JSON(http.StatusBadRequest,gin.H{"status":"error"})
				return
			}
		}
	
	}

	c.JSON(http.StatusOK,gin.H{
		"status": "success",
		"message": "Product added successfully",
		"redirect": "/admin/products",
	})
}