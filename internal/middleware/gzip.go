package middleware

import (
	"gophkeeper/pkg/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipMiddleware is a middleware that compresses the response and decompresses the request
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			gz := gzip.NewCompressWriter(c.Writer)
			c.Writer = gz
			defer gz.Close()
		}

		contentEncoding := c.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := gzip.NewCompressReader(c.Request.Body)
			if err != nil {
				c.Writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			c.Request.Body = cr
			defer cr.Close()
		}

		c.Next()
	}
}
