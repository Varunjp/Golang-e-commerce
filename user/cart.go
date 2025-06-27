package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var Limit = 5

func AddToCart(c *gin.Context){

	var product models.Product_Variant
	var cart models.CartItem
	var wishlist models.WishList
	
	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)
	quantity,_ := strconv.Atoi(c.PostForm("quantity"))
	productID,_ := strconv.Atoi(c.PostForm("product_id"))

	if err := db.Db.Preload("Product").First(&product,productID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Could not fetch details from db"})
		return 
	}


	if product.Product.SubCategory.IsBlocked || product.Product.SubCategory.Category.IsBlocked || product.Stock < 1 {
		c.JSON(http.StatusBadRequest,gin.H{"error":"Product or category not meets requirement"})
		return 
	}

	if quantity > product.Stock {
		c.JSON(http.StatusNotFound,gin.H{"error":"Item out of stock"})
	}

	if quantity > Limit {
		c.JSON(http.StatusConflict,gin.H{"error":"User exceeded buying limit"})
		return 
	}


	if err := db.Db.Where("user_id = ? AND product_id = ?",id,productID).First(&cart).Error; err == nil{
		
		if (cart.Quantity + quantity) > product.Stock {
			c.JSON(http.StatusConflict,gin.H{"error":"Item out of stock"})
			return 
		}

		if (cart.Quantity + quantity) > Limit {
			c.JSON(http.StatusConflict,gin.H{"error":"User exceeded limit to purchase the item"})
			return 
		}

		cart.Quantity = cart.Quantity + quantity

		if err := db.Db.Save(&cart).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Not able to add to cart"})
			return 
		}

		productIDStr:= c.PostForm("product_id")
		c.Redirect(http.StatusFound,"/user/product/"+productIDStr)

	}else if err == gorm.ErrRecordNotFound{

		newCart := models.CartItem{
			UserID: uint(id),
			ProductID: uint(productID),
			Price: product.Price,
			Quantity: quantity,
			AddAt: time.Now(),
		}

		if err := db.Db.Create(&newCart).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to add item"})
			return 
		}

		if err := db.Db.Where("user_id = ? AND product_id = ?",id,productID).First(&wishlist).Error; err != nil && err != gorm.ErrRecordNotFound{
			c.JSON(http.StatusNotFound,gin.H{"error":"Failed to load wishlist"})
			return 
		}

		if wishlist.ID != 0 {
			db.Db.Delete(&models.WishList{},wishlist.ID)
		}

		productIDstr := c.PostForm("product_id")

		c.Redirect(http.StatusFound,"/user/product/"+productIDstr)

	}else {
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Issue with database"})
	}

}

func ListCart(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")
	_,id,_ := helper.DecodeJWT(tokenStr)
	var cart []models.CartItem

	if err := db.Db.Preload("Product").Preload("Product.Product_images").Where("user_id = ?",id).Find(&cart).Error; err != nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"User cart not found",
		})
		return 
	}

	for i, item := range cart{

		for _,image := range item.Product.Product_images{
			if image.Order_no == 1 {
				cart[i].Product.Product_images = []models.Product_image{image}
				break
			}
		}

	}

	

	totalItems := 0
	var totalamount float64

	for _, item := range cart{

		totalItems += item.Quantity
		totalamount += (item.Price * float64(item.Quantity))

	}

	c.HTML(http.StatusOK,"cart.html",gin.H{"user":"done","CartItems":cart,"TotalItems":totalItems,"TotalAmount":totalamount})

}

func UpdateCartItem(c *gin.Context){

	productID := c.PostForm("item_id")
	action := c.PostForm("action")
	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)

	var cart models.CartItem
	var product models.Product_Variant

	if err := db.Db.Where("id = ?",productID).First(&product).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Product details not found"})
		return
	}

	if err := db.Db.Where("user_id = ? AND product_id = ?",userID,productID).First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound,gin.H{"error":"Not able to find cart details"})
		return 
	}

	switch action {
	case "inc":
		if cart.Quantity + 1 <= product.Stock && cart.Quantity + 1 <= Limit {
			cart.Quantity++
			db.Db.Save(&cart)
		}
	case "dec":
		if cart.Quantity == 1 {
			db.Db.Delete(&cart)
			c.Redirect(http.StatusFound,"/user/cart")
			return 
		}else{
			cart.Quantity--
			db.Db.Save(&cart)
		}
	}

	c.Redirect(http.StatusFound,"/user/cart")

}


func RemoveItem(c *gin.Context){


	ProductID := c.PostForm("item_id")
	var cart models.CartItem

	if err := db.Db.Where("id = ?",ProductID).First(&cart).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Item not found"})
		return 
	}

	if err := db.Db.Delete(&cart).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to remove from cart"})
		return 
	}

	c.Redirect(http.StatusFound,"/user/cart")

}