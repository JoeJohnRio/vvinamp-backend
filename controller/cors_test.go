package controller

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStaticMediaCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(StaticMediaCORSMiddleware())
	r.GET("/album/test.jpg", func(c *gin.Context) {
		c.String(http.StatusOK, "image bytes")
	})

	// Case 1: Request with Origin
	req, _ := http.NewRequest(http.MethodGet, "/album/test.jpg", nil)
	req.Header.Set("Origin", "http://localhost:8081")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Case 2: Request WITHOUT Origin (crucial for CDN caching!)
	req2, _ := http.NewRequest(http.MethodGet, "/album/test.jpg", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	if w2.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: * on empty Origin request, got '%s'", w2.Header().Get("Access-Control-Allow-Origin"))
	}

	// Case 3: OPTIONS Preflight
	req3, _ := http.NewRequest(http.MethodOptions, "/album/test.jpg", nil)
	req3.Header.Set("Origin", "http://localhost:8081")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w3.Code)
	}
	if w3.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: * on OPTIONS, got '%s'", w3.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestAPICORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("CORS_ALLOW_ORIGINS", "http://localhost:8081,http://localhost:3000")
	os.Setenv("CORS_ALLOW_CREDENTIALS", "true")

	r := gin.New()
	r.Use(CORSMiddleware())
	r.POST("/v1/graphql", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	// Case 1: Allowed Origin with credentials
	req, _ := http.NewRequest(http.MethodPost, "/v1/graphql", nil)
	req.Header.Set("Origin", "http://localhost:8081")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8081" {
		t.Errorf("expected Access-Control-Allow-Origin: http://localhost:8081, got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials: true, got '%s'", w.Header().Get("Access-Control-Allow-Credentials"))
	}

	// Case 3: Origin http://localhost:8080 (Wasm frontend)
	req3, _ := http.NewRequest(http.MethodOptions, "/v1/graphql", nil)
	req3.Header.Set("Origin", "http://localhost:8080")
	req3.Header.Set("Access-Control-Request-Headers", "authorization, content-type, apollo-require-preflight")
	req3.Header.Set("Access-Control-Request-Method", "POST")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w3.Code)
	}
	if w3.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8080" {
		t.Errorf("expected Access-Control-Allow-Origin: http://localhost:8080, got '%s'", w3.Header().Get("Access-Control-Allow-Origin"))
	}
	if w3.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials: true, got '%s'", w3.Header().Get("Access-Control-Allow-Credentials"))
	}
	if w3.Header().Get("Access-Control-Allow-Headers") != "authorization, content-type, apollo-require-preflight" {
		t.Errorf("expected reflected Access-Control-Allow-Headers, got '%s'", w3.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestStreamYoutubeRoute_CORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := &GraphQLController{}
	r.OPTIONS("/v1/stream/youtube/:id", StaticMediaCORSMiddleware())
	r.GET("/v1/stream/youtube/:id", StaticMediaCORSMiddleware(), ctrl.StreamYoutubeHandler)

	// Test OPTIONS preflight for the streaming endpoint
	req, _ := http.NewRequest(http.MethodOptions, "/v1/stream/youtube/dQw4w9WgXcQ", nil)
	req.Header.Set("Origin", "https://vvinamp.cirio.my.id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Expose-Headers") == "" {
		t.Errorf("expected non-empty Access-Control-Expose-Headers for media seeking")
	}
}
