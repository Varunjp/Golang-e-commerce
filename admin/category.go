package admin

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ViewCategory (c *gin.Context){
	
	var category []models.Category

	session := sessions.Default(c)
	name := session.Get("name").(string)
	
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

	if keyword == ""{
		adDB := db.Db.Model(&models.Category{})

		adDB.Count(&total)

		adDB.Where("deleted_at IS NULL").Order("category_id DESC").Limit(limit).Offset(offset).Find(&category)

		totalPage := int(math.Ceil(float64(total)/float64(limit)))

		//c.JSON(http.StatusOK,gin.H{"categories": category})
		

		c.HTML(http.StatusOK,"category_list.html",gin.H{
			"user":name,
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
			c.HTML(http.StatusNotFound,"category_list.html",gin.H{"error": "Category not found"})
			return 
		}

		totalPages := int(math.Ceil(float64(total)/ float64(limit)))

		c.HTML(http.StatusOK,"category_list.html",gin.H{
			"user":name,
			"categoriesList":category,
			"page":page,
			"limit":limit,
			"totalPages":totalPages,
		})
	}
	
	
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
		c.HTML(http.StatusConflict,"category_list.html",gin.H{"error":"Category already exists"})
		return 
	}

	var newCategory models.Category
	newCategory.CategoryName = categoryName

	if err := db.Db.Create(&newCategory).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"category_list.html",gin.H{"error":"Could not create category"})
		return 
	}

	c.Redirect(http.StatusFound,"/admin/categories")

}

func EditCategoryPage(c *gin.Context){
	
	categoryID,_ := strconv.Atoi(c.Param("id"))

	var category models.Category

	if err := db.Db.First(&category,categoryID).Error; err != nil{
		c.Redirect(http.StatusNotFound,"/admin/categories")
		return 
	}

	var subCategories []models.SubCategory

	if err := db.Db.Where("category_id = ?",categoryID).Find(&subCategories).Error; err != nil{
		c.Redirect(http.StatusNotFound,"/admin/categories")
		return 
	}

	c.HTML(http.StatusOK,"edit_category.html",gin.H{"Category":category,"Subcategories":subCategories})


}

func AddSubCategory(c *gin.Context){
	
	categoryIDStr := c.Param("id")
	categoryID,_ := strconv.Atoi(c.Param("id"))
	
	newName := c.PostForm("name")

	var category models.Category
	var subCategory models.SubCategory

	if err := db.Db.First(&category,categoryID).Error; err != nil{
		c.Redirect(http.StatusSeeOther,"/admin/categories")
		return 
	}

	if err := db.Db.Where("sub_category_name = ? AND  category_id = ?",newName,categoryID).First(&subCategory).Error; err == nil{

		log.Println("Category already exists")
		c.Redirect(http.StatusSeeOther,"/admin/categories")
		return 
	}

	NewsubCategory := models.SubCategory{
		CategoryID: uint(categoryID),
		SubCategoryName: newName,
	}

	if err := db.Db.Create(&NewsubCategory).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to create new sub category"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/admin/categories/edit/"+categoryIDStr)
	
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

	c.Redirect(http.StatusFound,"/admin/categories")
}

func EditSubCategoryPage(c *gin.Context){

	subCategoryID := c.Param("id")
	var subCategory models.SubCategory

	if err := db.Db.First(&subCategory,subCategoryID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Category not found"})
		return 
	}

	c.HTML(http.StatusOK,"edit_subCategory.html",gin.H{"Subcategory":subCategory})

}

func UpdateSubCategory(c *gin.Context){

	subCategoryID := c.Param("id")
	newName := c.PostForm("name")

	var subCategory models.SubCategory

	if err := db.Db.First(&subCategory,subCategoryID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Category not found"})
		return
	}

	subCategory.SubCategoryName = newName

	if err := db.Db.Save(&subCategory).Error; err != nil{
		c.String(http.StatusInternalServerError,"Error while updating subcategory")
		return
	}

	c.Redirect(http.StatusFound,"/admin/categories")
}

func DeleteCategory(c *gin.Context){
	
	categoryID := c.Param("id")
	var category models.Category

	if err := db.Db.Where("category_id = ?",categoryID).First(&category).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"category_list.html",gin.H{"error":"Failed delete category"})
		return 
	}
	
	var err error

	if category.IsBlocked{
		err = helper.UpdateAllUnderCategory(categoryID)
	}else{
		err = helper.DeleteAllUnderCategory(categoryID)
	}

	

	if err != nil {
		c.HTML(http.StatusInternalServerError,"category_list.html",gin.H{"error":"Failed delete category"})
		return
	}

	c.Redirect(http.StatusFound,"/admin/categories")
}

func DeleteSubCategory(c *gin.Context){

	subCategoryID := c.Param("id")


	err := helper.DeleteAllUnderSubCategory(subCategoryID)

	if err != nil{
		c.HTML(http.StatusInternalServerError,"edit_category.html",gin.H{
			"error":err.Error(),
		})
		return 
	}

	c.Redirect(http.StatusFound,"/admin/categories")

}