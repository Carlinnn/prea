# Configuration

An app runs in more than one place: your machine, the tests, production. The code stays the same, and the settings change. In Preá, the settings live in YAML files under `config/`, one per environment, and secrets come from environment variables.

::: info What you'll learn
- How the base file and the profile files combine
- How to pick the active profile
- How to bring in environment variables and secrets
- How a component reads a setting
:::

## Base and profiles

```
config/
├── application.yaml         always loaded
├── application-dev.yaml     on your machine
├── application-prod.yaml    in production
└── application-tests.yaml   in the tests
```

`application.yaml` holds the settings shared by every environment. A **profile** is a name, like `dev`, and its file holds only what is different there. Preá loads the base, then the active profile's file on top of it. A key in the profile file wins, and everything else comes from the base.

::: code-group

```yaml [application.yaml]
server:
  port: 8000
  mode: debug
logging:
  level: info
  format: text
```

```yaml [application-prod.yaml]
server:
  mode: release
logging:
  format: json
```

```yaml [result with prod]
server:
  port: 8000       # from the base
  mode: release    # from prod
logging:
  level: info      # from the base
  format: json     # from prod
```

:::

## Pick the profile

The active profile comes from the `PREA_PROFILE` environment variable. The CLI sets it for you:

| You run | Profile |
|---|---|
| `prea dev` | `dev`, unless you pass `--profile` |
| `prea run --profile prod` | `prod` |
| `prea tests` | `tests` |
| `./bin/shop` | whatever `PREA_PROFILE` says |

Without a profile, only the base file is loaded. A profile file that doesn't exist is skipped.

## Environment variables and secrets

A value can reference an environment variable with `${NAME}`:

```yaml
database:
  url: ${DATABASE_URL}
security:
  jwt:
    secret: ${JWT_SECRET}
```

Only the variables you reference come in, so nothing from the environment leaks into the settings by surprise. Keep every secret in the environment, never in a committed YAML file.

**`.env`.** On your machine, put the variables in a `.env` file at the project root. Preá reads it at startup. It never replaces a variable that is already set, so the real environment always wins. `.env` is in `.gitignore`; `.env.example` is its committed template.

```bash
# .env
PREA_PROFILE=dev
DATABASE_URL='postgres://postgres:postgres@localhost:5432/shop?sslmode=disable'
```

## Missing values

A key that isn't in any file, or a `${NAME}` whose variable isn't set, leaves the setting empty: zero for numbers, `""` for strings, `nil` for pointers. Preá doesn't invent defaults for your own settings, so check them where they are used. A component's `Init` is the right place:

<<< @/examples/todo/middleware/api_key.go#init

Two settings are required, and the app won't start without them: `server.port` and `server.mode`. [Configuration keys](/reference/configuration) lists every key Preá reads.

## Read a setting in a component

Tag a field with `value:"key"`:

```go
type Mailer struct {
	Addr    string        `value:"mail.addr"`
	Timeout time.Duration `value:"mail.timeout"` // "30s" in the YAML
}
```

```yaml
mail:
  addr: smtp.example.com:587
  timeout: 30s
```

The value is converted to the field's type: `"8000"` to `int`, `"true"` to `bool`, `"30s"` to `time.Duration`.

**A whole block.** Point the tag at a block, and describe the block with a struct and `yaml` tags:

```go
type Mailer struct {
	Mail MailSettings `value:"mail"`
}

type MailSettings struct {
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout"`
}
```

**The whole configuration.** A component can also ask for `*config.Config` and read any key, or check the active profile:

```go
type Service struct {
	Config *config.Config `inject:""`
}

func (s *Service) debugging() bool {
	return s.Config.Profile() == "dev"
}
```

## Settings in code

Some settings are part of the code, not of the environment, and they live in `preas/preas.go`: the preás, the middlewares, the [HTTP security](./http-security) (like the allowed origins) and the [logging](./logging). If a value should be the same everywhere your app runs, put it in code. If it changes per environment, put it in the configuration.

## Next steps

- [Middlewares](./middlewares)
- [Configuration keys](/reference/configuration): every key Preá reads
