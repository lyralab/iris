package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/internal/apperr"
)

// abortWithError writes err as {"status":"error","message":...}. The status
// comes from its apperr Kind; internal causes never reach the client.
func abortWithError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(apperr.HTTPStatus(err), gin.H{"status": "error", "message": apperr.PublicMessage(err)})
}

// abortWithMessage writes a client error in the same shape as abortWithError.
func abortWithMessage(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"status": "error", "message": msg})
}
