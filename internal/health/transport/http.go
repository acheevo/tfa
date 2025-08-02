package transport

import (
	"net/http"

	"github.com/acheevo/tfa/internal/health/service"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	service *service.HealthService
}

func NewHealthHandler(service *service.HealthService) *HealthHandler {
	return &HealthHandler{
		service: service,
	}
}

func (h *HealthHandler) GetHealth(c *gin.Context) {
	health := h.service.GetHealth()

	statusCode := http.StatusOK
	if health.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}
