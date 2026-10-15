# Persistence <span class="soon">Coming soon</span>

Preá is getting a persistence layer of its own. It will be the next way to work with data in a Preá app, built on everything the current database layer already does.

::: info What's coming
- **More databases** than PostgreSQL, behind the same models and repositories.
- **Relationships as structs.** A player has a `Team *Team`, not a `TeamID`. Saving a player with `Team{ID: 7}` links it to that team, and reading it brings the team back, without one query per row.
- **Many-to-many, cascades and lazy loading.** Tasks with `Tags []Tag`, saving a team together with its players, and relationships that load only when you use them.
- **PPRQL**, a query language for your models, so most queries need no hand-written SQL.
- **Specifications and pagination**, to build filters from small reusable pieces and page through results.
:::

## What carries over

The ideas of the [current database layer](/guide/models) are the starting point of the new one: a struct describes a table, a repository reads and writes it, and transactions travel in the context. What you learn today still applies.

## Until then

Everything you need to build an app with a database is already here:

- [Models & schema](/guide/models)
- [Migrations](/guide/migrations)
- [Repositories](/guide/repositories)
- [Transactions](/guide/transactions)

Follow the [What's next](/about/whats-next) page, or watch the [repository](https://github.com/prea-framework/prea), to know when it lands.
