# 4. Business rules

A title can't be blank or longer than 200 characters, and completing a task must not erase its title. These are **rules**, and rules live in the service. In this chapter you write them and learn how a rule tells the client what went wrong.

## One rule, one function

Both creating and renaming a task need the same title checks. Add them to the end of `preas/task/service.go`:

```go
const maxTitle = 200

func checkTitle(title string) error {
	switch {
	case title == "":
		return core.NewError(http.StatusBadRequest, "title can't be blank")
	case utf8.RuneCountInString(title) > maxTitle:
		return core.NewError(http.StatusBadRequest, "title is too long: 200 characters at most")
	}
	return nil
}
```

`core.NewError` creates an error that carries an HTTP status. The handler doesn't need to know which status fits; the error already says it.

`utf8.RuneCountInString` counts characters, not bytes, so "café" counts as 4. Add `"unicode/utf8"` to the imports.

## Use it when creating

```go
func (s *Service) Create(ctx context.Context, req dtos.CreateRequest) (Task, error) {
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" { // [!code --]
		return Task{}, core.NewError(http.StatusBadRequest, "title can't be blank") // [!code --]
	if err := checkTitle(req.Title); err != nil { // [!code ++]
		return Task{}, err // [!code ++]
	}
	return s.Repo.Create(ctx, req)
}
```

## Updates change only what was sent

A client completing a task sends just `{"done": true}`. The update must change `done` and leave the title alone. This is why the fields of `UpdateRequest` are pointers:

<<< @/examples/todo/preas/task/dtos/update_request.go#update

A field the client didn't send stays `nil`, and the repository's `Update` skips it. A field it sent, even `false`, is a pointer to that value, and gets written. With plain `bool` and `string`, there would be no way to tell "set it to false" from "not sent".

Check the title only when it was sent:

```go
func (s *Service) Update(ctx context.Context, id string, req dtos.UpdateRequest) (Task, error) {
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if err := checkTitle(title); err != nil {
			return Task{}, err
		}
		req.Title = &title
	}
	return s.Repo.Update(ctx, id, req)
}
```

## Try it

Complete the task you created, using its `id`:

```bash
curl -X PATCH localhost:8000/task/4caff630-444a-4a86-bb31-3b5e963cfb32 \
  -H 'Content-Type: application/json' -d '{"done": true}'
```

```json
{
    "code": 1,
    "data": {
        "id": "4caff630-444a-4a86-bb31-3b5e963cfb32",
        "title": "Buy coffee",
        "done": true
    }
}
```

The title survived. Now break a rule:

```bash
curl -X POST localhost:8000/task -H 'Content-Type: application/json' -d '{"title": "   "}'
```

```json
{
    "text": "title can't be blank",
    "code": 2
}
```

The status is 400, and `code: 2` marks a warning: the client asked for something it can fix.

Some errors come from the database itself. An `id` that isn't a valid UUID replies 400 `invalid input`, and an `id` that doesn't exist replies 404 `not found`. [Errors & responses](/guide/errors-and-responses) lists them all.

## What you learned

- Rules belong in the service, written once and shared.
- `core.NewError` gives an error its HTTP status, and the handler just passes it on.
- Pointer fields let an update tell "not sent" from a zero value.

[Next: a second preá →](./second-prea)
