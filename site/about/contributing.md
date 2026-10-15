# Contributing

Thanks for wanting to help. Preá is young, and every bug report, idea and fix counts.

## Report a bug or suggest an idea

Open an [issue](https://github.com/prea-framework/prea/issues). For a bug, include:

- what you did, ideally as the smallest project or code that shows it;
- what you expected, and what happened, with the full error message;
- your Preá version (`go list -m github.com/prea-framework/prea`) and Go version (`go version`).

For a larger change, open an issue first to talk it through before you write code.

## Set up the repository

```bash
git clone https://github.com/prea-framework/prea
cd prea
```

The repository is a Go workspace (`go.work`) with one module per folder: `core`, `di`, `database`, `config`, `security`, `logger`, `cmd/prea` and the root. The workspace links them, so a change in one is seen by the others without publishing.

## Run the tests

From the root, use the full path. `./...` only covers the root module:

```bash
go vet github.com/prea-framework/prea/...
go test github.com/prea-framework/prea/...
```

The database tests need PostgreSQL. Without the variable below, they are skipped:

```bash
export PREA_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/prea_test?sslmode=disable'
go test github.com/prea-framework/prea/...
```

`-short` skips the slow tests, the ones that create and build whole projects.

## The docs

The site lives in `site/` and is built with [VitePress](https://vitepress.dev):

```bash
cd site
npm ci
npm run docs:dev
```

`site/STYLE.md` describes how the docs are written. The code the tutorial shows lives in `site/examples/todo`, a real project that the CI builds and tests.

## Send a change

1. Fork the repository and create a branch.
2. Make the change, with tests that fail without it.
3. Run `gofmt`, `go vet` and the tests.
4. Open a pull request that explains what changed and why.

**Commits** are in English, in the form `type(scope): what it does`, split by topic: `feat(cli): add prea dev`, `fix(database): ...`, `docs: ...`.

## License

By contributing, you agree that your contributions are licensed under the [MIT License](https://github.com/prea-framework/prea/blob/main/LICENSE).
