# Components & injection

A handler needs a service, a service needs a repository and a repository needs the database. Someone has to create each piece and hand it to the pieces that use it. In Preá, you don't write that code. You **declare** the pieces, and each piece **asks** for what it needs.

::: info What you'll learn
- What a component is and how you declare one
- How a component asks for another with `inject:""`
- Where Preá looks for what a component asks for
- How components start, stop and read settings
- What to do when something is missing
:::

## Components

A **component** is a struct that a preá declares in its `Providers`:

```go
var Module = &core.Module{
	Name:      "task",
	Providers: []any{Repository{}, Service{}, Handler{}},
}
```

When the app starts, Preá creates **one instance** of each component. The handler, the service and the repository of `task` each exist once, and every piece that uses them gets that same instance.

## Asking for something

A component asks for another one with a field tagged `inject:""`:

<<< @/examples/todo/preas/task/service.go#service

When Preá creates the service, it sees the two tagged fields, finds a `*Repository` and a `*list.Service`, and fills them in. By the time your code runs, both fields are ready.

Three small rules:

- **The field is a pointer** to the component, `*Repository` and not `Repository`, or an [interface](#interfaces) it implements.
- **The field is exported**: `Repo`, not `repo`. Preá can't fill a field it can't see.
- **The tag is `inject:""`**, with the empty quotes. A bare `inject` is not a valid Go tag, and Preá rejects it at startup.

## Where Preá looks

Preá looks for a field's component in three places, in this order:

```
┌─ task preá ────────────────────────────┐
│  1. its own components                 │  Repository, Service, Handler
├────────────────────────────────────────┤
│  2. what its imports export            │  list.Service
├────────────────────────────────────────┤
│  3. the globals                        │  the database pool, the config
└────────────────────────────────────────┘
```

1. **Its own components.** Anything the preá declares in `Providers`.
2. **What its imports export.** Only the components that the imported preás wrap in `di.Export`. Everything else stays private. [Preás](./preas#use-another-preá) shows how imports work.
3. **The globals.** Values shared by the whole app. `prea.Run` adds two: the database pool (`*pgxpool.Pool`), when `database.url` is set, and the configuration (`*config.Config`).

The first place with a match wins. If no place has one, the app doesn't start, and the error names the preá, the struct and the field:

```
di: module task: Service.Lists: no provider for *list.Service
```

## Interfaces

A field can ask for an interface instead of a concrete type. Preá fills it with the component that implements it:

```go
// preas/task: what a task needs from whoever owns the lists
type ListChecker interface {
	Exists(ctx context.Context, id string) (bool, error)
}

type Service struct {
	Lists ListChecker `inject:""`
}
```

This is the way out of an import cycle. The `task` preá declares what it needs, and the `list` preá implements it without `task` importing anything back.

If two visible components implement the same interface, Preá doesn't pick one at random. It stops with an error and asks you to name them.

## Named components

Sometimes you have two components of the same type, like two payment gateways. Give each a name with `di.Named`, and pick one with `inject:"name"`:

```go
var Module = &core.Module{
	Name: "payment",
	Providers: []any{
		di.Named("card", Gateway{}),
		di.Named("pix", Gateway{}),
		Service{},
	},
}

type Service struct {
	Card *Gateway `inject:"card"`
	Pix  *Gateway `inject:"pix"`
}
```

- `inject:"card"` looks only at the components named `card`.
- `inject:""` prefers the one **without** a name. If there is only one candidate, named or not, it uses that one.
- `di.Named` and `di.Export` combine in any order: `di.Export(di.Named("card", Gateway{}))`.

## Start and stop

A component can do work when the app starts and when it stops. Add an `Init` method, a `Close` method, or both:

<<< @/examples/todo/middleware/api_key.go#init

```go
// Close runs when the app stops, after the last request finished.
func (m *Mailer) Close(ctx context.Context) error {
	return m.client.Quit()
}
```

- **`Init(ctx) error`** runs after every field is filled and before the app accepts requests. Use it to check settings, open connections or warm a cache. If it returns an error, the app doesn't start, and the components that already started are closed.
- **`Close(ctx) error`** runs on shutdown, after the in-flight requests finish and before the database closes. A plain `Close() error` works too.
- **Order:** if A uses B, B starts before A and stops after A. You never meet a dependency that isn't ready.

::: details Why `Init` and not a constructor?
Preá creates your components, so there is no `NewService` function for you to write. `Init` takes its place. Preá calls it at the right moment, when every dependency already exists and has started.
:::

## Settings

A field tagged `value:"key"` gets a setting from the [configuration](./configuration):

<<< @/examples/todo/middleware/api_key.go#component

```yaml
# config/application.yaml
todo:
  api_key: ${TODO_API_KEY}
```

The value is converted to the field's type: `"8000"` to an `int`, `"true"` to a `bool`, `"30s"` to a `time.Duration`. A whole block fills a struct. A key that isn't in the configuration leaves the field as it is.

## One instance, many requests

Each component exists once, and it serves every request at the same time, each request in its own goroutine. So:

- **Don't keep request data in a component's fields.** Two requests would overwrite each other. Use local variables, or the request's `gin.Context`.
- **Shared data that changes needs protection:** a `sync.Mutex` or the `sync/atomic` types.

```go
type Service struct {
	Repo     *Repository `inject:""`
	lastUser string      // ✗ every request writes here at the same time
}
```

## Declared as values

You write `Service{}` in `Providers`, and Preá creates its own copy when the app starts. The declaration itself is never changed. This keeps apps apart: two apps in the same process, like two tests running at once, each get their own components, and they never share state by accident.

Writing `&Service{}` stops the app with a reminder:

```
di: module task: provider *task.Service must be declared as a value, task.Service{}: each container creates its own instance
```

## Fields you fill yourself

Preá only fills fields that are empty. A field you fill in the declaration stays as it is, which is handy in tests:

```go
// A test where the service talks to a fake repository.
Providers: []any{Service{Repo: &fakeRepository{}}, Handler{}}
```

The copy is shallow: a pointer you put there is shared by every app built from that declaration.

## Embedded structs

Fields of an embedded struct are filled too. That is how a repository gets its database connection without you writing a field for it:

```go
type Repository struct {
	database.BaseRepository[Task] // has a DB *pgxpool.Pool `inject:""` inside
}
```

## When something goes wrong

Every one of these errors happens at startup, with the preá, the struct and the field in the message.

::: details `no provider for *list.Service`
Nothing visible provides that type. Check that the component is in some `Providers`. If it belongs to another preá, check that the preá is in `Imports` and that the component is wrapped in `di.Export`.
:::

::: details `no provider for *pgxpool.Pool: set database.url in the profile`
A repository needs the database, and the active profile has no `database.url`. Set it in `config/application.yaml` or in the profile's file.
:::

::: details `2 providers for *Gateway: name them with di.Named and pick one with inject:"<name>"`
Two components match the field. Name them with `di.Named`, and pick one with `inject:"name"`.
:::

::: details `duplicate provider task.Service`
The same type appears twice in one preá's `Providers`, without different names.
:::

::: details `Service.repo must be exported to be injected`
Rename the field so it starts with a capital letter.
:::

::: details `use the tag inject:"" (a bare inject is not a valid Go struct tag)`
Write `inject:""`, with the quotes.
:::

::: details `import cycle at module task`
Two preás import each other, directly or through others. Break the cycle with an [interface](#interfaces).
:::

::: details `di: *middleware.APIKey.Init: ...`
A component's `Init` returned an error. The rest of the message is that error.
:::

## Next steps

- [Handlers](./handlers): the component that answers HTTP
- [Configuration](./configuration): where `value:""` gets its settings
