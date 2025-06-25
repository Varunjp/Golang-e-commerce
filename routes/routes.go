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
	router.GET("/user/login",middleware.NoCacheMiddleware(),user.LoginPage)
	router.POST("/user/login",middleware.NoCacheMiddleware(),user.Login)
	router.GET("/user/register",middleware.NoCacheMiddleware(),user.RegisterPage)
	router.POST("/user/register",middleware.NoCacheMiddleware(),user.RegisterUser)
	router.POST("/verify-otp",middleware.NoCacheMiddleware(),user.VerfiyOTP)
	router.GET("/",user.HomePage)
	router.GET("/user/logout",middleware.NoCacheMiddleware(),user.UserLogout)
	router.GET("auth/google/login",middleware.NoCacheMiddleware(),user.HandleGoogleLogin)
	router.GET("/auth/google/callback",middleware.NoCacheMiddleware(),user.HandleGoogleCallback)
	router.POST("/user/resend-otp",middleware.NoCacheMiddleware(),user.ResendOTP)
	router.GET("/user/forgot-password",middleware.NoCacheMiddleware(),user.ResetPasswordOTP)
	router.POST("/user/forgot-password",middleware.NoCacheMiddleware(),user.ResetPasswordOTPSend)
	router.GET("/reset-password/verify-otp",middleware.NoCacheMiddleware(),user.ResetPasswordOTPpage)
	router.POST("/reset-password/verify-otp",middleware.NoCacheMiddleware(),user.ResetPasswordOTPVerify)
	router.POST("/reset-password/resend-otp",middleware.NoCacheMiddleware(),user.Resetpassword_ResendOTP)
	router.GET("/user/reset-password",middleware.NoCacheMiddleware(),user.ShowRestPasswordPage)
	router.POST("/user/reset-password",middleware.NoCacheMiddleware(),user.ResetPassword)

	// User profile page
	router.GET("/user/profile",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.UserProfilePage)
	router.GET("/user/edit-profile",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.EditProfilePage)
	router.POST("/user/update-profile",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.UpdateProfile)
	router.POST("/user/add-address,middleware.AuthVaildUser()",middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.AddAddress)
	router.POST("/user/edit-address",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.EditAddress)
	router.POST("/delete-address",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.DeleteAddress)
	router.GET("/user/change-password",middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.ChangePasswordPage)
	router.POST("/user/change-password",middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.ChangePassword)
	router.POST("/user/upload-profile-image",middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.UploadProfileImage)
	router.POST("/user/verify-email-otp",middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.UpdateEmail)

	// User referral
	router.GET("/user/create/referral",middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.GenerateReferralCode)
	router.POST("/user/referral",middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.SubmitReferralCode)
	router.GET("/user/referral",middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.SubmitReferralPage)

	// User product page
	router.GET("/user/shop",middleware.NoCacheMiddleware(),user.ShowProductList)
	router.GET("/user/product/:id",middleware.NoCacheMiddleware(),user.Product)

	// User orders
	router.POST("/cart/add",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),middleware.VerifyProduct(),user.AddToCart)
	router.GET("/user/cart",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),middleware.VerifyProduct(),user.ListCart)
	router.POST("/cart/update-quantity",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),middleware.VerifyProduct(),user.UpdateCartItem)
	router.POST("/cart/remove",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),middleware.VerifyProduct(),user.RemoveItem)
	router.GET("/user/orders",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),middleware.VerifyProduct(),user.ListOrders)
	router.POST("/user/cancel-order",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.ReturnOrder)
	router.GET("/user/order/:id",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.OrderItems)
	router.POST("/user/cancel-item",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.CancelItem)
	router.POST("/order/failed",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.OrderFailed)
	router.GET("/order/failed-page/:id",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.OrderFailedPage)


	// User checkout
	router.GET("/user/checkout",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.CheckOutPage)
	router.POST("/place-order",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.CheckOutOrder)
	router.GET("/user/add-address",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.AddNewAddressPage)
	router.POST("/user/save-address",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.AddNewAddress)
	router.GET("/user/invoice/:id",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.DownloadPdf)
	router.GET("/order/confirmation/:id",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.OrderConfirmation)

	// User wallet
	router.GET("/user/wallet-transactions",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.WalletTransaction)

	// User payment online
	router.POST("/create-razorpay-order",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),middleware.VerifyProduct(),user.CreateRazorpayOrder)
	router.POST("/payment/success",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),user.PaymentSuccess)

	// User wishlist
	router.GET("/user/add-wishlist/:id",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.AddToWishlist)
	router.GET("/user/remove-wishlist/:id",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.RemoveWishlist)
	router.GET("/user/wishlist",middleware.AuthVaildUser(),middleware.NoCacheMiddleware(),middleware.AuthUserMiddlerware("user"),user.WishlistPage)

	// Demo
	router.GET("/demo",user.DemoPage)



	
	//Admin
	router.GET("/admin",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.LoginPage)
	router.POST("/admin",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.LoginPage)
	router.GET("/admin/login",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.LoginPage)
	router.POST("/admin/login",middleware.NoCacheMiddleware(),admin.Login)
	router.GET("/admin/logout",middleware.NoCacheMiddleware(),admin.Logout)
	router.GET("/admin/sales-data",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.GetSalesData)
	
	// Admin users
	router.GET("/admin/users-list",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.ListUsers)
	router.GET("/admin/users",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.FindUser)
	router.GET("/admin/users/block/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.BlockUser)
	router.GET("/admin/users/unblock/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.UnblockUser)
	router.GET("/admin/users/delete/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.DeleteUser)

	// Admin categories
	router.GET("/admin/categories",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.ViewCategory)
	router.GET("/admin/categories/edit/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.EditCategoryPage)
	router.POST("/admin/categories/:id/update",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.EditCategory)
	router.POST("/admin/categories/add",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AddCategory)
	router.GET("/admin/categories/delete/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.DeleteCategory)
	router.POST("/admin/categories/subcategories/add/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AddSubCategory)
	router.GET("/admin/subcategories/edit/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.EditSubCategoryPage)
	router.POST("/admin/subcategories/update/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.UpdateSubCategory)
	router.POST("/admin/subcategories/delete/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.DeleteSubCategory)

	// Admin Products
	router.GET("/admin/products",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.ViewProducts)
	router.GET("/admin/product/create",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AddProductPage)
	router.POST("/admin/products/create",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AddProduct)
	router.GET("/admin/products/edit/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.UpdateProductPage)
	router.POST("/admin/products/edit/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.UpdateProduct)
	router.GET("/admin/products/delete/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.DeleteProduct)
	router.POST("/admin/products/images/delete/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.DeleteImage)
	router.GET("/admin/product/variant",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AddProductVariantPage)
	router.POST("/admin/variants/create",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AddProductVariant)
	
	// Admin Orders
	router.GET("/admin/orders",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AdminOrdersPage)
	router.POST("/admin/orders/cancel/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AdminOrderCancel)
	router.GET("/admin/order/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AdminOrderDetails)
	router.POST("/admin/orders/update-status/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AdminOrderUpdate)
	router.GET("/admin/order/item/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AdminItemOrder)
	router.GET("/admin/order/item-reject/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AdminItemCancel)
	router.GET("/admin/order/return-request",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AdminOrderReturnRequests)
	router.GET("/admin/order/item-admin-reject/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AdminSideItemCancel)
	
	// Admin excel
	router.POST("/admin/reports/excel",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.DownloadExcel)


	// Admin banner
	router.GET("/admin/banners",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.BannerPage)
	router.POST("/admin/banners/add",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AddBanner)
	router.POST("/admin/banners/delete/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.DeleteBanner)

	// Admin coupons
	router.GET("/admin/coupons",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.ListCoupons)
	router.POST("/admin/coupons/add",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.AddCoupon)
	router.GET("/admin/coupons/toggle/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.ToggleCoupon)
	router.GET("admin/coupons/delete/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.DeleteCoupon)
	router.POST("/admin/coupons/update/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.EditCoupon)
	router.GET("/admin/coupon/edit/:id",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.EditCouponPage)

	// Admin wallet
	router.GET("/admin/wallet-transactions",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.WalletTransactions)
	router.GET("/admin/refund-requests",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.WalletRefunds)
	router.POST("/admin/refund/approve",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.WalletRefundApproval)
	router.POST("/admin/refund/decline",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.WalletRefundDecline)

	// Admin reports
	router.GET("/admin/reports",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.SalesReportPage)
	router.POST("/admin/reports/download",middleware.NoCacheMiddleware(),middleware.AuthMiddlerware("admin"),admin.DownloadSalesReport)
	
}