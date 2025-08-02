package transport

import (
	"net/http"

	"github.com/acheevo/tfa/internal/info/service"
	"github.com/gin-gonic/gin"
)

type InfoHandler struct {
	service *service.InfoService
}

func NewInfoHandler(service *service.InfoService) *InfoHandler {
	return &InfoHandler{
		service: service,
	}
}

func (h *InfoHandler) GetInfo(c *gin.Context) {
	info := h.service.GetInfo()
	c.JSON(http.StatusOK, info)
}
