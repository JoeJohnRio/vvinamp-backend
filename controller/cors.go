package controller

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// StaticMediaCORSMiddleware applies permissive CORS headers for static media files (images, audio, HLS).
// It unconditionally sets Access-Control-Allow-Origin: * so that caching reverse proxies / CDNs (like Cloudflare)
// will always cache and serve the CORS header, preventing browser fetch/crossorigin errors.
func StaticMediaCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Range, Origin, Accept")
		c.Header("Access-Control-Expose-Headers", "Content-Range, Accept-Ranges, Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return true
	}
	for _, o := range allowedOrigins {
		if o == "*" || o == origin {
			return true
		}
		if (o == "http://localhost" || o == "http://127.0.0.1" || o == "localhost") &&
			(strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") || origin == "http://localhost" || origin == "http://127.0.0.1") {
			return true
		}
	}
	// Automatically allow any localhost or 127.0.0.1 origin on any port for development
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") || origin == "http://localhost" || origin == "http://127.0.0.1" {
		return true
	}
	// Allow any Vercel or cirio.my.id origin for web deployments
	if strings.HasSuffix(origin, ".vercel.app") || strings.HasSuffix(origin, ".cirio.my.id") || origin == "https://cirio.my.id" {
		return true
	}
	return false
}

// CORSMiddleware returns a gin middleware that applies CORS headers based on env vars
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed := parseAllowedOrigins(os.Getenv("CORS_ALLOW_ORIGINS"))
		allowCreds := parseBoolEnv(os.Getenv("CORS_ALLOW_CREDENTIALS"), true)

		origin := c.Request.Header.Get("Origin")
		c.Header("Vary", "Origin")

		if origin != "" && isOriginAllowed(origin, allowed) {
			if allowCreds {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
			} else {
				if contains(allowed, "*") {
					c.Header("Access-Control-Allow-Origin", "*")
				} else {
					c.Header("Access-Control-Allow-Origin", origin)
				}
			}
		} else if origin == "" && contains(allowed, "*") && !allowCreds {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS")

		reqHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
		if reqHeaders != "" {
			c.Header("Access-Control-Allow-Headers", reqHeaders)
		} else {
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Apollo-Require-Preflight, Range, Origin, Accept")
		}

		c.Header("Access-Control-Expose-Headers", "Content-Range, Accept-Ranges, Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
