package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := getAllowedOrigins()

	config := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			for _, allowed := range allowedOrigins {
				if allowed == "*" {
					return true
				}
				if strings.EqualFold(allowed, origin) {
					return true
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Accept", "Origin", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}

func getAllowedOrigins() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return []string{}
	}

	origins := strings.Split(raw, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}
	return origins
}
