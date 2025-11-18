package user

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	db "first-project/DB"
	"first-project/helper"
)

func TestAddToCart_UpdateExistingCart(t *testing.T) {

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	db.Db, err = gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm with sqlmock: %v", err)
	}

	db.Db = db.Db.Debug()
	
	pvRows := sqlmock.NewRows([]string{
		"id", "product_id", "variant_name", "size", "stock",
		"price", "tax", "discounted_price", "created_at",
		"updated_at", "is_active", "deleted_at",
	}).AddRow(
		10, 1, "Variant1", "M", 20,
		500.0, 50.0, 450.0,
		time.Now(), time.Now(), true, nil,
	)

	mock.ExpectQuery(`SELECT .* FROM "product_variants"`).
		WithArgs(10, sqlmock.AnyArg()).
		WillReturnRows(pvRows)

	pRows := sqlmock.NewRows([]string{
		"product_id", "product_name", "description",
		"sub_category_id", "created_at", "deleted_at",
	}).AddRow(
		1, "Shirt", "Cotton shirt", 5, time.Now(), nil,
	)

	mock.ExpectQuery(`SELECT .* FROM "products"`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(pRows)

	scRows := sqlmock.NewRows([]string{
		"sub_category_id", "sub_category_name",
		"category_id", "is_blocked", "category_discount",
		"deleted_at",
	}).AddRow(
		5, "Mens Wear", 2, false, 0, nil,
	)

	mock.ExpectQuery(`SELECT .* FROM "sub_categories"`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(scRows)

	catRows := sqlmock.NewRows([]string{
		"category_id", "category_name", "create_at",
		"is_blocked", "deleted_at",
	}).AddRow(
		2, "Clothing", time.Now(), false, nil,
	)

	mock.ExpectQuery(`SELECT .* FROM "categories"`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(catRows)

	cartRows := sqlmock.NewRows([]string{
		"id", "user_id", "product_id", "quantity", "price",
	}).AddRow(
		5, 1, 10, 2, 500,
	)

	mock.ExpectQuery(`SELECT .* FROM "cart_items"`).
		WithArgs(sqlmock.AnyArg(), 10, sqlmock.AnyArg()).
		WillReturnRows(cartRows)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "cart_items"`).
    WithArgs(
        sqlmock.AnyArg(), // user_id
        sqlmock.AnyArg(), // product_id
        sqlmock.AnyArg(), // quantity
        sqlmock.AnyArg(), // price
        sqlmock.AnyArg(), // add_at
        5,                // id
    ).
    WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)

	body := "product_id=10&quantity=3"
	req := httptest.NewRequest("POST", "/add", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  "JWT-User",
		Value: "dummy-token",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", w.Code)
	}

	found := false
	for _, c := range w.Result().Cookies() {
		if c.Name == "mysession" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected session cookie to be set")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestProductNotFound(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	// product_variants returns no rows
	mock.ExpectQuery(`SELECT .* FROM "product_variants"`).
		WithArgs(10, sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	router := setupRouterAndJWT()
	// override DecodeJWT
	
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}
	// tests
	router.POST("/add", AddToCart)
	w := runRequest(router, "product_id=10&quantity=1")

	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}

func TestProductBlocked(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	// create pv row
	pvCols := []string{"id", "product_id", "variant_name", "size", "stock", "price", "tax", "discounted_price", "created_at", "updated_at", "is_active", "deleted_at"}
	pvRows := sqlmock.NewRows(pvCols).AddRow(10, 1, "v", "M", 20, 500.0, 50.0, 450.0, time.Now(), time.Now(), true, nil)
	mock.ExpectQuery(`SELECT .* FROM "product_variants"`).
		WithArgs(10, sqlmock.AnyArg()).
		WillReturnRows(pvRows)

	// product row but subcategory.is_blocked true
	pCols := []string{"product_id", "product_name", "description", "sub_category_id", "created_at", "deleted_at"}
	pRows := sqlmock.NewRows(pCols).AddRow(1, "Shirt", "desc", 5, time.Now(), nil)
	mock.ExpectQuery(`SELECT .* FROM "products"`).
		WillReturnRows(pRows)

	scCols := []string{"sub_category_id", "sub_category_name", "category_id", "is_blocked", "category_discount", "deleted_at"}
	scRows := sqlmock.NewRows(scCols).AddRow(5, "Mens", 2, true, 0, nil) // blocked = true
	mock.ExpectQuery(`SELECT .* FROM "sub_categories"`).
		WillReturnRows(scRows)

	catCols := []string{"category_id", "category_name", "create_at", "is_blocked", "deleted_at"}
	catRows := sqlmock.NewRows(catCols).AddRow(2, "Cloth", time.Now(), false, nil)
	mock.ExpectQuery(`SELECT .* FROM "categories"`).
		WillReturnRows(catRows)

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)
	w := runRequest(router, "product_id=10&quantity=1")

	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}

// 3 Stock == 0
func TestStockZero(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	pvCols := []string{"id", "product_id", "variant_name", "size", "stock", "price", "tax", "discounted_price", "created_at", "updated_at", "is_active", "deleted_at"}
	pvRows := sqlmock.NewRows(pvCols).AddRow(10, 1, "v", "M", 0, 500.0, 50.0, 450.0, time.Now(), time.Now(), true, nil)
	mock.ExpectQuery(`SELECT .* FROM "product_variants"`).
		WithArgs(10, sqlmock.AnyArg()).
		WillReturnRows(pvRows)

	// product + subcategory + category still must be returned (preload)
	pCols := []string{"product_id", "product_name", "description", "sub_category_id", "created_at", "deleted_at"}
	mock.ExpectQuery(`SELECT .* FROM "products"`).WillReturnRows(sqlmock.NewRows(pCols).AddRow(1, "Shirt", "d", 5, time.Now(), nil))
	mock.ExpectQuery(`SELECT .* FROM "sub_categories"`).WillReturnRows(sqlmock.NewRows([]string{"sub_category_id", "sub_category_name", "category_id", "is_blocked", "category_discount", "deleted_at"}).AddRow(5, "S", 2, false, 0, nil))
	mock.ExpectQuery(`SELECT .* FROM "categories"`).WillReturnRows(sqlmock.NewRows([]string{"category_id", "category_name", "create_at", "is_blocked", "deleted_at"}).AddRow(2, "C", time.Now(), false, nil))

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)
	w := runRequest(router, "product_id=10&quantity=1")
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}

// // 4 Requested quantity > stock
func TestQuantityGreaterThanStock(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	// product variant stock = 2
	pvCols := []string{"id", "product_id", "variant_name", "size", "stock", "price", "tax", "discounted_price", "created_at", "updated_at", "is_active", "deleted_at"}
	pvRows := sqlmock.NewRows(pvCols).AddRow(10, 1, "v", "M", 2, 500.0, 50.0, 450.0, time.Now(), time.Now(), true, nil)
	mock.ExpectQuery(`SELECT .* FROM "product_variants"`).WithArgs(10, sqlmock.AnyArg()).WillReturnRows(pvRows)
	// preload chain
	mock.ExpectQuery(`SELECT .* FROM "products"`).WillReturnRows(sqlmock.NewRows([]string{"product_id", "product_name", "description", "sub_category_id", "created_at", "deleted_at"}).AddRow(1, "S", "d", 5, time.Now(), nil))
	mock.ExpectQuery(`SELECT .* FROM "sub_categories"`).WillReturnRows(sqlmock.NewRows([]string{"sub_category_id", "sub_category_name", "category_id", "is_blocked", "category_discount", "deleted_at"}).AddRow(5, "S", 2, false, 0, nil))
	mock.ExpectQuery(`SELECT .* FROM "categories"`).WillReturnRows(sqlmock.NewRows([]string{"category_id", "category_name", "create_at", "is_blocked", "deleted_at"}).AddRow(2, "C", time.Now(), false, nil))

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)
	// request quantity = 3 (> stock 2)
	w := runRequest(router, "product_id=10&quantity=3")
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}

// // 5 Requested quantity > Limit
func TestQuantityGreaterThanLimit(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	// pv stock large enough but request > Limit (Limit is 5)
	pvCols := []string{"id", "product_id", "variant_name", "size", "stock", "price", "tax", "discounted_price", "created_at", "updated_at", "is_active", "deleted_at"}
	pvRows := sqlmock.NewRows(pvCols).AddRow(10, 1, "v", "M", 20, 500.0, 50.0, 450.0, time.Now(), time.Now(), true, nil)
	mock.ExpectQuery(`SELECT .* FROM "product_variants"`).WithArgs(10, sqlmock.AnyArg()).WillReturnRows(pvRows)
	// preload
	mock.ExpectQuery(`SELECT .* FROM "products"`).WillReturnRows(sqlmock.NewRows([]string{"product_id", "product_name", "description", "sub_category_id", "created_at", "deleted_at"}).AddRow(1, "S", "d", 5, time.Now(), nil))
	mock.ExpectQuery(`SELECT .* FROM "sub_categories"`).WillReturnRows(sqlmock.NewRows([]string{"sub_category_id", "sub_category_name", "category_id", "is_blocked", "category_discount", "deleted_at"}).AddRow(5, "S", 2, false, 0, nil))
	mock.ExpectQuery(`SELECT .* FROM "categories"`).WillReturnRows(sqlmock.NewRows([]string{"category_id", "category_name", "create_at", "is_blocked", "deleted_at"}).AddRow(2, "C", time.Now(), false, nil))

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)
	// quantity = 6 (> Limit 5)
	w := runRequest(router, "product_id=10&quantity=6")
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}

// // 6 Cart exists + adding quantity exceeds stock
func TestCartExistsAddExceedsStock(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	mockPreloadChain(mock, 10)
	// cart exists with quantity 19, product stock 20, adding 2 => new 21 > 20
	// but we'll craft cart with quantity 19 to simulate
	cartRows := sqlmock.NewRows([]string{"id", "user_id", "product_id", "quantity", "price"}).AddRow(5, 1, 10, 19, 500)
	mock.ExpectQuery(`SELECT .* FROM "cart_items"`).
		WithArgs(sqlmock.AnyArg(), 10, sqlmock.AnyArg()).
		WillReturnRows(cartRows)

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)
	// request quantity 2 -> 19+2=21 > stock(20)
	w := runRequest(router, "product_id=10&quantity=2")
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}

// // 7 Cart exists + adding quantity exceeds limit
func TestCartExistsAddExceedsLimit(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	mockPreloadChain(mock, 10)
	// cart exists with quantity 4, Limit=5; adding 2 => 6 > 5
	cartRows := sqlmock.NewRows([]string{"id", "user_id", "product_id", "quantity", "price"}).AddRow(5, 1, 10, 4, 500)
	mock.ExpectQuery(`SELECT .* FROM "cart_items"`).
		WithArgs(sqlmock.AnyArg(), 10, sqlmock.AnyArg()).
		WillReturnRows(cartRows)

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)
	// request quantity 2 -> 4+2=6 > Limit(5)
	w := runRequest(router, "product_id=10&quantity=2")
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}

// // 8 New cart creation (no existing cart)
func TestNewCartCreation(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	mockPreloadChain(mock, 10)
	// cart not found => gorm ErrRecordNotFound
	mock.ExpectQuery(`SELECT .* FROM "cart_items"`).
		WithArgs(sqlmock.AnyArg(), 10, sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	// expect insert
	mockCartInsert(mock)

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)
	w := runRequest(router, "product_id=10&quantity=3")
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}
func TestUpdateFails(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	mockPreloadChain(mock, 10)
	// cart exists
	cartRows := sqlmock.NewRows([]string{"id", "user_id", "product_id", "quantity", "price"}).AddRow(5, 1, 10, 2, 500)
	mock.ExpectQuery(`SELECT .* FROM "cart_items"`).
		WithArgs(sqlmock.AnyArg(), 10, sqlmock.AnyArg()).
		WillReturnRows(cartRows)

	// update fails
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "cart_items"`).
		WillReturnError(assertError("some update error"))
	mock.ExpectRollback()

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add",AddToCart)
	w := runRequest(router, "product_id=10&quantity=3")
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}

// // 11 Insert fails
func TestInsertFails(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	mockPreloadChain(mock, 10)
	// cart not found
	mock.ExpectQuery(`SELECT .* FROM "cart_items"`).
		WithArgs(sqlmock.AnyArg(), 10, sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	// insert fails
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "cart_items"`).
        WithArgs(
            sqlmock.AnyArg(), // user_id
            sqlmock.AnyArg(), // product_id
            sqlmock.AnyArg(), // quantity
            sqlmock.AnyArg(), // price
            sqlmock.AnyArg(), // add_at
        ).
        WillReturnError(errors.New("insert fail"))
	mock.ExpectRollback()

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)
	w := runRequest(router, "product_id=10&quantity=3")
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}

// // 12 Wishlist query fails (error not ErrRecordNotFound)
func TestWishlistQueryFails(t *testing.T) {
	mock, closeDB := setupMockDB(t)
	defer closeDB()

	mockPreloadChain(mock, 10)
	// cart not found
	mock.ExpectQuery(`SELECT .* FROM "cart_items"`).
		WithArgs(sqlmock.AnyArg(), 10, sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	// insert succeeds
	mockCartInsert(mock)

	// wishlist select fails with other error
	mock.ExpectQuery(`SELECT .* FROM "wish_lists"`).
		WithArgs(sqlmock.AnyArg(), 10, sqlmock.AnyArg()).
		WillReturnError(assertError("db down"))

	router := setupRouterAndJWT()
	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 1.0, nil
	}

	router.POST("/add", AddToCart)
	w := runRequest(router, "product_id=10&quantity=2")
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect; got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet exps: %v", err)
	}
}