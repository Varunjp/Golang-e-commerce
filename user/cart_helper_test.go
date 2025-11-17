package user

import (
	db "first-project/DB"
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
)

func setupMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		_ = sqlDB.Close()
		t.Fatalf("gorm open error: %v", err)
	}

	// assign to global db
	db.Db = gdb

	return mock, func() { _ = sqlDB.Close() }
}

func setupRouterAndJWT() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	// override DecodeJWT in tests (restore done per-test)
	return router
}

// mock the preload chain for productVariant -> product -> subcategory -> category
func mockPreloadChain(mock sqlmock.Sqlmock, variantID int) {
	// product_variants
	pvCols := []string{
		"id", "product_id", "variant_name", "size", "stock",
		"price", "tax", "discounted_price", "created_at",
		"updated_at", "is_active", "deleted_at",
	}
	pvRows := sqlmock.NewRows(pvCols).AddRow(
		variantID, 1, "Variant1", "M", 20,
		500.0, 50.0, 450.0, time.Now(), time.Now(), true, nil,
	)
	mock.ExpectQuery(`SELECT .* FROM "product_variants"`).
		WithArgs(variantID, sqlmock.AnyArg()).
		WillReturnRows(pvRows)

	// products
	pCols := []string{"product_id", "product_name", "description", "sub_category_id", "created_at", "deleted_at"}
	pRows := sqlmock.NewRows(pCols).AddRow(1, "Shirt", "Cotton shirt", 5, time.Now(), nil)
	mock.ExpectQuery(`SELECT .* FROM "products"`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(pRows)

	// sub_categories
	scCols := []string{"sub_category_id", "sub_category_name", "category_id", "is_blocked", "category_discount", "deleted_at"}
	scRows := sqlmock.NewRows(scCols).AddRow(5, "Mens Wear", 2, false, 0, nil)
	mock.ExpectQuery(`SELECT .* FROM "sub_categories"`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(scRows)

	// categories
	catCols := []string{"category_id", "category_name", "create_at", "is_blocked", "deleted_at"}
	catRows := sqlmock.NewRows(catCols).AddRow(2, "Clothing", time.Now(), false, nil)
	mock.ExpectQuery(`SELECT .* FROM "categories"`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(catRows)
}

// mock cart select (First) — userID may be AnyArg (float/int). productID provided.
func mockCartSelect(mock sqlmock.Sqlmock, userArg interface{}, productID int, exists bool) {
	cols := []string{"id", "user_id", "product_id", "quantity", "price"}
	if exists {
		rows := sqlmock.NewRows(cols).AddRow(5, 1, productID, 2, 500)
		// userArg can be sqlmock.AnyArg() or exact
		mock.ExpectQuery(`SELECT .* FROM "cart_items"`).
			WithArgs(userArg, productID, sqlmock.AnyArg()).
			WillReturnRows(rows)
	} else {
		mock.ExpectQuery(`SELECT .* FROM "cart_items"`).
			WithArgs(userArg, productID, sqlmock.AnyArg()).
			WillReturnError(gorm.ErrRecordNotFound)
	}
}

// mock cart update (Save)
func mockCartUpdate(mock sqlmock.Sqlmock, id int) {
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "cart_items"`).
		WithArgs(
			sqlmock.AnyArg(), // user_id
			sqlmock.AnyArg(), // product_id
			sqlmock.AnyArg(), // quantity
			sqlmock.AnyArg(), // price
			sqlmock.AnyArg(), // add_at
			id,               // WHERE id = ?
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

// mock cart update failure
func mockCartUpdateFail(mock sqlmock.Sqlmock, id int) {
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "cart_items"`).
		WillReturnError(assertError("update failed"))
	mock.ExpectRollback()
}

// mock cart insert
func mockCartInsert(mock sqlmock.Sqlmock) {
    mock.ExpectBegin()
    mock.ExpectQuery(`INSERT INTO "cart_items"`).
        WithArgs(
            sqlmock.AnyArg(), // user_id
            sqlmock.AnyArg(), // product_id
            sqlmock.AnyArg(), // quantity
            sqlmock.AnyArg(), // price
            sqlmock.AnyArg(), // add_at
        ).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
    mock.ExpectCommit()
	
}

// mock cart insert failure
func mockCartInsertFail(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "cart_items"`).
		WillReturnError(assertError("insert failed"))
	mock.ExpectRollback()
}

// mock wishlist select (exists or not)
func mockWishlistSelect(mock sqlmock.Sqlmock, userArg interface{}, productID int, exists bool) {
	query := `SELECT \* FROM "wish_lists" WHERE user_id = \$1 AND product_id = \$2 ORDER BY "wish_lists"\."id" LIMIT \$3`

    cols := []string{"id", "user_id", "product_id", "created_at"}

    if exists {
        rows := sqlmock.NewRows(cols).
            AddRow(7, 1, productID, time.Now())

        mock.ExpectQuery(query).
            WithArgs(sqlmock.AnyArg(), productID, sqlmock.AnyArg()).
            WillReturnRows(rows)
    } else {
        mock.ExpectQuery(query).
            WithArgs(sqlmock.AnyArg(), productID, sqlmock.AnyArg()).
            WillReturnError(gorm.ErrRecordNotFound)
    }
}

// mock wishlist delete
func mockWishlistDelete(mock sqlmock.Sqlmock, id int) {
	mock.ExpectExec(`DELETE FROM "wish_lists"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// helper to create a sentinel error (sqlmock expects error type)
type simpleError string

func (e simpleError) Error() string { return string(e) }

func assertError(msg string) error { return simpleError(msg) }

// small helper to run request and return recorder
func runRequest(router *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/add", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "JWT-User", Value: "dummy-token"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}