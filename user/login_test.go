package user

import (
	"html/template"
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

	dbpkg "first-project/DB"
	"first-project/middleware"
	"first-project/utils"
)

var (
    checkPassword = utils.CheckPasswordHash
    createToken   = middleware.CreateToken
)

// Setup mock GORM + sqlmock
func setupMockDB(t *testing.T) (sqlmock.Sqlmock, *gorm.DB) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	dialector := postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	})
	gdb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}
	return mock, gdb
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// 1️⃣ Setup DB mock
	mock, gdb := setupMockDB(t)
	dbpkg.Db = gdb

	rows := sqlmock.NewRows([]string{"id", "email", "password", "username", "status"}).
		AddRow(1, "test@example.com", "hashedpass", "TestUser", "Active")

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE.*email.*deleted_at.*ORDER BY "users"."id" LIMIT`).
		WithArgs("test@example.com", sqlmock.AnyArg()).
		WillReturnRows(rows)

	// 2️⃣ Mock functions
	checkPassword = func(password, hash string) bool {
		return true // pretend password always matches
	}
	createToken = func(role, email string, id uint) (string, error) {
		return "mocktoken123", nil
	}

	// 3️⃣ Setup Gin + template
	router := gin.New()
	router.SetHTMLTemplate(template.Must(template.New("userLogin.html").Parse(`OK`)))

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	router.POST("/login", Login)

	// 4️⃣ Make a fake form request
	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", "password123")
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 5️⃣ Assert
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d. Body: %s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("DB expectations not met: %v", err)
	}
}
