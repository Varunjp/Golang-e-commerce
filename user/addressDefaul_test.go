package user

import (
	db "first-project/DB"
	"first-project/helper"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMakeAddressDefault(t *testing.T) {

	helper.DecodEJWT = func(token string) (string, float64, error) {
		return "varun", 10.0, nil
	}

	sqlDB, mock, _ := sqlmock.New()
	db.Db, _ = gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	// --- 3. Expected SELECT query ---
	rows := sqlmock.NewRows([]string{
		"address_id", "user_id", "is_default",
	}).AddRow(1, 10.0, false).
		AddRow(2, 10.0, true)

	mock.ExpectQuery(`SELECT .* FROM "addresses" WHERE user_id = \$1`).
		WithArgs(10.0).
		WillReturnRows(rows)

	// --- 4. Expected UPDATE queries ---
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "addresses"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "addresses"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// --- 5. Setup Gin router + session ---
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.POST("/make-default", MakeAddressDefault)

	// --- 6. Build HTTP request ---
	form := url.Values{}
	form.Add("address_id", "1")

	req := httptest.NewRequest("POST", "/make-default",
		strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Add cookie for token
	req.AddCookie(&http.Cookie{
		Name:  "JWT-User",
		Value: "dummy.jwt.token",
	})

	w := httptest.NewRecorder()

	// --- 7. Execute request ---
	r.ServeHTTP(w, req)

	// --- 8. Assertions ---
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", w.Code)
	}

	// Verify all mocks were executed
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}
