# Errors & HTTP status

`core.Error(c, err)` replies with the status of the first error in `err`'s chain that has an `HTTPStatus() int` method, and with that error's message.

| Error | Status | Message |
|---|---|---|
| `core.NewError(status, message)` | `status` | `message`, or the standard status text when empty |
| `database.ErrInvalidInput` | 400 | `invalid input` |
| `security.ErrInvalidToken` | 401 | `invalid token` |
| `database.ErrNotFound` | 404 | `not found` |
| `database.ErrConflict` | 409 | `conflict` |
| a type with `HTTPStatus() int` | what it returns | its `Error()` |
| anything else | 500 | `internal server error` |

## Database errors

| What happened in PostgreSQL | Becomes |
|---|---|
| a unique value already exists | `ErrConflict` |
| `Save` with a stale `version` | `ErrConflict` |
| no row | `ErrNotFound` |
| any data error: a malformed value (an invalid UUID...), a value too long for its column, a number out of range, an invalid date | `ErrInvalidInput` |
| a null in a `not null` column, a broken foreign key, a failed check | `ErrInvalidInput` |
| an update or insert with nothing to write | `ErrInvalidInput` |

## Reply codes

| Constant | `code` | Used by default for |
|---|---|---|
| `core.SuccessCode` | 1 | status below 400 |
| `core.WarningCode` | 2 | status 400–499 |
| `core.DangerCode` | 3 | status 500 and above |

The full error, with everything wrapped around it, is written to the request's log line, never to the client.
