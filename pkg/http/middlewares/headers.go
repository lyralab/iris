package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CheckContentTypeHeader(contentType string, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ContentType() == contentType {
			c.Next()
			return
		} else {
			logger.Infow("Invalid contentType header",
				"actualContentType", c.ContentType(),
				"expectedContentType", contentType,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"error", "InvalidContentType")
			c.AbortWithStatusJSON(400, gin.H{
				"status":  "error",
				"message": "Invalid contentType header",
			})
		}
	}
}
