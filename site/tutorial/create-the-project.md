# 1. Create the project

In this chapter you create the project, connect it to the database and run it.

## Create it

```bash
prea new todo --with-db
cd todo
```

```
preá "item" created at preas/item and registered in preas/preas.go
project todo created: cd todo && cp .env.example .env && prea dev
```

`--with-db` connects the project to a database and adds an example preá called `item`. You will write your own preás, so remove it before the first run:

```bash
rm -r preas/item
```

Then open `preas/preas.go` and delete the two lines that mention it:

```go
import (
	"github.com/prea-framework/prea"
	"github.com/prea-framework/prea/core"

	"todo/preas/health"
	"todo/preas/item" // [!code --]
)

var Modules = []*core.Module{
	health.Module,
	item.Module, // [!code --]
}
```

`Modules` is the list of features your app is made of. Right now, only `health` is left.

## Connect the database

The settings that change from one machine to another, like the database address, live in a `.env` file. Create it from the template:

```bash
cp .env.example .env
```

Open `.env` and point the two URLs at your databases:

```bash
PREA_PROFILE=dev
DATABASE_URL='postgres://postgres:postgres@localhost:5432/todo?sslmode=disable'
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/todo_test?sslmode=disable'
```

`PREA_PROFILE=dev` picks the **profile**, the set of settings for your machine. You'll meet the others in [chapter 7](./configuration).

## Run it

```bash
prea dev
```

```
level=INFO msg="prea: config loaded" profile=dev
level=INFO msg="database: the models and the database are in sync"
level=INFO msg="database: no migrations in migrations"
level=INFO msg="prea: listening" addr=[::]:8000
```

The app loaded the `dev` settings, connected to the database, found nothing to do there yet and started listening. In another terminal:

```bash
curl localhost:8000/health
```

```json
{
    "code": 1,
    "data": {
        "status": "ok"
    }
}
```

Leave `prea dev` running. It watches your files, and every time you save, it rebuilds and restarts the app.

## Look around

```
todo/
├── main.go            starts the app; you won't need to touch it
├── preas/
│   ├── preas.go       the list of preás, middlewares and protections
│   └── health/        the GET /health route
├── config/            the settings: a base file and one per profile
├── migrations/        empty for now
├── middleware/        empty for now
└── tests/             tests of the whole app
```

[Project structure](/guide/project-structure) describes every file, if you are curious.

## What you learned

- `prea new` creates a working project, and `prea dev` runs it and restarts it on every change.
- An app is a list of preás, in `preas/preas.go`.
- Settings that change per machine go in `.env`.

[Next: your first preá →](./first-prea)
