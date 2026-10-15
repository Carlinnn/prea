package health

import (
	"net/http"

	"github.com/prea-framework/prea/core"

	"github.com/gin-gonic/gin"
)

// Handler answers GET /health, for load balancers and uptime checks.
type Handler struct{}

var _ core.Router = (*Handler)(nil)

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/health", h.health)
}

func (h *Handler) health(c *gin.Context) {
	core.Reply(c, http.StatusOK, core.Data(gin.H{"status": "ok"}))
}
