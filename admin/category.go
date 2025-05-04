package admin

import (
	db "first-project/DB"
	"first-project/models"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ViewCategory (c *gin.Context){
	
	var category []models.Category
	
	// result := db.Db.Raw(
	// 	`SELECT category_id,category_name
	// 	FROM categories
	// 	WHERE deleted_at IS NULL
	// 	ORDER BY category_id DESC`).Scan(&category)
	
	// if result.Error != nil{
	// 	c.JSON(http.StatusInternalServerError,gin.H{"error": result.Error.Error()})
	// 	return 
	// }

	// if len(category) == 0{
	// 	c.JSON(http.StatusOK,gin.H{"message":"No categories listed"})
	// 	return
	// }

	if err := db.Db.Find(&models.Category{}).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to retrieve categories"})
		return 
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

	var total int64

	keyword := c.Query("search")

	//delete
	fmt.Println(keyword)

	if keyword == ""{
		adDB := db.Db.Model(&models.Category{})

		adDB.Count(&total)

		adDB.Where("deleted_at IS NULL").Order("category_id DESC").Limit(limit).Offset(offset).Find(&category)

		totalPage := int(math.Ceil(float64(total)/float64(limit)))

		//c.JSON(http.StatusOK,gin.H{"categories": category})

		c.HTML(http.StatusOK,"category_list.html",gin.H{
			"categoriesList":category,
			"page":page,
			"limit":limit,
			"totalPages":totalPage,
		})
	}else{

		adDb := db.Db.Where("LOWER(category_name) LIKE ?", "%"+strings.ToLower(keyword)+"%").Order("category_id desc").Limit(limit).Offset(offset).Find(&category)

		if err := adDb.Error; err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"Database error"})
			return 
		}
		
		adDb.Count(&total)

		if total < 1 {
			c.JSON(http.StatusNotFound,gin.H{"category":"Not found"})
			return 
		}

		totalPages := int(math.Ceil(float64(total)/ float64(limit)))


		c.HTML(http.StatusOK,"category_list.html",gin.H{
			"categoriesList":category,
			"page":page,
			"limit":limit,
			"totalPages":totalPages,
		})
	}
	
	
}

func FindCategory (c *gin.Context){
	
	var category []models.Category
	

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

	if keyword != ""{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid data passed"})
		return
	}

	adDb := db.Db.Where("LOWER(category_name) LIKE ?", "%"+strings.ToLower(keyword)+"%").Order("category_id desc").Limit(limit).Offset(offset).Find(&category)

	if err := adDb.Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Database error"})
		return 
	}
	
	adDb.Count(&total)

	if total < 1 {
		c.JSON(http.StatusNotFound,gin.H{"category":"Not found"})
		return 
	}

	totalPages := int(math.Ceil(float64(total)/ float64(limit)))


	c.HTML(http.StatusOK,"category_list.html",gin.H{
		"categoriesList":category,
		"page":page,
		"limit":limit,
		"totalPages":totalPages,
	})

}

func AddCategory (c *gin.Context){
	
	categoryName := c.PostForm("name")

	if categoryName == ""{
		// check status code
		c.Redirect(http.StatusNotFound,"/admin/categories")
		return
	}

	var existingCategory models.Category
	if err := db.Db.Where("category_name = ?",categoryName).First(&existingCategory).Error; err == nil {
		c.JSON(http.StatusConflict,gin.H{"error":"Category already exists"})
		return 
	}

	var newCategory models.Category
	newCategory.CategoryName = categoryName

	if err := db.Db.Create(&newCategory).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Could not create category"})
		return 
	}

	c.JSON(http.StatusCreated,gin.H{"Category created successfully": newCategory})

}

func EditCategory(c *gin.Context){
	categoryID := c.Param("id")
	newName := c.PostForm("name")
	var category models.Category

	if err := db.Db.First(&category, categoryID).Error;err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Category not found"})
		return 
	}

	id,_ := strconv.Atoi(categoryID) 
	category.CategoryID = uint(id)
	category.CategoryName = newName

	if err := db.Db.Save(&category).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to update category"})
		return 
	}

	c.JSON(http.StatusOK,gin.H{"Category updated successfully": newName})
}

func DeleteCategory(c *gin.Context){
	categoryID := c.Param("id")
	var category models.Category

	if err := db.Db.Where("category_id = ?",categoryID).First(&category).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"category not found"})
		return 
	}

	if err := db.Db.Model(&category).Update("deleted_at",time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed delete category"})
		return 
	}

	c.JSON(http.StatusOK,gin.H{"message":"Category deleted successfully"})
	
}