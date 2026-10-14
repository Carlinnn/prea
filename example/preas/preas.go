package preas

import (
	"time"

	"github.com/prea-framework/prea/core"
	"github.com/prea-framework/prea/example/internal/preas/auth"
	"github.com/prea-framework/prea/example/internal/preas/user"
)

var Modules = []*core.Module{
	user.Module,
	auth.Module,
}

var Security = core.Security{
	CORS:    &core.CORS{AllowOrigins: []string{"http://localhost:3000"}},
	CSRF:    &core.CSRF{},
	Headers: core.DefaultHeaders,
}

var Server = core.Server{
	ReadTimeout:     30 * time.Second,
	WriteTimeout:    30 * time.Second,
	ShutdownTimeout: 15 * time.Second,
}

var Middlewares = []any{
	core.RequestLogger,
	core.Recovery,
}
