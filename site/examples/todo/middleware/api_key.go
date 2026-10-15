package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/prea-framework/prea/core"

	"github.com/gin-gonic/gin"
)

// #region component
// APIKey protects the writes: POST, PATCH and DELETE need the X-API-Key
// header when todo.api_key is set. Reads stay open.
type APIKey struct {
	Key string `value:"todo.api_key"`
}

var _ core.MiddlewareComponent = (*APIKey)(nil)

// #endregion component

// #region init
// Init runs once, at startup: a short key stops the app before it serves.
func (m *APIKey) Init(ctx context.Context) error {
	if m.Key != "" && len(m.Key) < 16 {
		return errors.New("todo.api_key must have at least 16 characters")
	}
	return nil
}

// #endregion init

// #region handle
func (m *APIKey) Handle(c *gin.Context) {
	if m.Key == "" || c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
		c.Next()
		return
	}
	if subtle.ConstantTimeCompare([]byte(c.GetHeader("X-API-Key")), []byte(m.Key)) != 1 {
		core.Error(c, core.NewError(http.StatusUnauthorized, "missing or wrong X-API-Key"))
		c.Abort()
		return
	}
	c.Next()
}

// #endregion handle
