package user

import (
	db "first-project/DB"
	"first-project/middleware"
	"first-project/utils"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestLogin_Success(t *testing.T) {

	sqlDB, mock, _ := sqlmock.New()
	db.Db, _ = gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "status",
	}).AddRow(1, "varun", "test@example.com", "hashedpwd", "Active")

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
    WithArgs("test@example.com", sqlmock.AnyArg()).
    WillReturnRows(rows)


	utils.ChecKPasswordHash = func(p, h string) bool {
		return true
	}

	middleware.CReateToken = func(role, email string, id uint) (string, error) {
		return "test.jwt.token", nil
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpl := template.Must(template.New("userLogin.html").Parse(`
		<html>{{ .error }}</html>
	`))
	r.SetHTMLTemplate(tmpl)


	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.POST("/login", Login)

	body := strings.NewReader("email=test@example.com&password=12345")
	req, _ := http.NewRequest("POST", "/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	t.Log("Response Body:", w.Body.String())

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	// Check if cookie was set
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "JWT-User" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected JWT-User cookie to be set")
	}
}
