# Preás

A preá is one feature of your app. It is a folder in `preas/` with a `module.go` that declares what the feature is made of. Your app is the list of its preás.

::: info What you'll learn
- How to create and declare a preá
- How one preá uses another: imports and exports
- How to turn a preá on only in some environments
:::

## Create a preá

```bash
prea g p task
```

```
preá "task" created at preas/task and registered in preas/preas.go
```

The CLI creates `preas/task/` with a module, a handler, a service, a repository, a model and two request bodies, and adds `task.Module` to `preas/preas.go`. The new preá already answers `GET`, `POST`, `PATCH` and `DELETE` on `/task`.

## The declaration

Open `preas/task/module.go`:

```go
var Module = &core.Module{
	Name: "task",
	Providers: []any{
		Repository{},
		di.Export(Service{}),
		Handler{},
	},
}
```

| Field | What it says |
|---|---|
| `Name` | the preá's name, used in error messages |
| `Providers` | the components of the preá: the structs Preá creates for it |
| `Imports` | other preás whose exported components this one uses |
| `Middlewares` | middlewares that run on every route of this preá |
| `Profiles` | the profiles where this preá is part of the app (empty means always) |

The components are written as **values**: `Service{}`, not `&Service{}`. A module is only a description. Preá creates the real instances when the app starts. [Components & injection](./dependency-injection) explains why.

## Use another preá

Preás are private by default. A component of one preá cannot see the components of another, unless two things happen:

1. The owner **exports** the component with `di.Export`.
2. The user **imports** the owner's module.

```go{5}
// preas/list/module.go: the list preá shares its service
var Module = &core.Module{
	Name: "list",
	Providers: []any{
		Repository{},
		di.Export(Service{}),
		Handler{},
	},
}
```

```go{4}
// preas/task/module.go: the task preá uses it
var Module = &core.Module{
	Name:      "task",
	Imports:   []*core.Module{list.Module},
	Providers: []any{Repository{}, Service{}, Handler{}},
}
```

Now the task service can ask for the list service:

```go
type Service struct {
	Repo  *Repository   `inject:""`
	Lists *list.Service `inject:""` // exported by the list preá
}
```

Keep what you export small. A preá that exports only its service can change its repository and its table freely, because no other preá depends on them.

::: warning Import cycles
If `task` imports `list` and `list` imports `task`, the app won't start: `di: import cycle at module task`. When two features need each other, let one of them declare an **interface** with what it needs, and let the other implement it. [Components & injection](./dependency-injection#interfaces) shows how.
:::

## Register a preá

A preá is part of the app when it is in `Modules`, in `preas/preas.go`:

```go
var Modules = []*core.Module{
	health.Module,
	task.Module,
	list.Module,
}
```

The order doesn't matter. Preá loads each imported preá before the one that imports it, and it loads a preá only once, even when several preás import it.

## Turn a preá on per profile

`Profiles` limits a preá to some environments. This one only exists in `dev`:

```go
var Module = &core.Module{
	Name:      "debug",
	Providers: []any{Handler{}},
	Profiles:  []string{"dev"},
}
```

In any other profile, the preá and its routes are simply not there. This applies to the preás listed in `Modules`. A preá that is only imported by others always comes along with them.

## Next steps

- [Components & injection](./dependency-injection): how the pieces of a preá find each other
- [Handlers](./handlers): give a preá its routes
