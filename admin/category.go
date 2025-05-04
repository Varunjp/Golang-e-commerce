package admin

import (
	db "first-project/DB"
	"first-project/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ViewCategory (c *gin.Context){
	var category []models.Category
	
	result := db.Db.Raw(
		`SELECT category_id,category_name
		FROM categories
		WHERE deleted_at IS NULL`).Scan(&category)
	
	if result.Error != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error": result.Error.Error()})
		return 
	}

	if len(category) == 0{
		c.JSON(http.StatusOK,gin.H{"message":"No categories listed"})
		return
	}

	if err := db.Db.Find(&category).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to retrieve categories"})
		return 
	}

	c.JSON(http.StatusOK,gin.H{"categories": category})
}