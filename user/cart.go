package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"first-project/models/responsemodels"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var Limit = 5

func AddToCart(c *gin.Context) {

	var product models.Product_Variant
	var cart models.CartItem
	var wishlist models.WishList

	tokenStr, _ := c.Cookie("JWT-User")
	_, id, _ := helper.DecodEJWT(tokenStr)
	quantity, _ := strconv.Atoi(c.PostForm("quantity"))
	productID, _ := strconv.Atoi(c.PostForm("product_id"))
	productIDStr := c.PostForm("product_id")
	session := sessions.Default(c)

	if err := db.Db.
		Preload("Product").
		Preload("Product.SubCategory").
		Preload("Product.SubCategory.Category").
		First(&product, productID).Error; err != nil {

		session.Set("error", "Could not fetch details from db")
		session.Save()
		c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
		return
	}

	if product.Product.SubCategory.IsBlocked || product.Product.SubCategory.Category.IsBlocked || product.Stock < 1 {
		session.Set("error", "Product or category not meets requirement")
		session.Save()
		c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
		return
	}

	if quantity > product.Stock {
		session.Set("error", "Item out of stock")
		session.Save()
		c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
		return
	}

	if quantity > Limit {
		session.Set("error", "User exceeded buying limit")
		session.Save()
		c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
		return
	}

	if err := db.Db.Where("user_id = ? AND product_id = ?", id, productID).First(&cart).Error; err == nil {

		if (cart.Quantity + quantity) > product.Stock {
			session.Set("error", "Item out of stock")
			session.Save()
			c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
			return
		}

		if (cart.Quantity + quantity) > Limit {
			session.Set("error", "User exceeded limit to purchase the item")
			session.Save()
			c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
			return
		}

		cart.Quantity = cart.Quantity + quantity

		if err := db.Db.Save(&cart).Error; err != nil {
			session.Set("error", "Not able to add to cart")
			session.Save()
			c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
			return
		}

		session.Set("flash", "Added to cart")
		session.Save()

		c.Redirect(http.StatusFound, "/user/product/"+productIDStr)

	} else if err == gorm.ErrRecordNotFound {

		newCart := models.CartItem{
			UserID:    uint(id),
			ProductID: uint(productID),
			Price:     product.Price,
			Quantity:  quantity,
			AddAt:     time.Now(),
		}

		if err := db.Db.Create(&newCart).Error; err != nil {
			session.Set("error", "Failed to add item")
			session.Save()
			c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
			return
		}

		if err := db.Db.Where("user_id = ? AND product_id = ?", id, productID).First(&wishlist).Error; err != nil && err != gorm.ErrRecordNotFound {
			session.Set("error", "Failed to load wishlist")
			session.Save()
			c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
			return
		}

		if wishlist.ID != 0 {
			db.Db.Delete(&models.WishList{}, wishlist.ID)
		}

		session.Set("flash", "Added to cart")
		session.Save()
		c.Redirect(http.StatusFound, "/user/product/"+productIDStr)

	} else {
		session.Set("error", "Issue with database")
		session.Save()
		c.Redirect(http.StatusFound, "/user/product/"+productIDStr)
		return
	}

}

func ListCart(c *gin.Context) {

	tokenStr, _ := c.Cookie("JWT-User")
	_, id, _ := helper.DecodeJWT(tokenStr)
	totalItems := 0
	var totalamount float64
	var cart []models.CartItem

	if err := db.Db.Preload("Product").Preload("Product.Product_images").Where("user_id = ?", id).Find(&cart).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User cart not found",
		})
		return
	}

	var responseCart []responsemodels.ResponseCartItem
	today := time.Now().Format("2006-01-02")
	for _, item := range cart {

		var productOffer models.ProductOffer
		var categoryOffer models.CategoryOffer
		var productVariant models.Product_Variant
		var imageUrl string

		if err := db.Db.Preload("Product").Where("id = ?", item.ProductID).First(&productVariant).Error; err != nil {
			log.Println("Could not get product details :", err)
		}

		if err := db.Db.Where("created_at <= ? AND product_id = ? AND active = true ", today+" 23:00:00", productVariant.ProductID).First(&productOffer).Error; err != nil {
			log.Println("No offer available ", err)
		}

		if err := db.Db.Where("created_at <= ? AND category_id = ? AND active = true", today+" 23:00:00", productVariant.Product.SubCategoryID).First(&categoryOffer).Error; err != nil {
			log.Println("No offer available as category ", err)
		}

		for _, image := range item.Product.Product_images {
			if image.Order_no == 1 {
				imageUrl = image.Image_url
				break
			}
		}

		var offername string
		var dicountper float64

		if productOffer.DiscountPercentage > categoryOffer.DiscountPercentage {
			offername = productOffer.OfferName
			dicountper = productOffer.DiscountPercentage
		} else if categoryOffer.DiscountPercentage > productOffer.DiscountPercentage {
			offername = categoryOffer.OfferName
			dicountper = categoryOffer.DiscountPercentage
		} else {
			offername = ""
			dicountper = 0
		}

		temp := responsemodels.ResponseProductVariant{
			ProductID:    productVariant.ID,
			Variant_name: productVariant.Variant_name,
			Image_url:    imageUrl,
			Size:         productVariant.Size,
			Price:        item.Price,
			Quantity:     item.Quantity,
			OfferName:    offername,
			Discount:     dicountper,
		}

		responseCart = append(responseCart, responsemodels.ResponseCartItem{
			ID:      item.ID,
			Product: temp,
		})

		discount := 0.0

		if dicountper > 0 {
			discount = ((item.Price * dicountper) / 100) * (float64(item.Quantity))
		}
		totalItems += item.Quantity
		totalamount += (item.Price*float64(item.Quantity) - discount)
	}

	c.HTML(http.StatusOK, "cart.html", gin.H{"user": "done", "CartItems": responseCart, "TotalItems": totalItems, "TotalAmount": totalamount})

}

func UpdateCartItem(c *gin.Context) {

	productID := c.PostForm("item_id")
	action := c.PostForm("action")
	tokenStr, _ := c.Cookie("JWT-User")
	_, userID, _ := helper.DecodeJWT(tokenStr)

	var cart models.CartItem
	var product models.Product_Variant

	if err := db.Db.Where("id = ?", productID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product details not found"})
		return
	}

	if err := db.Db.Where("user_id = ? AND product_id = ?", userID, productID).First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not able to find cart details"})
		return
	}

	switch action {
	case "inc":
		if cart.Quantity+1 <= product.Stock && cart.Quantity+1 <= Limit {
			cart.Quantity++
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Maximum available quantity reached",
				"quantity": cart.Quantity,
			})
			return
		}
	case "dec":
		if cart.Quantity != 1 {
			cart.Quantity--
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Quantity cannot be less than 1",
				"quantity": cart.Quantity,
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid action",
		})
		return
	}

	if err := db.Db.Save(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"quantity":     cart.Quantity,
		"price":        product.Price,
		"total_items":  getTotalItems(userID),
		"total_amount": getTotalAmount(userID),
	})

}

func RemoveItem(c *gin.Context) {

	ProductID := c.PostForm("item_id")
	var cart models.CartItem

	if err := db.Db.Where("id = ?", ProductID).First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if err := db.Db.Delete(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from cart"})
		return
	}

	c.Redirect(http.StatusFound, "/user/cart")

}

func getTotalItems(userID float64) int {
	var total int
	db.Db.Model(&models.CartItem{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(quantity),0)").
		Scan(&total)

	return total
}

func getTotalAmount(userID float64) float64 {
	var total float64

	var items []models.CartItem

	db.Db.
		Where("user_id = ?", userID).
		Preload("Products").
		Find(&items)

	for _, item := range items {
		total += float64(item.Quantity) * item.Price
	}

	return total
}
