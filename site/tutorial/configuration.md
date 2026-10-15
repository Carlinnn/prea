# 7. Configuration

Your app will run in three places: on your machine, in the tests and in production. Each needs different settings, like another database, other logs or another key. In this chapter you see how **profiles** handle that.

## One base, one file per profile

```
config/
├── application.yaml         shared by every profile
├── application-dev.yaml     your machine
├── application-tests.yaml   the tests
└── application-prod.yaml    production
```

Preá always loads `application.yaml`, then the active profile's file on top of it. A profile's file only holds what is different.

Look at what the project already has:

::: code-group

```yaml [application-dev.yaml]
logging:
  level: debug
database:
  migrations:
    from_models: true # a model change becomes a migration at startup...
    on_start: true # ...that is applied right away
```

```yaml [application-tests.yaml]
server:
  mode: test
logging:
  level: error
database:
  url: ${TEST_DATABASE_URL}
  schema: create-drop
```

```yaml [application-prod.yaml]
server:
  mode: release
logging:
  format: json
```

:::

- **dev** shows detailed logs and writes migrations from your models, as you've seen.
- **tests** uses the test database, and `create-drop` gives every test run fresh, empty tables.
- **prod** turns off debug output and writes logs as JSON, which log collectors read.

## A key for the tests

The tests will need an API key too, and a fixed one is fine there. Add it to `config/application-tests.yaml`:

```yaml
todo:
  api_key: tests-key-0123456789
```

The tests don't depend on anyone's `.env`, so they behave the same on every machine and in CI.

## Pick a profile

The active profile comes from `PREA_PROFILE`, which your `.env` sets to `dev`. The CLI can override it:

```bash
prea run --profile prod
```

```json
{"level":"INFO","msg":"prea: config loaded","profile":"prod"}
{"level":"INFO","msg":"prea: listening","addr":"[::]:8000"}
{"level":"SUCCESS","msg":"request","method":"GET","path":"/health","status":200,"latency":440708,"ip":"::1"}
```

The same app, now with JSON logs and in release mode. Stop it with Ctrl+C.

## What goes where

| Kind of setting | Where |
|---|---|
| the same everywhere, like the port | `application.yaml` |
| different per environment, like the log format | `application-<profile>.yaml` |
| secret, or different per machine, like passwords and keys | an environment variable, referenced as `${NAME}` |
| part of the code, like which preás exist | `preas/preas.go` |

Secrets never go in a YAML file. The files are committed, and the environment isn't.

## What you learned

- `application.yaml` is the base, and each profile's file adds what is different.
- `PREA_PROFILE`, or `--profile`, picks the profile.
- Secrets come from the environment with `${NAME}`.

[Next: testing →](./testing)
