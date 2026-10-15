# 6. A middleware

Anyone who can reach the API can delete your tasks. In this chapter you protect every change with an **API key**: reads stay open, and writes need the key. The check runs before every handler, as a **middleware**.

## Generate it

```bash
prea g m api_key
```

```
middleware APIKey created at middleware/api_key.go
```

A middleware runs around a handler. It can let the request continue, with `c.Next()`, or stop it, with `c.Abort()`.

## Read the key from the settings

The key is a setting: different on your machine, in the tests and in production. Add it to `config/application.yaml`:

```yaml
todo:
  api_key: ${TODO_API_KEY} # 16+ characters; empty leaves the writes open
```

`${TODO_API_KEY}` reads the environment variable. Add it to your `.env`, and to `.env.example` with an empty value, so others know it exists:

```bash
TODO_API_KEY=my-local-key-0123456789
```

The middleware is a **component**, like a service, so it can receive settings. Replace `middleware/api_key.go`:

<<< @/examples/todo/middleware/api_key.go#component

The `value:"todo.api_key"` tag fills the field from the configuration when the app starts.

## Check the key once, at startup

A key that is too short is a mistake you want to find before the app serves anyone. Components can have an `Init` method, which runs once at startup, after their fields are filled:

<<< @/examples/todo/middleware/api_key.go#init

If `Init` returns an error, the app doesn't start.

## Check every request

<<< @/examples/todo/middleware/api_key.go#handle

`subtle.ConstantTimeCompare` compares the keys without leaking, through timing, how much of a wrong key was right.

Add the imports the file uses: `context`, `crypto/subtle`, `errors` and `net/http`.

## Turn it on for the whole app

Middlewares in `preas.Middlewares`, in `preas/preas.go`, run on every request:

<<< @/examples/todo/preas/preas.go#middlewares

Import `"todo/middleware"` at the top of the file.

## Try it

```bash
curl -X POST localhost:8000/list -H 'Content-Type: application/json' -d '{"name": "Work"}'
```

```json
{
    "text": "missing or wrong X-API-Key",
    "code": 2
}
```

```bash
curl -X POST localhost:8000/list -H 'Content-Type: application/json' \
  -H 'X-API-Key: my-local-key-0123456789' -d '{"name": "Work"}'
# → 201
curl localhost:8000/list
# → 200: reads need no key
```

Now set `TODO_API_KEY=short` in `.env` and save. The restart fails, and the app doesn't serve with a weak key:

```
level=ERROR msg="prea: startup failed" error="di: *middleware.APIKey.Init: todo.api_key must have at least 16 characters"
```

Put the long key back before you go on.

## What you learned

- A middleware runs before handlers and can stop a request.
- A component can read settings with `value:""` and check them in `Init`.
- Middlewares in `preas.Middlewares` cover the whole app.

[Next: configuration →](./configuration)
