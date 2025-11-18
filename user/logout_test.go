package user

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func TestUserLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/set-session", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("name", "varun")
		session.Save()
		c.String(200, "ok")
	})

	r.GET("/logout", UserLogout)

	// 1️⃣ First request: Set session
	req1, _ := http.NewRequest("GET", "/set-session", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// Extract session cookie from first response
	var sessionCookie *http.Cookie
	for _, c := range w1.Result().Cookies() {
		if c.Name == "mysession" {
			sessionCookie = c
		}
	}

	if sessionCookie == nil {
		t.Fatal("session cookie not set")
	}

	// 2️⃣ Second request: Call logout with session cookie
	req2, _ := http.NewRequest("GET", "/logout", nil)
	req2.AddCookie(sessionCookie)
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)

	// Expect redirect (302)
	if w2.Code != 302 {
		t.Errorf("expected 302, got %d", w2.Code)
	}

	// JWT cookie should be cleared
	found := false
	for _, c := range w2.Result().Cookies() {
		if c.Name == "JWT-User" && c.MaxAge == -1 {
			found = true
		}
	}
	if !found {
		t.Errorf("JWT-User cookie was not cleared")
	}
}