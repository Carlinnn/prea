# Middlewares

A middleware is code that runs around a handler: before it, after it, or instead of it. Logging every request, checking an API key and limiting the request rate are all middlewares.

::: info What you'll learn
- The two kinds of middleware: functions and components
- The three places a middleware can run: the whole app, a preá or a route
- The order in which everything runs
:::

## A middleware

```go
func Timing(c *gin.Context) {
	start := time.Now()   // before the handler

	c.Next()              // run the rest: other middlewares, then the handler

	logger.Info("took", "path", c.Request.URL.Path, "time", time.Since(start)) // after
}
```

`c.Next()` continues the chain. To stop it, for example because a check failed, reply and call `c.Abort()`:

<<< @/examples/todo/middleware/api_key.go#handle

## Two kinds

**Functions** are the simplest kind: a `func(c *gin.Context)`, also known as `core.Middleware`. They can't ask for components.

```bash
prea g m timing --func
```

**Components** are structs with a `Handle(c *gin.Context)` method. Preá creates them like any other component, so they can use `inject:""` and `value:""`, and have `Init` and `Close`:

```bash
prea g m api_key
```

<<< @/examples/todo/middleware/api_key.go#component

Pick a function when the middleware needs nothing, and a component when it needs a service, the database or a setting.

## Three places

### The whole app

Middlewares in `preas.Middlewares` run on every request:

<<< @/examples/todo/preas/preas.go#middlewares

A component here can ask for the globals and for anything any preá exports.

### One preá

Middlewares in a module's `Middlewares` run on every route of that preá:

```go
var Module = &core.Module{
	Name:        "admin",
	Providers:   []any{Handler{}},
	Middlewares: []any{middleware.Timing, middleware.AdminOnly{}},
}
```

A component listed here is created for the preá automatically. You don't need to repeat it in `Providers`.

### One route

Put the middleware before the handler, in the route itself:

```go
g.GET("/report", middleware.Timing, h.report)
```

For a component, declare it in `Providers`, ask for it in the handler and use its `Handle`:

```go
var Module = &core.Module{
	Name:      "report",
	Providers: []any{middleware.AdminOnly{}, Handler{}},
}

type Handler struct {
	AdminOnly *middleware.AdminOnly `inject:""`
}

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/report", h.AdminOnly.Handle, h.report)
}
```

::: warning Always let Preá create components
`(&middleware.AdminOnly{}).Handle` compiles, but that instance never had its fields filled, and it panics on the first request. Pass `middleware.AdminOnly{}` to a list, or inject it.
:::

## The order

For each request, the chain runs in this order:

```
security  →  app middlewares  →  preá middlewares  →  route middlewares  →  handler
```

Inside each list, middlewares run in the order you wrote them. [HTTP security](./http-security) runs first, so a blocked request never reaches your code.

## Next steps

- [App lifecycle](./lifecycle): what happens from start to stop
- [HTTP security](./http-security): CORS, CSRF and security headers
