package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/gin-gonic/gin"
)

func TokenAuthMiddleware(expectedToken string) gin.HandlerFunc {
	exp := []byte(expectedToken)

	return func(c *gin.Context) {
		got := extractToken(c)
		if got == "" || subtle.ConstantTimeCompare([]byte(got), exp) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			logger.Log.Warnf("Unauthorized request from %s", c.ClientIP())
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
		if len(parts) == 2 && strings.EqualFold(parts[0], "Token") {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}
