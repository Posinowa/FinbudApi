package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{}, 1)

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// İstek süre içinde tamamlandı
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "İstek zaman aşımına uğradı. Lütfen tekrar deneyin.",
			})
		}
	}
}
