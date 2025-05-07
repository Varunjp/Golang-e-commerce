package routes

import (
	"first-project/admin"
	"first-project/middleware"
	"first-project/user"

	"github.com/gin-gonic/gin"
)

func GetUrl(router *gin.Engine){

	// Welcome
	router.GET("/",user.WelcomePage)
	
	// User
	router.GET("/user/login",user.LoginPage)
	router.POST("/user/login",user.Login)
	router.GET("/user/register",user.RegisterPage)
	router.POST("/user/register",user.RegisterUser)
	router.POST("/verify-otp",user.VerfiyOTP)
	router.GET("/user/home",middleware.AuthUserMiddlerware("user"),user.HomePage)
	router.GET("/user/logout",user.UserLogout)



	// Demo
	router.GET("/demo",user.DemoPage)

	//Admin
	router.GET("/admin",middleware.AuthMiddlerware("admin"),admin.LoginPage)
	router.GET("/admin/login",admin.LoginPage)
	router.POST("/admin/login",admin.Login)
	
	// Admin users
	router.GET("/admin/users-list",middleware.AuthMiddlerware("admin"),admin.ListUsers)
	router.GET("/admin/users",middleware.AuthMiddlerware("admin"),admin.FindUser)
	router.GET("/admin/users/block/:id",middleware.AuthMiddlerware("admin"),admin.BlockUser)
	router.GET("/admin/users/unblock/:id",middleware.AuthMiddlerware("admin"),admin.UnblockUser)

	// Admin categories
	router.GET("/admin/categories",middleware.AuthMiddlerware("admin"),admin.ViewCategory)
	router.GET("/admin/categories/edit/:id",middleware.AuthMiddlerware("admin"),admin.EditCategoryPage)
	router.POST("/admin/categories/add",middleware.AuthMiddlerware("admin"),admin.AddCategory)
	router.GET("/admin/categories/delete/:id",middleware.AuthMiddlerware("admin"),admin.DeleteCategory)
	router.POST("/admin/categories/subcategories/add/:id",middleware.AuthMiddlerware("admin"),admin.AddSubCategory)
	router.GET("/admin/subcategories/edit/:id",middleware.AuthMiddlerware("admin"),admin.EditSubCategoryPage)
	router.POST("/admin/subcategories/update/:id",middleware.AuthMiddlerware("admin"),admin.UpdateSubCategory)
	router.POST("/admin/subcategories/delete/:id",middleware.AuthMiddlerware("admin"),admin.DeleteSubCategory)

	// Admin Products
	router.GET("/admin/products",middleware.AuthMiddlerware("admin"),admin.ViewProducts)
	router.GET("/admin/product/create",middleware.AuthMiddlerware("admin"),admin.AddProductPage)
	router.POST("/admin/products/create",middleware.AuthMiddlerware("admin"),admin.AddProduct)
	router.GET("/admin/products/edit/:id",middleware.AuthMiddlerware("admin"),admin.UpdateProductPage)
	router.POST("/admin/products/edit/:id",middleware.AuthMiddlerware("admin"),admin.UpdateProduct)
	router.POST("/admin/products/delete/:id",middleware.AuthMiddlerware("admin"),admin.DeleteProduct)
	

}