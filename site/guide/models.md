# Models & schema

A model is a Go struct that describes a table. You write the struct once, and Preá uses it to create the table, to keep the database in step with your code and to read and write rows.

::: info What you'll learn
- How a struct maps to a table
- The tags that describe columns, keys and indexes
- How Go types become column types
- How Preá checks the database against your models
:::

## A model

<<< @/examples/todo/preas/task/model.go#model

That struct becomes this table:

```sql
create table "task" (
	"id" uuid not null default gen_random_uuid(),
	"title" varchar(200) not null,
	"done" bool not null default false,
	"list_id" uuid references "list" on delete cascade,
	primary key ("id")
);
create index "task_list_id_idx" on "task" ("list_id");
```

- **The table name** comes from the `Table()` method.
- **Column names** come from the field names, in snake_case: `ListID` becomes `list_id`.
- **`null` or `not null`** comes from the Go type. A pointer (`*string`) may be null; a plain value (`string`) may not.

A struct is a model as soon as a [repository](./repositories) uses it. There is nothing else to register.

## Tags

Each tag describes one thing about a column. A field can have several.

| Tag | What it does |
|---|---|
| `pk:""` | the primary key. On several fields, a composite key. Without any `pk`, the `id` column is the key |
| `generated:"uuid"` | the database generates a UUID (`gen_random_uuid()`); on a string key |
| `generated:"auto_increment"` | the database generates a number; on an integer key |
| `generated:"auto"` | `uuid` for a string key, `auto_increment` for an integer key |
| `size:"80"` | a string with at most 80 characters (`varchar(80)`); without it, `text` |
| `type:"jsonb"` | an explicit PostgreSQL type, for anything the Go type doesn't say |
| `default:"false"` | a default value, checked against the field's type: `default:"1"`, `default:"pt-BR"` |
| `unique:""` | no two rows with the same value |
| `unique:"name"` | the fields with the same name are unique together |
| `index:""` / `index:"name"` | an index, alone or shared by the fields with the same name |
| `fk:"list"` | references the primary key of the `list` table |
| `ondelete:"cascade"` | with `fk`: `cascade`, `restrict` or `set null` |
| `version:""` | a version counter for [safe concurrent updates](./repositories#save-the-whole-row) |
| `db:"name"` | the column's name, when it isn't the snake_cased field name |
| `db:"-"` | not a column |

The [struct tags reference](/reference/struct-tags) has every tag in one place.

::: tip Foreign keys take the type of their key
A foreign key column gets the type of the key it points to. `list.id` is a `uuid`, so `ListID *string` with `fk:"list"` becomes a `uuid` column too, with no `type` tag. A `type` that differs from the key's stops the app at startup.
:::

## Types

| Go | PostgreSQL |
|---|---|
| `string` | `text`, or `varchar(n)` with `size` |
| `int64`, `int` | `int8` |
| `int32` | `int4` |
| `int16`, `int8` | `int2` |
| `bool` | `bool` |
| `float64` | `float8` |
| `float32` | `float4` |
| `time.Time` | `timestamptz` |
| `[]byte` | `bytea` |

Any other type needs `type:"..."`, like `type:"jsonb"` or `type:"numeric"`.

## Checked at startup

Preá reads every model when the app starts. A tag with a wrong value stops the app, and every problem is listed at once, with the field that has it:

```
database: task.Task.Title: size:"x" needs a string field and a positive number
```

The checks cover tag values, `generated` outside the key, `size` on a non-string, a `default` that doesn't fit the type, a `version` that isn't an integer, a missing key, an `fk` to a table no model describes, and foreign keys in a cycle.

::: warning A misspelled tag name is not caught
`uniqe:""` looks like a tag from another library to Go, so Preá can't tell it is a typo. Compare the generated migration with what you meant.
:::

## The schema setting

`database.schema` decides what Preá does with the tables at startup:

| Value | What happens |
|---|---|
| `none` (default) | only the tags are checked |
| `validate` | the tables are compared with the models. A missing table or column, a different type or different nullability stops the app, all listed at once |
| `update` | missing schemas, tables, columns and indexes are created. Nothing is ever dropped or changed |
| `create` | the models' tables are dropped and created again, empty |
| `create-drop` | like `create`, and the tables are dropped again when the app stops. Made for tests |

A new project uses `validate` in the base file and `create-drop` in the `tests` profile. The tables themselves come from [migrations](./migrations), which keep a history you can review.

## Next steps

- [Migrations](./migrations): turn model changes into reviewed SQL
- [Repositories](./repositories): read and write rows
