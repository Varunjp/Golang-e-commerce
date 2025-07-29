package user

import (
	db "first-project/DB"
	"first-project/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Review(c *gin.Context) {
	
	orderIDstr := c.PostForm("order_id")
	userIDstr := c.PostForm("user_id")
	productIDstr := c.PostForm("product_id")
	ratingstr := c.PostForm("rating")
	review := c.PostForm("review")

	orderID,orderr := strconv.Atoi(orderIDstr)
	if orderr != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Order id incorrect"})
		return 
	}
	userID, userr := strconv.Atoi(userIDstr)
	if userr != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid user id"})
		return 
	}
	productID,perr := strconv.Atoi(productIDstr)
	if perr != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid product id"})
		return 
	}
	rating,err := strconv.Atoi(ratingstr)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"rating value incorrect"})
		return 
	}

	userReview := models.Review{
		ProductID: uint(productID),
		UserID: uint(userID),
		OrderID: uint(orderID),
		Rating: rating,
		Comment: review,
		CreatedAt: time.Now(),
	}

	if err := db.Db.Create(&userReview).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"please try again later"})
		return 
	}
	
	c.Redirect(http.StatusSeeOther,c.Request.Referer())
}