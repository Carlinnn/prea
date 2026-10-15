# Testing

A Preá app is tested through HTTP, the way its clients use it. You build the app in the test, send it requests and check the replies. No server, no ports.

::: info What you'll learn
- Where tests live
- How `prea.Test` builds the app for a test
- How to test one preá, and the whole app
- How the tests get their own database
:::

## Run the tests

```bash
prea tests
```

`prea tests` runs `go test ./...` with the `tests` profile, one package at a time. Go's test flags pass through: `prea tests -run TestTask -v`, `prea tests ./preas/task`.

## Where tests live

- **A preá's tests** sit next to its code, like `preas/task/task_test.go`, in the same package. They can see the preá's unexported parts and build only that preá.
- **Tests of the whole app** live in `tests/`. They build every preá, middleware and protection, exactly like `main.go`.

## Test a preá

`prea.Test` builds the app like `prea.Run` does, with the `tests` profile, and returns it as an `http.Handler`:

<<< @/examples/todo/preas/task/task_test.go#test

It only builds the preás you pass, here `task` and the `list` it imports. The small `send` helper makes the requests:

<<< @/examples/todo/preas/task/task_test.go#helper

`prea.Test` works from any package. It finds the project root, so `config/` and `migrations/` are always found, and it closes the app and the database when the test ends.

## Test the whole app

Tests in `tests/` build the app with everything `main.go` uses:

```go
func app(t *testing.T) http.Handler {
	return prea.Test(t, prea.Config{
		Modules:     preas.Modules,
		Middlewares: preas.Middlewares,
		Security:    preas.Security,
		Logging:     preas.Logging,
	})
}
```

That makes them the right place to check what crosses preás, like a global middleware:

<<< @/examples/todo/tests/app_test.go#apikey

## The test database

The `tests` profile points to its own database and recreates the tables on every run:

```yaml
# config/application-tests.yaml
database:
  url: ${TEST_DATABASE_URL}
  schema: create-drop
```

`create-drop` creates the models' tables when the test builds the app and drops them when it ends. Every test starts with empty tables, and your development data is never touched.

::: info One package at a time
Every test package uses the same test database, so `prea tests` runs the packages one after the other: two packages creating the same tables at once would collide. Running `go test` yourself, pass `-p 1` for the same effect.
:::

## Replace a component

To test a service without its real dependencies, fill the field in the declaration. Preá keeps fields you fill. This works best when the field is an [interface](./dependency-injection#interfaces):

```go
type Service struct {
	Repo  *Repository `inject:""`
	Lists ListChecker `inject:""` // an interface: Exists(ctx, id) (bool, error)
}
```

```go
type fakeLists struct{}

func (fakeLists) Exists(ctx context.Context, id string) (bool, error) { return true, nil }

func TestWithFakeLists(t *testing.T) {
	module := &core.Module{
		Name:      "task",
		Providers: []any{Repository{}, Service{Lists: fakeLists{}}, Handler{}},
	}
	app := prea.Test(t, prea.Config{Modules: []*core.Module{module}})
	// ...
}
```

The fake replaces the real list service only in this test. Depending on interfaces between preás makes this possible.

## Next steps

- [Deploying](./deploying)
- [Build a todo API](/tutorial/), with tests in chapter 8
