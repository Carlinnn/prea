# Services

The service holds the rules of a feature: what is allowed, what must be true and what happens. Handlers call it, and it calls the repository and other preás' services.

::: info What you'll learn
- What belongs in a service
- How a service rejects a request with the right HTTP status
- How a service uses another preá
:::

## A service

<<< @/examples/todo/preas/task/service.go#service

A service is a component like any other. It asks for its repository and, here, for the service of the `list` preá, which `task` imports.

## Rules

A rule is a decision the service makes before or after it touches the data:

<<< @/examples/todo/preas/task/service.go#create

The service cleans the title, checks it, makes sure the list exists, and only then creates the task. The handler doesn't know about any of this; it only hands over the request.

## Say no with a status

When a rule fails, return an error made with `core.NewError`, with the HTTP status that fits:

<<< @/examples/todo/preas/task/service.go#rules

The handler passes the error to `core.Error`, and the client gets that status and message:

```json
{
    "text": "title can't be blank",
    "code": 2
}
```

Use the `net/http` constants for the status. `http.StatusBadRequest` (400) means the request is wrong; `http.StatusUnprocessableEntity` (422) means it is well-formed but refers to something that doesn't work, like a list that doesn't exist; `http.StatusConflict` (409) means it clashes with the current state.

## Errors from the database

Many database errors already carry a status. `database.ErrNotFound` replies 404, a duplicate unique value replies 409, and a malformed value replies 400. Return them as they are, unless the client deserves a clearer message, like `checkList` above, which turns a missing list into "list not found".

Compare them with `errors.Is`:

```go
if errors.Is(err, database.ErrNotFound) {
	// ...
}
```

## Pass the context along

Every service method takes a `context.Context` first and passes it to every call it makes. The request's deadline and cancellation travel with it, and so does the [transaction](./transactions), when there is one.

## Next steps

- [Errors & responses](./errors-and-responses): the full list of statuses
- [Repositories](./repositories): what the service can ask of the database
