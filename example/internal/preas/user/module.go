package user

import (
	"github.com/prea-framework/prea/core"
	"github.com/prea-framework/prea/di"
	"github.com/prea-framework/prea/example/internal/middleware"
)

var Module = &core.Module{
	Name: "user",
	Providers: []any{
		Repository{},
		di.Export(Service{}),
		middleware.Auth{},
		Handler{},
	},
}
