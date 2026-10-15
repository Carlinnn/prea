# Writing the Preá docs

The site is built with [VitePress](https://vitepress.dev). Run it with:

```bash
cd site
npm ci
npm run docs:dev      # http://localhost:5173/prea/
npm run docs:build    # fails on a dead internal link
```

## Voice

- **English, second person, present tense:** "You declare a component", "Preá creates it".
- **Short sentences.** Explain one idea at a time, and keep each page to one concept.
- **Explain Preá on its own terms.** Don't compare it with other frameworks. Name the problem a feature solves and show it with an example.
- **Use plain words before precise ones.** Explain a term the first time it shows up and link to its page after that.
- **Show what really happens.** Use real commands, real output and the real error messages, copied from a run and not typed from memory.

## Page shape

1. A one-paragraph intro, then `::: info What you'll learn` with 2–4 bullets.
2. The smallest example that works.
3. The details, one `##` per idea.
4. A `## Next steps` section at the end, with 1–3 links.

Use the built-in features. Don't add plugins.

- **`::: tip`:** a good habit.
- **`::: warning`:** something that breaks.
- **`::: details Why?`:** the reasoning, for readers who want it.
- **`::: code-group`:** the same thing in two forms, like yaml and env.
- **`// [!code highlight]` / `{3-5}`:** the lines that matter in a block.

## Terms

| Term | Meaning |
|---|---|
| **preá** | a module of the app: one folder in `preas/`, one feature |
| **component** | a struct that a preá declares, which Preá creates and injects |
| **provider** | an entry in `Providers`: the declaration of a component |
| **global** | a ready value shared by every preá, like the database pool |
| **profile** | a named set of configuration: `dev`, `prod`, `tests` |

Write "Preá" for the framework and "preá" (lowercase) for a module of the app.

## Code that stays true

Longer Go snippets live in `site/examples/`, a Go module in the workspace. The CI vets and tests it, so a change in the API breaks the build, not the docs. Mark a region in the Go file:

```go
// #region service
type Service struct { ... }
// #endregion service
```

Then import it into the page:

```md
<<< @/examples/todo/preas/task/service.go#service
```

Short snippets of two or three lines can stay inline.
