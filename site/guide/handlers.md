# Handlers

The handler is the part of a preá that speaks HTTP. It declares the routes, reads the request, calls the service and writes the reply. It doesn't make decisions; that is the service's job.

::: info What you'll learn
- How a handler declares its routes
- How to read path parameters, query strings and JSON bodies
- How to reply, and how to reply with an error
:::

## Declare routes

Any component with a `Register` method gets its routes mounted when the app starts:

<<< @/examples/todo/preas/task/handler.go#register

`r.Group("/task")` gives every route of the group the `/task` prefix. The methods (`h.list`, `h.create`...) are plain functions that receive the request's context.

Every `GET` route also answers `HEAD`, with the same headers and no body, as uptime checks and load balancers expect. The whole chain runs once, as for the `GET`. A `HEAD` route you declare yourself takes its place.

The line below the struct makes the compiler check that `Register` has the right signature:

```go
var _ core.Router = (*Handler)(nil)
```

Without it, a typo in `Register` would compile, and the routes would silently not exist.

::: tip Routing and the request context come from Gin
Preá serves HTTP with [Gin](https://gin-gonic.com). `r` is a `gin.IRouter`, and `c` is a `*gin.Context`, so all of Gin's routing and request helpers are available.
:::

## Read the request

```go
func (h *Handler) get(c *gin.Context) {
	id := c.Param("id")       // /task/:id
	listID := c.Query("list") // /task?list=...
	// ...
}
```

To read a JSON body, bind it to a request struct from `dtos/`:

<<< @/examples/todo/preas/task/handler.go#create

The `binding:"required"` tag on a field makes the bind fail when the field is missing:

<<< @/examples/todo/preas/task/dtos/create_request.go#create

## Pass the request's context down

Call the service with `c.Request.Context()`, not `context.Background()`:

```go
items, err := h.Service.List(c.Request.Context(), c.Query("list"))
```

That context ends when the client gives up or the server shuts down. Every database query made with it stops too, instead of running for nobody.

## Reply

Replies use one JSON shape across the whole app. `core.Reply` builds it:

```go
core.Reply(c, http.StatusOK, core.Data(items))      // {"code": 1, "data": [...]}
core.Reply(c, http.StatusCreated, core.Data(task))  // {"code": 1, "data": {...}}
c.Status(http.StatusNoContent)                       // 204, no body
```

When something fails, hand the error to `core.Error`. It picks the HTTP status from the error and writes the reply:

```go
if err != nil {
	core.Error(c, err)
	return
}
```

[Errors & responses](./errors-and-responses) explains both.

## Keep handlers thin

A handler that only translates between HTTP and the service is easy to read, and the rules it would otherwise hold stay in one place. Whenever a handler starts to decide something, like "is this title too long?" or "does this list exist?", move that decision into the [service](./services).

## Next steps

- [Services](./services): where the decisions go
- [Middlewares](./middlewares): run code before a handler
