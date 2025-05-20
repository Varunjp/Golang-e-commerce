package routes

import (
	"first-project/admin"
	"first-project/middleware"
	"first-project/user"

	"github.com/gin-gonic/gin"
)

func GetUrl(router *gin.Engine){

	// Welcome
	//router.GET("/",user.WelcomePage)
	
	// User
	router.GET("/user/login",user.LoginPage)
	router.POST("/user/login",user.Login)
	router.GET("/user/register",user.RegisterPage)
	router.POST("/user/register",user.RegisterUser)
	router.POST("/verify-otp",user.VerfiyOTP)
	router.GET("/",user.HomePage)
	router.GET("/user/logout",user.UserLogout)
	router.GET("auth/google/login",user.HandleGoogleLogin)
	router.GET("/auth/google/callback",user.HandleGoogleCallback)
	router.POST("/user/resend-otp",user.ResendOTP)
	router.GET("/user/forgot-password",user.ResetPasswordOTP)
	router.POST("/user/forgot-password",user.ResetPasswordOTPSend)
	router.GET("/reset-password/verify-otp",user.ResetPasswordOTPpage)
	router.POST("/reset-password/verify-otp",user.ResetPasswordOTPVerify)
	router.POST("/reset-password/resend-otp",user.Resetpassword_ResendOTP)
	router.GET("/user/reset-password",user.ShowRestPasswordPage)
	router.POST("/user/reset-password",user.ResetPassword)

	// User profile page
	router.GET("/user/profile",middleware.AuthUserMiddlerware("user"),user.UserProfilePage)
	router.GET("/user/edit-profile",middleware.AuthUserMiddlerware("user"),user.EditProfilePage)
	router.POST("/user/update-profile",middleware.AuthUserMiddlerware("user"),user.UpdateProfile)
	router.POST("/user/add-address",middleware.AuthUserMiddlerware("user"),user.AddAddress)
	router.POST("/user/edit-address",middleware.AuthUserMiddlerware("user"),user.EditAddress)
	router.GET("/user/change-password",middleware.AuthUserMiddlerware("user"),user.ChangePasswordPage)
	router.POST("/user/change-password",middleware.AuthUserMiddlerware("user"),user.ChangePassword)
	router.POST("/user/upload-profile-image",middleware.AuthUserMiddlerware("user"),user.UploadProfileImage)
	router.POST("/user/verify-email-otp",middleware.AuthUserMiddlerware("user"),user.UpdateEmail)

	// User product page
	router.GET("/user/shop",user.ShowProductList)
	router.GET("/user/product/:id",user.Product)

	// User orders
	router.POST("/cart/add",middleware.AuthUserMiddlerware("user"),user.AddToCart)
	router.GET("/user/cart",middleware.AuthUserMiddlerware("user"),user.ListCart)
	router.POST("/cart/update-quantity",middleware.AuthUserMiddlerware("user"),user.UpdateCartItem)
	router.POST("/cart/remove",middleware.AuthUserMiddlerware("user"),user.RemoveItem)
	router.GET("/user/orders",middleware.AuthUserMiddlerware("user"),user.ListOrders)
	router.POST("/user/cancel-order",middleware.AuthUserMiddlerware("user"),user.ReturnOrder)
	router.GET("/user/order/:id",middleware.AuthUserMiddlerware("user"),user.OrderItems)


	// User checkout
	router.GET("/user/checkout",middleware.AuthUserMiddlerware("user"),user.CheckOutPage)
	router.POST("/place-order",middleware.AuthUserMiddlerware("user"),user.CheckOutOrder)
	router.GET("/user/add-address",middleware.AuthUserMiddlerware("user"),user.AddNewAddressPage)
	router.POST("/user/save-address",middleware.AuthUserMiddlerware("user"),user.AddNewAddress)
	router.GET("/user/invoice/:id",middleware.AuthUserMiddlerware("user"),user.DownloadPdf)

	// Demo
	router.GET("/demo",user.DemoPage)



	
	//Admin
	router.GET("/admin",middleware.AuthMiddlerware("admin"),admin.LoginPage)
	router.POST("/admin",middleware.AuthMiddlerware("admin"),admin.LoginPage)
	router.GET("/admin/login",middleware.AuthMiddlerware("admin"),admin.LoginPage)
	router.POST("/admin/login",admin.Login)
	router.GET("/admin/logout",admin.Logout)
	
	// Admin users
	router.GET("/admin/users-list",middleware.AuthMiddlerware("admin"),admin.ListUsers)
	router.GET("/admin/users",middleware.AuthMiddlerware("admin"),admin.FindUser)
	router.GET("/admin/users/block/:id",middleware.AuthMiddlerware("admin"),admin.BlockUser)
	router.GET("/admin/users/unblock/:id",middleware.AuthMiddlerware("admin"),admin.UnblockUser)

	// Admin categories
	router.GET("/admin/categories",middleware.AuthMiddlerware("admin"),admin.ViewCategory)
	router.GET("/admin/categories/edit/:id",middleware.AuthMiddlerware("admin"),admin.EditCategoryPage)
	router.POST("/admin/categories/:id/update",middleware.AuthMiddlerware("admin"),admin.EditCategory)
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
	router.POST("/admin/products/images/delete/:id",middleware.AuthMiddlerware("admin"),admin.DeleteImage)
	
	// Admin Orders
	router.GET("/admin/orders",middleware.AuthMiddlerware("admin"),admin.AdminOrdersPage)
}