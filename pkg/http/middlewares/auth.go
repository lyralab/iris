package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/auth"
	"go.uber.org/zap"
)

func BasicAuth(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{username: password})
}

// passwordChangeRoutes are the only routes a token flagged with must_change_password may reach.
var passwordChangeRoutes = map[string]bool{
	"PUT /v0/users/password": true,
	"GET /v0/users/me":       true,
}

func ValidateJWTToken(aths auth.AuthServiceInterface, role string, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		jwtToken := strings.TrimPrefix(authHeader, "Bearer ")
		extractedUsername, extractedRole, mustChangePassword, err := aths.ValidateToken(jwtToken)
		if apperr.Is(err, apperr.ErrTokenExpired) {
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
		if mustChangePassword && !passwordChangeRoutes[c.Request.Method+" "+c.FullPath()] {
			logger.Infow("password change required before using this route",
				"username", extractedUsername, "path", c.FullPath())
			c.AbortWithStatusJSON(403, gin.H{
				"status":  "error",
				"message": "password change required",
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
		logger.Infow("Successfully validate jwtToken", "username", extractedUsername)
		c.Next()
	}
}
