package core

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// headKey marks a GET request that answers a HEAD (see headAsGet).
type headKey struct{}

// headAsGet answers HEAD on the GET routes, as net/http does: gin matches
// only the exact method, so HEAD /health was a 404. A HEAD without a route of
// its own is dispatched again as a GET, on a copy of the request: the server
// still sees the original HEAD and sends the headers without the body.
//
// It is the first handler of the engine, so it stops the 404 chain before
// any other middleware runs, and the GET chain runs once. A HEAD route the
// app declares itself is used as is.
func headAsGet(engine *gin.Engine) Middleware {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodHead || c.FullPath() != "" {
			c.Next()
			return
		}
		c.Abort()
		get := c.Request.Clone(context.WithValue(c.Request.Context(), headKey{}, true))
		get.Method = http.MethodGet
		c.Request = get
		engine.HandleContext(c)
	}
}

// method is the method the client sent: HEAD for a GET that answers a HEAD.
func method(r *http.Request) string {
	if head, _ := r.Context().Value(headKey{}).(bool); head {
		return http.MethodHead
	}
	return r.Method
}
