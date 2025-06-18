package user

import (
	db "first-project/DB"
	"first-project/helper"
	"first-project/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CheckOutPage(c *gin.Context) {

	var CartItems []models.CartItem
	var Addresses []models.Address
	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)
	var coupons []models.Coupons
	var usedCoupon []models.UsedCoupon

	if err := db.Db.Preload("Product").Where("user_id = ?",userID).Find(&CartItems).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to load data from DB"})
		return 
	}

	if err := db.Db.Where("user_id = ?",userID).Find(&usedCoupon).Error; err != nil{
		if err != gorm.ErrRecordNotFound {
			c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"failed to load coupon details"})
			return 
		}
	}

	usedCouponId := make([]uint,len(usedCoupon))

	if len(usedCoupon) > 0 {
		for i,coupon := range usedCoupon{
			usedCouponId[i] = coupon.CouponID
		}
	}

	if err := db.Db.Where("user_id = ?",userID).Find(&Addresses).Error; err != nil{

		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Not able to fetch address from db"})
			return
		}
	
	}

	// product name, imageurl, total sum, total tax, applicable discount 

	var Response []struct {
		ID 				uint 
		Name 			string
		Quantity 		int
		Price  			float64 
		TotalSum		float64
		TotalTax		float64
		TotalDiscount	float64
		GrandTotal		float64
	}

	subCatID := make([]int,0)

	for _,item := range CartItems{

		res := helper.ValidateProduct(item.ProductID,item.Quantity)

		if res {

			Response = append(Response, struct{ID uint; Name string; Quantity int;Price float64; TotalSum float64; TotalTax float64; TotalDiscount float64; GrandTotal float64}{
				ID: item.ProductID,
				Name: item.Product.Variant_name,
				Quantity: item.Quantity,
				Price: item.Price,
				TotalSum: (item.Price * float64(item.Quantity)),
				TotalTax: (item.Product.Tax*float64(item.Quantity)),
				TotalDiscount: 0.0,
				GrandTotal: (item.Price * float64(item.Quantity))+(item.Product.Tax*float64(item.Quantity)),
			})

			var tempProduct models.Product_Variant
			db.Db.Preload("Product").Where("id = ?",item.ProductID).First(&tempProduct)
			subCatID = append(subCatID, int(tempProduct.Product.SubCategoryID))
		}

	}

	dbcoupons := db.Db.Model(&models.Coupons{}).Where("is_active = ?",true)
	dbcoupons = dbcoupons.Where("category_id IN ? OR category_id is NULL OR category_id = 0",subCatID)
	dbcoupons = dbcoupons.Where("user_id = ? OR user_id is NULL OR user_id = 0",userID)
	dbcoupons = dbcoupons.Not("type = ?","Base")

	if len(usedCoupon) != 0 {
		if err := dbcoupons.Where("id NOT IN ?",usedCouponId).Find(&coupons).Error; err != nil{
		log.Println("Error while loading coupons :",err)
		}
	}else{
		if err := dbcoupons.Find(&coupons).Error; err != nil{
		log.Println("Error while loading coupons :",err)
		}
	}


	var totalamount float64

	for _, item := range Response{
		totalamount += item.GrandTotal
	}

	var wallet models.Wallet

	if err := db.Db.Where("user_id = ?",userID).First(&wallet).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			errCreate := helper.CreateWallet(uint(userID))
			if errCreate == nil{
				db.Db.Where("user_id = ?",userID).First(&wallet)
			}else{
				c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Failed to load wallet details, please try again later"})
				return
			}
		}else{
			c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Failed to load wallet details, please try again later"})
			return 
		}
	}

	c.HTML(http.StatusOK,"checkOut.html",gin.H{"user":"done","CartItems":Response,"Addresses":Addresses,"TotalAmount":totalamount,"Coupons":coupons,"Balance":wallet.Balance})

}

func CheckOutOrder(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)
	addressOption := c.PostForm("address_id")
	paymentOption := c.PostForm("payment_method")
	couponCode := c.PostForm("coupon_code")
	isWallet := c.PostForm("use_wallet")

	var addressID uint 
	var orderitems []models.OrderItem
	var usedcouponcheck models.UsedCoupon
	
	if addressOption == ""{
		c.HTML(http.StatusBadRequest,"checkOut.html",gin.H{"error":"Need to provide a address details"})
		return 
	}

	if err := db.Db.Where("user_id = ? AND coupon_id = ?",userID,couponCode).First(&usedcouponcheck).Error; err == nil{

		c.HTML(http.StatusBadRequest,"checkOut.html",gin.H{"error":"Coupon already used"})
		return 
	}

	// if paymentOption != "cod"{
	// 	c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Online option not available yet please change payment method to cod"})
	// 	return
	// }

	if err := db.Db.Where("user_id = ? AND deleted_at IS NULL",userID).Find(&orderitems).Error; err != nil{
		if err != gorm.ErrRecordNotFound{
			c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Failed to load user details please try again later"})
			return 
		}
	}

	if addressOption == "new" {
		newAddress := models.Address{
			UserID: uint(userID),
			AddressLine1: c.PostForm("line1"),
			AddressLine2: c.PostForm("line2"),
			Country: c.PostForm("country"),
			State: c.PostForm("state"),
			PostalCode: c.PostForm("postalCode"),
		}

		if err := db.Db.Create(&newAddress).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to save address"})
			return 
		}

		addressID = newAddress.AddressID
	}else {

		id,_ := strconv.Atoi(addressOption)
		addressID = uint(id)

	}

	var CartItems []models.CartItem

	if err := db.Db.Where("user_id = ?",userID).Find(&CartItems).Error; err != nil {
		
		c.JSON(http.StatusNotFound,gin.H{"error":"Not able to load cart items"})
		return 
	}

	if len(CartItems) == 0 {
		c.JSON(http.StatusBadRequest,gin.H{"error":"Cart is empty"})
		return 
	}

	var total float64

	for _,item := range CartItems{
		var product models.Product_Variant
		total += item.Price * float64(item.Quantity)
		db.Db.Where("id = ?",item.ProductID).First(&product)
		tax := product.Tax * float64(item.Quantity)
		total +=tax
	}

	var coupon models.Coupons
	var discount float64

	if couponCode != ""{

		if err := db.Db.Where("id = ?",couponCode).First(&coupon).Error; err != nil{
			log.Println(err)
			if err != gorm.ErrRecordNotFound{
				c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to load coupon details"})
				return 
			}
			
		}

		
		if coupon.ID != 0 {
			discount = (total * coupon.Discount)/100

			if discount > coupon.MaxAmount {
				discount = coupon.MaxAmount
			}
		}


	}

	var finalAmount	float64
	var walletUsed float64

	if isWallet == "on"{

		var wallet models.Wallet

		if err := db.Db.Where("user_id = ?",userID).First(&wallet).Error; err != nil{
			if err != gorm.ErrRecordNotFound{
				c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Could not load wallet balance, please try again later"})
				return 
			}
		}

		orderTotal := total - discount
		walletUsed = math.Min(wallet.Balance,orderTotal)
		finalAmount = orderTotal - walletUsed

	}else{
		finalAmount = total - discount
	}

	order := models.Order{
		UserID: uint(userID),
		AddressID: addressID,
		TotalAmount: finalAmount,
		SubTotal: total,
		DiscountTotal: discount+walletUsed,
		Status: "Processing",
		PaymentMethod: paymentOption,
		PaymentStatus: "Pending",
		CreateAt: time.Now(),
	}

	if err := db.Db.Create(&order).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to create order"})
		return 
	}

	if isWallet == "on" {
		if err := helper.DebitWallet(uint(userID),walletUsed,order.ID,"Purchase order :"+strconv.Itoa(int(order.ID))); err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to update wallet"})
			return 
		}
	}

	

	if couponCode != ""{
		var coupon models.Coupons

		if err := db.Db.Where("id = ?",couponCode).First(&coupon).Error; err != nil{
			log.Println(err)
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to load coupon details"})
			return 
		}
		
		usedcoupon := models.UsedCoupon{
			UserID: order.UserID,
			CouponID: coupon.ID,
			OrderID: order.ID,
		}

		if err := db.Db.Create(&usedcoupon).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error while saving coupon upadate please try again later."})
			return 
		}

	}

	for _,item := range CartItems{

		itemCount := 0

		if err := db.Db.Model(&models.Product_Variant{}).Where("id = ? AND stock >= ?",item.ProductID,item.Quantity).Update("stock",gorm.Expr("stock - ?",item.Quantity)).Error; err != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":"Insufficient stock"})
			return 
		}

		for _, oritems := range orderitems{
			if item.ProductID == oritems.ProductID{
				itemCount = item.Quantity + oritems.Quantity
			}
		}

		if itemCount > 5 {
			db.Db.Model(&models.Product_Variant{}).Where("id = ?",item.ProductID).Update("stock",gorm.Expr("stock + ?",item.Quantity))
			db.Db.Delete(&models.Order{},order.ID)
			c.HTML(http.StatusBadRequest,"checkOut.html",gin.H{"user":"done","error":"User exceeded product purchase limit"})
			return 
		}

		orderItem := models.OrderItem{
			UserID: uint(userID),
			OrderID: order.ID,
			ProductID: item.ProductID,
			Quantity: item.Quantity,
			Status: "Processing",
			Price: item.Price,		
		}

		if err := db.Db.Create(&orderItem).Error; err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to add order items"})
			return 
		}


		// remove from wishlist
		var wishlist models.WishList

		if err := db.Db.Where("user_id = ? AND product_id = ?",userID,item.ProductID).First(&wishlist).Error; err == nil{
			db.Db.Delete(&wishlist)
		}

	}


	if err := db.Db.Delete(&CartItems).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to clear cart items"})
		return 
	}
	

	c.HTML(http.StatusOK,"orderSuccess.html",gin.H{"OrderID":order.ID,"user":"done"})

}

func OrderConfirmation(c *gin.Context){

	id := c.Param("id")
	c.HTML(http.StatusOK,"orderSuccess.html",gin.H{"OrderID":id,"user":"done"})
}


func AddNewAddressPage(c *gin.Context){
	c.HTML(http.StatusOK,"addAddress.html",gin.H{"user":"done"})
}

func AddNewAddress(c *gin.Context){

	tokenStr,_ := c.Cookie("JWT-User")
	_,userID,_ := helper.DecodeJWT(tokenStr)

	AddressLine1:= c.PostForm("line1")
	AddressLine2:= c.PostForm("line2")
	Country:= c.PostForm("country")
	State:= c.PostForm("state")
	PostalCode:= c.PostForm("postalcode")
	City := c.PostForm("city")

	// address.AddressLine1 = AddressLine1
	// address.AddressLine2 = AddressLine2
	// address.Country = Country
	// address.State = State
	// address.PostalCode = PostalCode
	// address.City = City
	// address.UserID = uint(userID)

	address := models.Address{
		UserID: uint(userID),
		AddressLine1: AddressLine1,
		AddressLine2: AddressLine2,
		Country: Country,
		State: State,
		PostalCode: PostalCode,
		City: City,
	}

	if err := db.Db.Create(&address).Error; err != nil{
		c.HTML(http.StatusInternalServerError,"checkOut.html",gin.H{"error":"Failed to save new address"})
		return 
	}

	c.Redirect(http.StatusSeeOther,"/user/checkout")

}