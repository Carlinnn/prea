# Repositories

A repository is the component that reads and writes a model's table. Preá gives every repository the common operations, and you add your own queries in plain SQL.

::: info What you'll learn
- The operations every repository has
- The difference between `Update` and `Save`
- How to write your own queries with named parameters
- How database errors reach the client
:::

## A repository

```go
type Repository struct {
	database.BaseRepository[Task]
}
```

Embedding `database.BaseRepository[Task]` gives the repository its database connection and every operation below, for the `Task` model. Declare it in the preá's `Providers`, and ask for it in the service:

```go
type Service struct {
	Repo *Repository `inject:""`
}
```

## Ready-made operations

| Method | What it does |
|---|---|
| `GetAll(ctx)` | every row, ordered by the primary key |
| `GetByID(ctx, id)` | one row, or `database.ErrNotFound` |
| `Where(ctx, database.Args{"list_id": id})` | the rows where every column equals its value (`nil` means `is null`) |
| `First(ctx, args)` | the first row that matches, or `ErrNotFound` |
| `Count(ctx, args)` | how many rows match (`nil` counts them all) |
| `Exists(ctx, args)` | whether any row matches |
| `Create(ctx, dto)` | inserts a row from a request struct and returns it |
| `Update(ctx, id, dto)` | changes the fields you sent and returns the row |
| `Save(ctx, entity)` | writes the whole row, inserting or updating |
| `Delete(ctx, id)` | deletes the row, or returns `ErrNotFound` |

The fields of the struct you pass to `Create` and `Update` match the columns by name, like the model's.

## Create

`Create` inserts every field of the struct you pass, usually a request struct from `dtos/`:

<<< @/examples/todo/preas/task/dtos/create_request.go#create

```go
task, err := s.Repo.Create(ctx, req)
```

Columns the struct doesn't have, like `id` and `done`, get their database defaults. So do pointer fields left `nil`, which is how an optional field of a request falls back to the default:

```go
type CreateRequest struct {
	Name string  `json:"name"`
	Role *string `json:"role"` // not sent: the column default ('member')
}
```

A plain `string` field is always written, even empty: `""` replaces the default. The returned row has every column filled in.

## Update: change what was sent

`Update` changes only the fields that are **not zero**. A request with only `title` leaves `done` alone. That is what a `PATCH` needs.

The catch is that a zero value can't be sent: `false`, `0` and `""` are skipped. To allow them, use a pointer. `nil` means "keep it", and a pointer to `false` means "set it to false":

<<< @/examples/todo/preas/task/dtos/update_request.go#update

```go
task, err := s.Repo.Update(ctx, id, req) // {"done": false} now writes false
```

You can also pass `database.Args{"done": false}` for a quick update from code. An update with nothing to change returns `database.ErrInvalidInput` (400).

## Save the whole row

`Save` writes **every** column of the model, zeros included. Use it to load a row, change it in code and write it back:

```go
err := s.Transaction(ctx, func(ctx context.Context) error {
	task, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	task.Done = !task.Done // false is written too
	_, err = s.Repo.Save(ctx, task)
	return err
})
```

What `Save` does depends on the key:

- **Empty generated key:** it inserts, and the database generates the key.
- **Filled generated key:** it updates, or returns `ErrNotFound` if the row is gone.
- **Natural key** (not generated, composite included): it inserts or updates.

**Safe concurrent updates.** If two requests load the same row and both save it, the second one silently overwrites the first. A `version:""` field prevents that. `Save` only writes if the row still has the version it loaded, and then increments it. Otherwise it returns `database.ErrConflict` (409), and the client can reload and try again.

```go
type Task struct {
	ID      string `pk:"" generated:"uuid"`
	Title   string
	Version int64  `version:""`
}
```

## Your own queries

`r.Prepare` runs any SQL. Parameters are named with `@`, and the rows map to the model:

```go
func (r *Repository) Search(ctx context.Context, text string) ([]Task, error) {
	return r.Prepare("select * from task where title ilike @q order by title").
		Bind("q", "%"+text+"%").
		All(ctx)
}
```

| Method | Returns |
|---|---|
| `All(ctx)` | every row |
| `One(ctx)` | exactly one row, or `ErrNotFound` |
| `Exec(ctx)` | runs a statement without rows and returns how many rows it affected |

Bind several parameters at once with `BindArgs(database.Args{...})`.

**Other shapes.** To read only some columns, or a value that isn't a model, use `database.Prepare` with the type you want:

```go
type Summary struct {
	ID    string
	Title string
}

summaries, err := database.Prepare[Summary](r.DB, "select id, title from task").All(ctx)
pending, err := database.Prepare[int64](r.DB, "select count(*) from task where not done").One(ctx)
```

The ready-made operations select every column. When you need fewer, write the query.

## Errors

Database errors become errors with an HTTP status, so a service can usually return them as they are:

| What happened | Error | Status |
|---|---|---|
| no row found | `database.ErrNotFound` | 404 |
| a unique value already exists | `database.ErrConflict` | 409 |
| a malformed value (like an invalid UUID), a value too long for its column or out of its type's range, a null in a `not null` column, a broken foreign key or a failed check | `database.ErrInvalidInput` | 400 |

The client sees only the short message, like `"not found"`. The database's details go to the log.

## Next steps

- [Transactions](./transactions): several writes that succeed or fail together
- [Models & schema](./models)
