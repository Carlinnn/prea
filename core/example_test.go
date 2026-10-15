package core_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/prea-framework/prea/core"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/hello", func(c *gin.Context) {
		core.Reply(c, http.StatusOK, core.Text("hello"))
	})
}

// An App builds its modules and mounts the routes of every core.Router.
func ExampleApp() {
	gin.SetMode(gin.ReleaseMode)
	app := core.App{
		Modules: []*core.Module{{Name: "hello", Providers: []any{Handler{}}}},
	}

	engine, closer, err := app.Build(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer closer(context.Background())

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/hello", nil))
	fmt.Println(w.Code, w.Body)
	// Output: 200 {"text":"hello","code":1}
}

// Error replies with the status the error carries.
func ExampleError() {
	gin.SetMode(gin.ReleaseMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	core.Error(c, core.NewError(http.StatusForbidden, "not the owner"))
	fmt.Println(w.Code, w.Body)
	// Output: 403 {"text":"not the owner","code":2}
}
