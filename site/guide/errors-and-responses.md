# Errors & responses

Every reply of a Preá app has the same shape, so a client can handle all of them the same way. Errors carry their own HTTP status, so the code that finds a problem also decides how it is reported.

::: info What you'll learn
- The shape of every reply
- How `core.Reply` builds a reply from parts
- How an error picks its HTTP status
- What the client sees when something unexpected fails
:::

## The reply shape

```json
{
    "text": "task created",
    "code": 1,
    "data": { "id": "…", "title": "Buy coffee" }
}
```

| Field | Meaning |
|---|---|
| `code` | always there: `1` success, `2` warning, `3` danger |
| `text` | a message for people; left out when empty |
| `data` | the payload; left out when empty |

## Build a reply

`core.Reply` takes the status and any combination of parts:

```go
core.Reply(c, 200, core.Data(tasks))                         // {"code":1,"data":[…]}
core.Reply(c, 201, core.Text("created"), core.Data(task))    // {"text":"created","code":1,"data":{…}}
core.Reply(c, 200, core.Text("saved"))                       // {"text":"saved","code":1}
core.Reply(c, 200, core.Text("saved, with a catch"), core.Code(core.WarningCode))
```

Without `core.Code`, the code follows the HTTP status: below 400 is `core.SuccessCode` (1), below 500 is `core.WarningCode` (2), and anything else is `core.DangerCode` (3).

When you need a different shape, like for a webhook that expects its own format, `core.Response` sends any value as JSON:

```go
core.Response(c, 200, gin.H{"ok": true})
```

In `debug` mode the JSON is indented, to read it easily; in `release` it is compact.

## Errors with a status

A service returns an error made with `core.NewError`:

```go
return core.NewError(http.StatusForbidden, "this list belongs to someone else")
```

The handler passes any error to `core.Error`:

```go
if err != nil {
	core.Error(c, err)
	return
}
```

`core.Error` looks for the status in the error and replies with it and its message:

```json
{
    "text": "this list belongs to someone else",
    "code": 2
}
```

## Where the status comes from

`core.Error` walks the error chain and uses the first error that has an `HTTPStatus() int` method:

| Error | Status |
|---|---|
| `core.NewError(status, message)` | `status` |
| `database.ErrInvalidInput` | 400 |
| `security.ErrInvalidToken` | 401 |
| `database.ErrNotFound` | 404 |
| `database.ErrConflict` | 409 |
| your own type with `HTTPStatus() int` | what it returns |
| anything else | 500, with the text `internal server error` |

Your own error types work the same way:

```go
type QuotaError struct{ Limit int }

func (e QuotaError) Error() string    { return fmt.Sprintf("you can have %d tasks at most", e.Limit) }
func (e QuotaError) HTTPStatus() int  { return http.StatusTooManyRequests }
```

## What the client sees, and what the log sees

Only the message of the error with the status reaches the client. Anything wrapped around it stays on your side:

```go
err := fmt.Errorf("saving task %s: %w", id, database.ErrNotFound)
core.Error(c, err) // the client gets 404 "not found"
```

The whole error, wrapping included, goes to the request's log line. An unexpected error, like a lost connection, reaches the client only as `internal server error`, while the log keeps the real cause:

```
level=ERROR msg=request method=POST path=/task status=500 … error="Error #01: saving task: connection refused\n"
```

## Next steps

- [Configuration](./configuration): settings and profiles
- [Errors & HTTP status](/reference/errors): the reference table
