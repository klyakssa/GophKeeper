package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware is a middleware that logs the request
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("request",
			zap.String("Accept-Encoding", c.Request.Header.Get("Accept-Encoding")),
			zap.Any("Header", c.Request.Header),
		)
		start := time.Now()

		c.Next()

		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("time", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		)
	}
}
