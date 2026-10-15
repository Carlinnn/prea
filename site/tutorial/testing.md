# 8. Testing

Up to now you tested by hand, with `curl`. In this chapter you write tests that do the same thing in a second, every time you change something. They test through HTTP, the way clients use the API, against the real test database.

## Test the task preá

A preá's tests live next to its code. Create `preas/task/task_test.go`, starting with a small helper that sends a request and decodes the reply:

<<< @/examples/todo/preas/task/task_test.go#helper

Then the test:

<<< @/examples/todo/preas/task/task_test.go#test

`prea.Test` builds the app the way `prea dev` does, but with the `tests` profile, and only with the preás you pass: here `task`, and `list`, which it imports. It returns the app as an `http.Handler`, so the requests never touch the network. When the test ends, it closes everything.

Remember `create-drop` in the tests profile: the tables are created when the test builds the app, and dropped at the end. Every run starts empty.

## Test the whole app

Tests in `tests/` build the app with everything `main.go` has, including the API key middleware. The project already has `tests/app_test.go`, with a helper that builds the whole app. Add a test for the key:

<<< @/examples/todo/tests/app_test.go#apikey

The key comes from `config/application-tests.yaml`, the one you added in the last chapter.

## Run them

```bash
prea tests
```

```
ok  	todo/preas/health	0.330s
ok  	todo/preas/task	0.389s
ok  	todo/tests	0.411s
```

`prea tests` runs the packages one after the other, because they all use the same test database.

Go's test flags work as usual: `prea tests -run TestTask -v` runs one test and shows each step.

Try breaking a rule on purpose, like changing `maxTitle` to `2`, and run the tests again. They catch it.

## What you learned

- `prea.Test` builds the app for a test, with the `tests` profile, and returns an `http.Handler`.
- A preá's tests live next to it; tests of the whole app live in `tests/`.
- `create-drop` gives every run empty tables.

[Next: ship it →](./ship-it)
