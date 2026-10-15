# Deploying

A Preá app is a single binary, plus its configuration and migrations. This page shows what to build, what to ship and how to run it in production.

::: info What you'll learn
- How to build the binary
- What travels with it
- How to configure and run it in production
- How migrations fit into a release
:::

## Build

```bash
prea build
```

The binary goes to `bin/<project name>`. Go's build flags pass through, and so do `GOOS` and `GOARCH` for another platform:

```bash
GOOS=linux GOARCH=amd64 prea build -trimpath -ldflags="-s -w"
```

## What to ship

```
release/
├── shop            the binary
├── config/         the YAML files
└── migrations/     the SQL files
```

The app reads `config/` and `migrations/` from the folder it runs in, so keep the three together.

## Configure production

The `prod` profile of a new project switches the server to `release` mode and the logs to JSON:

```yaml
# config/application-prod.yaml
server:
  mode: release
logging:
  format: json
```

Everything secret comes from the environment:

```bash
PREA_PROFILE=prod
DATABASE_URL=postgres://...
JWT_SECRET=...          # if the app uses tokens
```

A checklist for `application-prod.yaml`:

- `database.schema: validate`, so the app never runs against tables it doesn't expect;
- `database.migrations.from_models: false`, so production never writes migration files;
- `server.trusted_proxies`, if a load balancer sits in front of the app ([HTTP security](./http-security#behind-a-proxy));
- the [server timeouts](/reference/configuration#server), if you stream or serve large downloads.

## Migrations in a release

Pick one of the two:

- **At startup:** `database.migrations.on_start: true`. Every instance tries, and a lock lets only one of them migrate.
- **As a deploy step:** run `prea migrate --profile prod up` before starting the new version, and leave `on_start` off.

Either way, `schema: validate` checks the result before the app serves.

## Run

```bash
PREA_PROFILE=prod ./shop
```

The app stops gracefully on `SIGTERM`, which is what deploy platforms and orchestrators send. It finishes the requests in flight, up to `server.shutdown_timeout`, before it exits. Set your platform's grace period a bit longer than that.

A container image needs nothing more than the binary and the two folders:

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN go install github.com/prea-framework/prea/cmd/prea@latest && prea build -trimpath

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /src/bin/shop ./
COPY --from=build /src/config ./config
COPY --from=build /src/migrations ./migrations
ENV PREA_PROFILE=prod
CMD ["./shop"]
```

## Next steps

- [Configuration keys](/reference/configuration)
- [CLI reference](/reference/cli)
