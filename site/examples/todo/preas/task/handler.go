package task

import (
	"net/http"

	"github.com/prea-framework/prea/core"

	"github.com/gin-gonic/gin"

	"todo/preas/task/dtos"
)

// Handler serves the /task routes.
type Handler struct {
	Service *Service `inject:""`
}

var _ core.Router = (*Handler)(nil)

// #region register
func (h *Handler) Register(r gin.IRouter) {
	g := r.Group("/task")
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.POST("", h.create)
	g.PATCH("/:id", h.update)
	g.DELETE("/:id", h.delete)
}

// #endregion register

func (h *Handler) list(c *gin.Context) {
	// GET /task?list=<id> returns only the tasks of that list.
	items, err := h.Service.List(c.Request.Context(), c.Query("list"))
	if err != nil {
		core.Error(c, err)
		return
	}
	core.Reply(c, http.StatusOK, core.Data(items))
}

func (h *Handler) get(c *gin.Context) {
	item, err := h.Service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		core.Error(c, err)
		return
	}
	core.Reply(c, http.StatusOK, core.Data(item))
}

// #region create
func (h *Handler) create(c *gin.Context) {
	var req dtos.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Error(c, core.NewError(http.StatusBadRequest, err.Error()))
		return
	}

	item, err := h.Service.Create(c.Request.Context(), req)
	if err != nil {
		core.Error(c, err)
		return
	}
	core.Reply(c, http.StatusCreated, core.Data(item))
}

// #endregion create

func (h *Handler) update(c *gin.Context) {
	var req dtos.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Error(c, core.NewError(http.StatusBadRequest, err.Error()))
		return
	}

	item, err := h.Service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		core.Error(c, err)
		return
	}
	core.Reply(c, http.StatusOK, core.Data(item))
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.Service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		core.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
