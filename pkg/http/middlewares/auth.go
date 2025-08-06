package middlewares

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/auth"
	iris_error "github.com/root-ali/iris/pkg/errors"
	"go.uber.org/zap"
	"strings"
)

func BasicAuth(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{username: password})
}

func ValidateJWTToken(aths auth.AuthServiceInterface, role string, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		jwtToken := strings.TrimPrefix(authHeader, "Bearer ")
		logger.Infow("token is ", "token", jwtToken)
		//jwtToken, err := c.Cookie("jwt")
		extractedUsername, extractedRole, err := aths.ValidateToken(jwtToken)
		if errors.Is(err, iris_error.ErrTokenExpired) {
			logger.Error("Token expired")
			c.AbortWithStatusJSON(401, gin.H{
				"status":  "unauthorized",
				"message": "token expired",
			})
			return
		} else if err != nil {
			logger.Errorw("Invalid token", "error", err)
			c.AbortWithStatusJSON(401, gin.H{
				"status":  "error",
				"message": "invalid token",
			})
			return
		}
		if role == "" {
			c.Set("username", extractedUsername)
			c.Set("role", extractedRole)
			c.Next()
			return
		}
		roles := strings.Split(role, ",")
		var validRole = false
		for _, r := range roles {
			if extractedRole == r {
				validRole = true
			}
		}
		if !validRole {
			logger.Errorw("user not authorize to do this actions",
				"username", extractedUsername, "role", extractedRole)
			c.AbortWithStatusJSON(401, gin.H{
				"status":  "unauthorized",
				"message": "user not valid to do this action",
			})
			return
		}
		if err != nil {
			logger.Errorw("cannot validate token", "error", err)
			c.AbortWithStatusJSON(500, gin.H{
				"status":  "error",
				"message": "cannot validate token",
			})
			return
		}
		c.Set("username", extractedUsername)
		c.Set("role", extractedRole)
		logger.Infow("Successfully validate jwtToken", "Token", jwtToken)
		c.Next()
	}
}
