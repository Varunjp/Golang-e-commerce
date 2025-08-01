package admin

import (
	"bytes"
	"encoding/base64"
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"first-project/models/responsemodels"
	"first-project/utils"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
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

func isValidImage(data []byte) bool {
    reader := bytes.NewReader(data)
    _, format, err := image.Decode(reader)
    if err != nil {
        return false
    }
    switch format {
    case "jpeg", "png", "gif":
        return true
    default:
        return false
    }
}

func ViewProducts(c *gin.Context){

	var Products []models.Product_Variant

	session := sessions.Default(c)
	name := session.Get("admin-name").(string)

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

		db.Db.Model(&models.Product_Variant{}).Count(&total)

		err := db.Db.Model(&models.Product_Variant{}).Preload("Product_images").Order("id DESC").Offset(offset).Find(&Products).Error
		
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

	Errors := make(map[string]string)

	ProductName := c.PostForm("name")
	ProductSubCat := c.PostForm("subcategory_id")
	ProductDescription := c.PostForm("description")
	
	ProductName = strings.TrimLeft(ProductName," ")
	ProductDescription = strings.TrimLeft(ProductDescription," ")
	// product variant details
	ProductVariantName := c.PostForm("variant_name")
	ProductSize := c.PostForm("size")
	ProductStock,_ := strconv.Atoi(c.PostForm("stock"))
	ProductPrice,_ := strconv.ParseFloat(c.PostForm("price"),64) 
	ProductTax,_ := strconv.ParseFloat(c.PostForm("tax"),64)
	ProductDiscount,_ := strconv.ParseFloat(c.PostForm("discount_price"),64)

	ProductVariantName = strings.TrimLeft(ProductVariantName," ")
	ProductSize = strings.TrimSpace(ProductSize)

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

	ProductSize = utils.SizeAdjust(ProductSize)

	var pcount int64

	if err := db.Db.Model(models.Product{}).Where("product_name ILIKE ","%"+ProductName+"%").Count(&pcount).Error; err == nil{
		if pcount > 0 {
			Errors["name"] = "Product already exist"
			
		}
	}

	icount := 0

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

			icount++
		}
	}

	var subCat models.SubCategory
	var ProductCheck models.Product
	
	if ProductPrice <= 0 {
		Errors["price"] = "Price must be greater than 0"
		
	}

	if err := db.Db.Where("product_name ILIKE ?",ProductName).First(&ProductCheck).Error; err == nil{
		Errors["name"] = "Already exist"
		
	}

	if ProductStock < 1 {
		Errors["stock"] = "Stock must be greater than 0"
		
	}

	if icount == 0 {
		Errors["cropped_image0"] = "At least one image is required"
	}

	if ProductDiscount > 0 {
		if ProductDiscount < ProductPrice {
			Errors["discount_price"] = "Discount price need to be greater than sale price"
		}
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest,gin.H{"status":"error","errors":Errors})
		return
	}

	if err := db.Db.Where("sub_category_id = ?",ProductSubCat).First(&subCat).Error; err != nil{
		
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error","message": "Something went wrong"})
		return
	}

	
	if subCat.CategoryID == 0 {
		
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error","message": "Something went wrong"})
		return
	}

	productSubcatInt,_ := strconv.Atoi(ProductSubCat)

	product := models.Product{
		ProductName: ProductName,
		Description: ProductDescription,
		SubCategoryID: uint(productSubcatInt),
	}

	if err := db.Db.Create(&product).Error; err != nil{
		
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error","message": "Something went wrong"})
		return
	}

	variant := models.Product_Variant{
		ProductID: product.ProductID,
		Variant_name: ProductVariantName,
		Size: ProductSize,
		Stock: ProductStock,
		Price: ProductPrice,
		DiscountedPrice: ProductDiscount,
		Tax: ProductTax,
	}
	
	if err := db.Db.Create(&variant).Error; err != nil{
		
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error","message": "Something went wrong"})
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
				
				c.JSON(http.StatusBadRequest,gin.H{"status":"error"})
				return
			}

			if !isValidImage(decoded) {
		
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
				
				c.JSON(http.StatusBadRequest,gin.H{"status":"error"})
				return
			}

			imageCount++
		}
	
		
	}

	if imageCount < 1 {
		db.Db.Delete(&models.Product{},product.ProductID)
		db.Db.Delete(&models.Product_Variant{},variant.ID)
		
		c.JSON(http.StatusBadRequest,gin.H{"status":"error","message": "Upload atleast 1 image"})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status": "success",
		"message": "Product added successfully",
		"redirect": "/admin/products",
	})

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
	Errors := make(map[string]string )
	// form data
	productID := c.PostForm("product_id")
	variant_name := c.PostForm("variant_name")
	size := c.PostForm("size")
	ProductStock,_ := strconv.Atoi(c.PostForm("stock"))
	ProductPrice,_ := strconv.ParseFloat(c.PostForm("price"),64) 

	if strings.TrimSpace(variant_name) == ""{
		Errors["variant_name"] = "Invalid name entry"
	}

	if strings.TrimSpace(size) == ""{
		Errors["size"] = "Invalid size entry"
	}

	if ProductStock < 1 {
		Errors["stock"] = "Stock cannot be less than zero"
	}

	if ProductPrice < 1 {
		Errors["price"] = "Price cannot be less than zero"
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest,gin.H{"status":"error","errors":Errors})
		return 
	}

	var Product models.Product
	var ProductImage []models.Product_image
	var ProductVariant models.Product_Variant

	if err := db.Db.Preload("Product_variants").Where("product_id = ?",productID).First(&Product).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error","message":"Could not load product details"})
		return 
	}

	productVariantID := Product.Product_variants[0].ID

	if err := db.Db.Where("product_variant_id = ?",productVariantID).Find(&ProductImage).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error","message":"Could not load product images."})
		return 
	}

	if err := db.Db.Where("product_id = ? AND size ILIKE ?",productID,size).First(&ProductVariant).Error; err == nil{
		ProductVariant.Stock = ProductVariant.Stock + ProductStock
		ProductVariant.Price = ProductPrice

		db.Db.Save(&ProductVariant)
		
		c.JSON(http.StatusOK,gin.H{
			"status":"success",
			"message":"Variant updated",
			"redirect":"/admin/products",
		})
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
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error","message":"Error while creating new variant, please try again later"})
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

	c.JSON(http.StatusOK,gin.H{
		"status":"success",
		"message":"Variant added successfully",
		"redirect":"/admin/products",
	})
}



func DeleteImage(c *gin.Context){
	
	ID := c.Param("id")
	var Image models.Product_image
	var productImages []models.Product_image

	if err := db.Db.Where("product_image_id = ?",ID).First(&Image).Error; err != nil{
		c.Redirect(http.StatusSeeOther,"/admin/products")
		return
	}

	if err := db.Db.Where("product_variant_id = ?",Image.ProductVariantID).Find(&productImages).Error; err != nil{
		c.Redirect(http.StatusSeeOther,"/admin/products/edit/"+strconv.Itoa(int(Image.ProductVariantID)))
		return 
	}

	if len(productImages) == 1 {
		c.Redirect(http.StatusSeeOther,"/admin/products/edit/"+strconv.Itoa(int(Image.ProductVariantID)))
		return 
	}

	if err := db.Db.Delete(&Image).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"edit_Product.html",gin.H{
			"error":"Failed to delete image "+err.Error(),
		})
		return
	}

	c.Redirect(http.StatusSeeOther,"/admin/products")
}

func DeleteProduct(c *gin.Context){
	
	id := c.Param("id")
	var Product_variant models.Product_Variant
	var Product models.Product

	if err := db.Db.First(&Product_variant,id).Error; err!=nil{
		c.HTML(http.StatusNotFound,"admin_product_list.html", gin.H{"error":"Product not found"})
        return
	}


	err := helper.CancelOrderForProduct(id)

	if err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"Failed to remove product from orders"})
		return 
	}

	if err := db.Db.Delete(&Product_variant).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html", gin.H{"error":"Failed to delete product"})
        return
	}

	if err := db.Db.Preload("Product_variants",func(db *gorm.DB)*gorm.DB{
		return db.Unscoped()
	}).Where("product_id = ?",Product_variant.ProductID).First(&Product).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"admin_product_list.html",gin.H{"error":"Failed get product details"})
		return 
	}

	totalVariant := len(Product.Product_variants)
	var delCount int 

	for _,vari := range Product.Product_variants{
		if vari.DeletedAt.Valid {
			delCount ++
		}
	}

	if totalVariant == delCount {
		db.Db.Delete(&Product)
	}


	c.Redirect(http.StatusSeeOther,"/admin/products")
}