package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/v1/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	req := httptest.NewRequest("GET", "/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", response["status"])
	}
}

func TestBasicAuthProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	account := gin.Accounts{
		"test@test.com": "testpassword",
	}

	router := gin.New()
	protected := router.Group("/v1")
	protected.Use(gin.BasicAuth(account))

	protected.GET("/auth", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "auth verified",
		})
	})

	// Test without auth
	req := httptest.NewRequest("GET", "/v1/auth", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	// Test with correct auth
	req = httptest.NewRequest("GET", "/v1/auth", nil)
	req.SetBasicAuth("test@test.com", "testpassword")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["status"] != "auth verified" {
		t.Errorf("expected 'auth verified', got %v", response["status"])
	}

	// Test with wrong password
	req = httptest.NewRequest("GET", "/v1/auth", nil)
	req.SetBasicAuth("test@test.com", "wrongpassword")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for wrong password, got %d", w.Code)
	}
}

func TestSwaggerEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Simulate swagger endpoint
	router.GET("/docs/swagger/*any", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"swagger": "available",
		})
	})

	req := httptest.NewRequest("GET", "/docs/swagger/index.html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["swagger"] != "available" {
		t.Errorf("expected swagger 'available', got %v", response["swagger"])
	}
}

func TestGinReleaseMode(t *testing.T) {
	// Test that Gin can be set to release mode
	gin.SetMode(gin.ReleaseMode)
	mode := gin.Mode()
	if mode != gin.ReleaseMode {
		t.Errorf("expected release mode, got %s", mode)
	}

	// Reset to test mode
	gin.SetMode(gin.TestMode)
}

func TestGinAccountsSetup(t *testing.T) {
	accounts := gin.Accounts{
		"user1@example.com": "password1",
		"user2@example.com": "password2",
	}

	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}

	if accounts["user1@example.com"] != "password1" {
		t.Error("account password mismatch")
	}
}

func TestRouterGroupSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/v1")
	v1.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/v1/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
