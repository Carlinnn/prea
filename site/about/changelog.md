# Changelog

What each version brought. Preá is pre-alpha, and the API may change between versions until `v1`.

Each module has its own version. The CLI and the root module follow the newest one.

## v0.0.2

*October 2026*

**Run an app from its configuration**
- `prea.Run` starts an app from `config/` and the active profile: logs, database, migrations, token secret and server.
- `prea.Test` builds the app for a test, with the `tests` profile.
- The server timeouts come from `server.*` in the configuration.
- The token secret is optional: an app that doesn't use tokens doesn't need it.

**CLI**
- `prea new` creates a project, with or without a database (`--with-db`).
- `prea dev` runs the app and restarts it on every change.
- `prea build` builds the binary into `bin/`.
- `prea tests` runs the tests with the `tests` profile.

**Components**
- A missing dependency is a typed error, `*di.MissingError`, and a repository without a database says to set `database.url`.

## v0.0.1

*October 2026*

The first version.

- **Preás and components**, with `inject:""`, `value:""`, named components, exports and imports, and `Init`/`Close`.
- **Configuration** with profiles, `.env` and environment variables.
- **HTTP**: routes, middlewares at three levels, the standard reply, errors with a status, CORS, CSRF, security headers and graceful shutdown.
- **Database** for PostgreSQL: models described by tags, repositories, named-parameter queries, transactions, schema checks and migrations written from the models.
- **Security**: signed tokens and password hashing.
- **Logs**: four levels, text or JSON, one line per request.
- **CLI**: `prea run`, `prea generate` and `prea migrate`.
- **Modules**: the framework is split into Go modules, each with its own version.
