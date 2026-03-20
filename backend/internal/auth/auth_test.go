package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edge-setting/backend/internal/auth"
	"github.com/edge-setting/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func setupConfig() {
	config.Global.Auth.Tokens = []config.TokenEntry{
		{Token: "ro-token", Permission: auth.PermReadOnly},
		{Token: "rw-token", Permission: auth.PermReadWrite},
	}
}

func newEngine(middlewares ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares...)
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.POST("/write", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func TestRequireAuth_NoToken(t *testing.T) {
	setupConfig()
	r := newEngine(auth.RequireAuth())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401，实际 %d", w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	setupConfig()
	r := newEngine(auth.RequireAuth())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401，实际 %d", w.Code)
	}
}

func TestRequireAuth_ValidReadOnly(t *testing.T) {
	setupConfig()
	r := newEngine(auth.RequireAuth())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ro-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
}

func TestRequireAuth_ValidReadWrite(t *testing.T) {
	setupConfig()
	r := newEngine(auth.RequireAuth())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer rw-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
}

func TestRequireReadWrite_ReadOnlyToken(t *testing.T) {
	setupConfig()
	r := newEngine(auth.RequireAuth(), auth.RequireReadWrite())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/write", nil)
	req.Header.Set("Authorization", "Bearer ro-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("期望 403，实际 %d", w.Code)
	}
}

func TestRequireReadWrite_ReadWriteToken(t *testing.T) {
	setupConfig()
	r := newEngine(auth.RequireAuth(), auth.RequireReadWrite())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/write", nil)
	req.Header.Set("Authorization", "Bearer rw-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际 %d", w.Code)
	}
}

func TestQueryParamToken(t *testing.T) {
	setupConfig()
	r := newEngine(auth.RequireAuth())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?token=ro-token", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Query token 应通过认证，实际 %d", w.Code)
	}
}
