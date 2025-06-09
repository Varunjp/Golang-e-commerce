package user

import (
	db "first-project/DB"
	"first-project/helper"
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

	var subcat []models.SubCategory
	type responsecat struct{
		SubID 		uint
		SubCategoryName	string
		CategoryName   string
	}

	
	// retrieve filters

	if err := db.Db.Where("is_blocked = ?",false).Find(&subcat).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to load categories"})
		return 
	}

	responeSubcategory := make([]responsecat,len(subcat))

	for i,subItem := range subcat {
		var category models.Category
		if err := db.Db.Where("category_id = ?",subItem.CategoryID).First(&category).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to load categories, please try again later"})
			return 
		}

		responeSubcategory[i] = responsecat{
			SubID: subItem.SubCategoryID,
			SubCategoryName: subItem.SubCategoryName,
			CategoryName: category.CategoryName,
		}

	}

	categories := c.QueryArray("category")
	size := c.Query("size")
	minPrice,_ := strconv.ParseFloat(c.DefaultQuery("min_price","0"),64)
	maxPrice,_ := strconv.ParseFloat(c.DefaultQuery("max_price","10000"),64)
	sortBy := c.DefaultQuery("sort","")

	query := db.Db.
    Preload("Product_images", func(db *gorm.DB) *gorm.DB {
        return db.Order("order_no ASC")
    }).Where("is_active = ? AND product_variants.deleted_at IS NULL",true).Model(&models.Product_Variant{})
	

	if len(categories) > 0 {
		query = query.Joins("JOIN products ON products.product_id = product_variants.product_id").Where("products.sub_category_id IN ?",categories)
	}

	if size != ""{
		query = query.Where("size = ?",size)
	}

	query = query.Where("price BETWEEN ? AND ?",minPrice,maxPrice)

	if sortBy == "price_asc"{
		query = query.Order("price ASC")
	}else if sortBy == "price_desc"{
		query = query.Order("price DESC")
	}

	page,_ := strconv.Atoi(pageStr)
	limit,_ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit
	
	var total int64
	var products []models.Product_Variant

	query.Count(&total)
	query.Where("stock > 0").Offset(offset).Limit(limit).Find(&products)
	

	// db.Db.Model(&models.Product_Variant{}).Count(&total)

	// err := db.Db.
    // Preload("Product").
    // Preload("Product.SubCategory").
    // Preload("Product_images", func(db *gorm.DB) *gorm.DB {
    //     return db.Order("order_no ASC")
    // }).
    // Offset(offset).
    // Limit(limit).
    // Find(&products).Error

	
	if query.Error != nil{
		c.HTML(http.StatusInternalServerError, "product_list.html", gin.H{"error": "Error fetching products"})
        return
	}

	formatted := make([]struct{
		ID					uint
		Name				string
		Price				float64
		DiscountedPrice		string 
		ImageURL 			string
		Wishlist 			bool 
	},len(products))

	tokenStr,errtok:= c.Cookie("JWT-User")
	var userID float64

	if errtok == nil {
		_,userId,_ := helper.DecodeJWT(tokenStr)
		userID = userId
	}
	

	for i,p := range products{

		var pro models.Product
		var subCat models.SubCategory

		db.Db.Where("product_id  = ?",p.ProductID).First(&pro)
		db.Db.Where("sub_category_id = ?",pro.SubCategoryID).First(&subCat)

		if !subCat.IsBlocked{

			var wishlist models.WishList
			isWishlist := false

			if userID != 0 {
				if err := db.Db.Where("user_id = ? AND product_id = ?",userID,p.ID).First(&wishlist).Error; err == nil{
					isWishlist = true
				}
			}
			

			if p.Stock > 0 {
				formatted[i] = struct{ID uint; Name string; Price float64; DiscountedPrice string; ImageURL string; Wishlist bool}{
					ID: p.ID,
					Name: p.Variant_name,
					Price: p.Price,
					DiscountedPrice: "",
					Wishlist: isWishlist,
				}

				if len(p.Product_images) > 0{
					formatted[i].ImageURL = p.Product_images[0].Image_url
				}else{
					formatted[i].ImageURL = ""
				}
			}
		}
		

		
	}

	totalPage := int((total + int64(limit) -1 )/ int64(limit))

	session := sessions.Default(c)
	Name,_ := session.Get("name").(string)

	queryString := "&"+c.Request.URL.Query().Encode()

	if Name != ""{
		c.HTML(http.StatusOK,"product_list.html",gin.H{
			"user":Name,
			"Products": formatted,
			"pagetitle":"product_list",
			"subcategory": responeSubcategory,
			"Page": page,
			"TotalPages": totalPage,
			"QueryString": queryString,
			"sort": sortBy,
		})
		return
	}

	c.HTML(http.StatusOK,"product_list.html",gin.H{
		"Products": formatted,
		"pagetitle":"Product list",
		"Page": page,
		"subcategory": responeSubcategory,
		"TotalPages": totalPage,
		"QueryString": queryString,
		"sort": sortBy,
	})
}