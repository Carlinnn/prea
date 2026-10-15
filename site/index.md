---
layout: home

hero:
  name: preá.go
  text: A modular web framework for Go
  tagline: Split your app into small preás. Declare what each one needs, and Preá wires it, checks it and serves it.
  image:
    light: /logo-light.svg
    dark: /logo.svg
    alt: A preá between curly braces
  actions:
    - theme: brand
      text: Get started
      link: /guide/what-is-prea
    - theme: alt
      text: Build a todo API
      link: /tutorial/
    - theme: alt
      text: GitHub
      link: https://github.com/prea-framework/prea

features:
  - icon: 🧩
    title: One feature, one preá
    details: A preá keeps a feature together, from its routes to its rules and its table. Your app is the list of its preás.
  - icon: 🔌
    title: Ask, and it is wired
    details: Tag a field with inject:"" and Preá fills it at startup. A missing dependency stops the app before it serves a single request.
  - icon: 🗄️
    title: Your models are the schema
    details: Struct tags describe the tables. Preá writes the migrations, applies them and gives every model a ready-made repository.
  - icon: ⚙️
    title: One file per environment
    details: A base configuration plus one file for dev, prod and tests. Secrets come from the environment, never from the repository.
  - icon: 🛡️
    title: Safe defaults
    details: CORS, CSRF protection, security headers, timeouts and a graceful shutdown come configured in every new project.
  - icon: 🚀
    title: A CLI for the whole cycle
    details: prea new, prea dev with hot reload, generators, migrations, tests and builds, all from the same command.
---

<div class="vp-doc" style="max-width: 760px; margin: 64px auto 0; padding: 0 24px">

## From zero to a running API

```bash
curl -fsSL https://raw.githubusercontent.com/prea-framework/prea/main/install.sh | sh

prea new todo --with-db
cd todo
prea dev
```

A preá is a few plain Go files. You declare its components, and every component asks for what it needs:

```go
var Module = &core.Module{
	Name:      "task",
	Providers: []any{Repository{}, Service{}, Handler{}},
}

type Service struct {
	Repo *Repository `inject:""` // Preá fills this when the app starts
}
```

[Read the quick start →](/guide/quick-start)

</div>
