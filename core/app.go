package core

import (
	"context"
	"errors"
	"slices"

	"prea/config"
	"prea/di"

	"github.com/gin-gonic/gin"
)

// Router is implemented by components that expose HTTP routes.
type Router interface {
	Register(r gin.IRouter)
}

// App wires modules through the di container and serves their routes.
// Security runs first, then Middlewares, on every request; Globals are
// injectable in every module.
//
// Config fills the value:"key" fields, is injectable as *config.Config and
// picks the modules: one with Profiles loads only under one of them.
//
// Middlewares (and Security.Middlewares) take functions and components, like
// Module.Middlewares. Global components live in an implicit "app" module that
// imports every module, so they can inject Globals and the modules' exports.
type App struct {
	Modules     []*Module
	Security    Security
	Server      Server
	Middlewares []any
	Globals     []any
	Config      *config.Config

	// Setup runs after the modules load and before any Init, with every
	// component; teardown runs after they close. The database uses it to
	// migrate and sync the schema: Setup: database.Setup(pool, opts).
	Setup func(ctx context.Context, components []any) (teardown func(context.Context) error, err error)
}

// Build loads the modules, runs Setup and the components' Init (dependencies
// first) and mounts their routes. closer runs their Close in reverse order,
// then Setup's teardown; call it once the engine stops serving. Globals are
// not closed: their owner does it.
func (a App) Build(ctx context.Context) (engine *gin.Engine, closer func(context.Context) error, err error) {
	modules := slices.DeleteFunc(slices.Clone(a.Modules), func(m *Module) bool {
		return len(m.Profiles) > 0 && !slices.Contains(m.Profiles, a.Config.Profile())
	})
	root := &Module{Name: "app", Imports: modules, Middlewares: slices.Concat(a.Security.Middlewares, a.Middlewares)}
	globals := a.Globals
	if a.Config != nil {
		globals = append(slices.Clip(globals), a.Config)
	}
	c := di.New(globals...)
	if err := c.Load(root); err != nil {
		return nil, nil, err
	}
	teardown := func(context.Context) error { return nil }
	if a.Setup != nil {
		var components []any
		for _, m := range c.Modules() {
			components = append(components, c.Providers(m)...)
		}
		if teardown, err = a.Setup(ctx, components); err != nil {
			return nil, nil, err
		}
	}
	closer = func(ctx context.Context) error { return errors.Join(c.Close(ctx), teardown(ctx)) }

	if err := c.Init(ctx); err != nil {
		return nil, nil, errors.Join(err, teardown(ctx))
	}
	engine, err = a.engine(c, root)
	if err != nil {
		return nil, nil, errors.Join(err, closer(ctx))
	}
	return engine, closer, nil
}

func (a App) engine(c *di.Container, root *Module) (*gin.Engine, error) {
	engine := gin.New()
	if err := a.Security.apply(engine); err != nil {
		return nil, err
	}
	global, err := root.middlewares(c.Providers(root))
	if err != nil {
		return nil, err
	}
	engine.Use(global...)

	for _, m := range c.Modules() {
		if m == root {
			continue
		}
		group := engine.Group("")
		if mod, ok := m.(*Module); ok {
			handlers, err := mod.middlewares(c.Providers(m))
			if err != nil {
				return nil, err
			}
			group.Use(handlers...)
		}
		for _, component := range c.Providers(m) {
			if router, ok := component.(Router); ok {
				router.Register(group)
			}
		}
	}
	return engine, nil
}
