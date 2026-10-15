// Package preas declares the app: its modules and what runs around them.
// main.go passes these variables to prea.Run; prea g p registers new preás
// in Modules.
package preas

import (
	"github.com/prea-framework/prea"
	"github.com/prea-framework/prea/core"

	"todo/middleware"
	"todo/preas/health"
	"todo/preas/list"
	"todo/preas/task"
)

// #region modules
var Modules = []*core.Module{
	health.Module,
	task.Module,
	list.Module,
}

// #endregion modules

// #region middlewares
// Middlewares run on every request, after Security.
var Middlewares = []any{
	core.RequestLogger,
	core.Recovery,
	middleware.APIKey{},
}

// #endregion middlewares

// Security runs before everything else. Add the origins of your frontend to
// CORS: &core.CORS{AllowOrigins: []string{"http://localhost:3000"}}.
var Security = core.Security{
	CSRF:    &core.CSRF{},
	Headers: core.DefaultHeaders,
}

// Logging left empty takes logging.level and logging.format from the
// profile: debug in dev, json in prod.
var Logging = prea.Logging{}
