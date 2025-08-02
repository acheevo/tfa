package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("panic recovered",
			"error", recovered,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
}
