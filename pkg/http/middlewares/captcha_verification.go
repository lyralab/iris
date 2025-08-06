package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/captcha"
	"go.uber.org/zap"
)

type CaptchaVerification struct {
	Answer string `json:"captcha_answer"`
}

func VerifyCaptchaMiddleware(cs captcha.CaptchaServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("captcha_id")
		if id == "" {
			logger.Errorw("Captcha ID is missing",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"remote_addr", c.Request.RemoteAddr,
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Captcha ID is required",
			})
			c.Abort()
			return
		}
		rawBody, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
			return
		}
		// restore the io.Reader stream for the next handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

		// 3) Unmarshal only the captcha_answer
		var p CaptchaVerification
		if err := json.Unmarshal(rawBody, &p); err != nil {
			logger.Errorw("Invalid JSON format for captcha verification",
				"method", c.Request.Method,
				"remote_addr", c.Request.RemoteAddr,
				"error", err,
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		if p.Answer == "" {
			logger.Errorw("Missing captcha answer",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"remote_addr", c.Request.RemoteAddr,
				"captcha_id", id,
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Captcha answer is required",
			})
			return
		}

		// Verify captcha with the service
		isValid := cs.VerifyCaptcha(id, p.Answer)
		if !isValid {
			logger.Warnw("Captcha verification failed",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"remote_addr", c.Request.RemoteAddr,
				"captcha_id", id,
				"captcha_answer", p.Answer,
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Captcha verification failed",
			})
			c.Abort()
			return
		}

		logger.Infow("Captcha verified successfully",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote_addr", c.Request.RemoteAddr,
			"captcha_id", id,
			"captcha_answer", p.Answer,
		)

		c.Next()
	}
}
