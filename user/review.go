package user

import (
	db "first-project/DB"
	"first-project/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Review(c *gin.Context) {
	
	orderIDstr := c.PostForm("order_id")
	userIDstr := c.PostForm("user_id")
	productIDstr := c.PostForm("product_id")
	ratingstr := c.PostForm("rating")
	review := c.PostForm("review")
	Errors := make(map[string]string)

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
	if strings.TrimSpace(review) == ""{
		Errors["review"] = "Please provide review about product"
	}
	rating,err := strconv.Atoi(ratingstr)
	if err != nil{
		Errors["rating"] = "Please provide any value"
		c.JSON(http.StatusBadRequest,gin.H{"error":"rating value incorrect"})
		return 
	}

	if rating < 1 || rating > 5 {
		Errors["rating"] = "Please provide rating between 1 - 5"
		c.JSON(http.StatusBadRequest,gin.H{"error":"Please provide rating between 1 - 5"})
		return 
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest,gin.H{"status":"error","errors":Errors})
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

	c.JSON(http.StatusOK,gin.H{"redirect":c.Request.Referer()})
}

func ReviewEditPage(c *gin.Context){
	reviewID := c.Param("id")
	var review models.Review

	if err := db.Db.Where("id = ?",reviewID).First(&review).Error; err != nil{
		c.Redirect(http.StatusSeeOther,c.Request.Referer())
		return
	}

	redirect := c.Request.Referer()

	c.HTML(http.StatusOK,"edit_review.html",gin.H{"Review":review,"user":"done","Redirect":redirect})
}

func ReviewEdit(c *gin.Context){
	reviewID := c.Param("id")
	var review models.Review
	Errors := make(map[string]string)

	comment := c.PostForm("comment")
	ratingstr := c.PostForm("rating")
	redirect := c.PostForm("redirect")

	rating,err := strconv.Atoi(ratingstr)

	if err != nil || rating < 1 || rating > 5{
		Errors["rating"] = "Please provide correct rating"
	}

	if strings.TrimSpace(comment) == ""{
		Errors["comment"] = "Please provide correc comment"
	}

	if len(Errors) > 0 {
		c.JSON(http.StatusBadRequest,gin.H{"status":"error","errors":Errors})
		return 
	}

	if err := db.Db.Where("id = ?",reviewID).First(&review).Error; err != nil{
		c.Redirect(http.StatusSeeOther,c.Request.Referer())
		return 
	}

	review.Rating = rating
	review.Comment = comment

	if err := db.Db.Save(&review).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to save review"})
		return 
	}

	c.JSON(http.StatusOK,gin.H{"redirect":redirect})
}