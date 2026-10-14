package preas

import (
	"github.com/prea-framework/prea"
	"github.com/prea-framework/prea/core"
	"github.com/prea-framework/prea/example/preas/auth"
	"github.com/prea-framework/prea/example/preas/user"
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

// Logging left empty takes logging.level and logging.format from the
// profile: debug in dev, json in prod.
var Logging = prea.Logging{}

var Middlewares = []any{
	core.RequestLogger,
	core.Recovery,
}
