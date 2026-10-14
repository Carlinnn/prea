package auth

import (
	"github.com/prea-framework/prea/core"
	"github.com/prea-framework/prea/example/middleware"
	"github.com/prea-framework/prea/example/preas/user"
)

var Module = &core.Module{
	Name:    "auth",
	Imports: []*core.Module{user.Module},
	Providers: []any{
		Service{},
		middleware.Auth{},
		Handler{},
	},
}
