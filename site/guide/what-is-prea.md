# What is Preá?

Preá is a web framework for Go. It helps you build HTTP APIs by splitting the application into small, self-contained pieces called **preás**, and by wiring those pieces together for you.

::: info What you'll learn
- The idea behind Preá, in one page
- The four ideas you will meet everywhere in these docs
- What Preá is good at, and what it leaves to you
:::

## The idea

Most applications start small and grow feature by feature. Preá gives each feature a home: a folder with its routes, its rules, its data access and its table. That folder is a **preá**.

```
preas/
├── task/      everything about tasks
├── list/      everything about lists
└── health/    GET /health
```

Inside a preá you write plain Go structs. You don't create them, and you don't pass them around by hand. You **declare** them, and each one **asks** for what it needs:

```go
var Module = &core.Module{
	Name:      "task",
	Providers: []any{Repository{}, Service{}, Handler{}},
}

type Service struct {
	Repo *Repository `inject:""` // "I need the task repository"
}
```

When the app starts, Preá reads these declarations, creates every piece, connects them and checks that nothing is missing. Only then does it start serving requests.

## Four ideas

These four ideas show up on almost every page.

**Preás.** A preá is one feature of your app, declared in `module.go`. A preá can use another one by **importing** it, and it decides what it shares by **exporting** it. [Learn about preás →](./preas)

**Components.** The structs a preá declares, like its handler, service and repository, are its components. Preá creates one of each and fills the fields you tag with `inject:""`. [Learn about components →](./dependency-injection)

**Errors at startup, not in production.** A missing dependency, a typo in a model tag or a required setting left empty: Preá finds these when the app starts and stops it with a clear message. A running app has already passed these checks.

**Configuration by profile.** One base file, `config/application.yaml`, plus one file per environment: `dev`, `prod` and `tests`. Secrets come from environment variables. [Learn about configuration →](./configuration)

## What you get

- **An HTTP layer** with routing, middlewares at three levels, a standard JSON reply format and errors that carry their own HTTP status.
- **A database layer** for PostgreSQL. Your models describe the tables, Preá generates and applies the migrations, and every model gets a ready-made repository with transactions.
- **Safe defaults:** CORS, CSRF protection, security headers, server timeouts and a graceful shutdown.
- **A CLI** that creates projects, generates preás, runs the app with hot reload, migrates the database, runs the tests and builds the binary.

## What Preá leaves to you

Preá is pre-alpha (`v0.0.x`). Some pieces are still on the way:

- **Authentication** has its building blocks, tokens and password hashing, but no ready-made login flow yet.
- **Databases:** PostgreSQL is the only one today. A new persistence layer with more databases is [coming](/persistence/).
- **Validation** uses the `binding` tags of the request structs. A dedicated validation layer is planned.

The [What's next](/about/whats-next) page lists what is coming.

## Next steps

- [Install the CLI](./installation)
- [Create and run your first app](./quick-start)
