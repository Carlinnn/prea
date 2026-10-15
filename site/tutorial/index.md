# Build a todo API

In this tutorial you build a small but complete API for a todo app, one step at a time. Each chapter adds one idea, and each one ends with an app that runs.

::: info What you'll build
An HTTP API where you can:
- create, list, update, complete and delete **tasks**;
- group tasks into **lists**, and see the tasks of one list;
- protect every change with an **API key**;
- run **tests** against a real database;
- **build** a binary ready for production.
:::

## The chapters

| | Chapter | What you learn |
|---|---|---|
| 1 | [Create the project](./create-the-project) | `prea new`, `prea dev`, the project layout |
| 2 | [Your first preá](./first-prea) | generating a preá, its files and routes |
| 3 | [The model](./the-model) | describing a table with a struct, migrations |
| 4 | [Business rules](./business-rules) | services, errors with a status, partial updates |
| 5 | [A second preá](./second-prea) | imports, exports and relations between tables |
| 6 | [A middleware](./middleware) | components with settings and `Init` |
| 7 | [Configuration](./configuration) | profiles for dev, tests and production |
| 8 | [Testing](./testing) | testing a preá and the whole app |
| 9 | [Ship it](./ship-it) | building, migrating and running in production |

It takes about an hour. You can stop after any chapter and come back later.

## Before you start

You need:

- **Go 1.26 or newer** and the **`prea` CLI**. See [Installation](/guide/installation).
- **PostgreSQL**, with two empty databases, `todo` and `todo_test`. A container works well:

```bash
docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:17-alpine
docker exec postgres createdb -U postgres todo
docker exec postgres createdb -U postgres todo_test
```

- **A way to send HTTP requests.** The examples use `curl`, but any client works.

You don't need to know Preá. Some Go helps: structs, methods and errors.

::: tip The finished app
The code of the finished app lives in the Preá repository, in [`site/examples/todo`](https://github.com/prea-framework/prea/tree/main/site/examples/todo). The CI builds and tests it on every change, so it always works with the current version.
:::

[Start with chapter 1 →](./create-the-project)
