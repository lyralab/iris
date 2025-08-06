package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/captcha"
	"go.uber.org/zap"
)

type CaptchaChallangeResponse struct {
	ID  string `json:"id"`
	B64 string `json:"b64"`
}

// GenerateCaptchaHandler creates a new captcha challenge
func GenerateCaptchaHandler(cs captcha.CaptchaServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, b64, err := cs.GenerateCaptcha()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate captcha",
			})
			return
		}
		captchaChallenge := CaptchaChallangeResponse{
			ID:  id,
			B64: b64,
		}

		logger.Infow("Captcha generated successfully",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote_addr", c.Request.RemoteAddr,
			"captcha_id", captchaChallenge.ID,
		)

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   captchaChallenge,
		})
	}
}

// // VerifyCaptchaHandler verifies a captcha answer
// func VerifyCaptchaHandler(cs captcha.CaptchaServiceInterface, logger *zap.SugaredLogger) gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		if err := c.ShouldBindJSON(&verification); err != nil {
// 			logger.Errorw("Invalid request body for captcha verification",
// 				"method", c.Request.Method,
// 				"path", c.Request.URL.Path,
// 				"remote_addr", c.Request.RemoteAddr,
// 				"error", err,
// 			)
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"status":  "error",
// 				"message": "Invalid request body",
// 			})
// 			return
// 		}

// 		if verification.ID == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"status":  "error",
// 				"message": "Captcha ID is required",
// 			})
// 			return
// 		}

// 		isValid := cs.VerifyCaptcha(&verification)

// 		logger.Infow("Captcha verification attempt",
// 			"method", c.Request.Method,
// 			"path", c.Request.URL.Path,
// 			"remote_addr", c.Request.RemoteAddr,
// 			"captcha_id", verification.ID,
// 			"valid", isValid,
// 		)

// 		if isValid {
// 			c.JSON(http.StatusOK, gin.H{
// 				"status": "success",
// 				"valid":  true,
// 			})
// 		} else {
// 			c.JSON(http.StatusOK, gin.H{
// 				"status": "success",
// 				"valid":  false,
// 			})
// 		}
// 	}
// }
