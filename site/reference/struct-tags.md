# Struct tags

Every struct tag Preá reads, in one place.

## Components

| Tag | On | What it does |
|---|---|---|
| `inject:""` | a field of a component | fills it with the component of that type: from the preá, its imports' exports, or the globals. The field must be exported, and a pointer or an interface |
| `inject:"name"` | a field of a component | fills it with the component named `name` (declared with `di.Named`) |
| `value:"key"` | a field of a component | fills it with the configuration value at `key`, converted to the field's type. A missing key leaves the field as it is |

Fields already filled in the declaration are kept. [Components & injection](/guide/dependency-injection)

## Models

| Tag | What it does |
|---|---|
| `db:"name"` | the column's name. Default: the field name in snake_case (`ListID` → `list_id`) |
| `db:"-"` | the field is not a column |
| `pk:""` | part of the primary key. Without any `pk`, the `id` column is the key. Can't be a pointer |
| `generated:"uuid"` | the database generates a UUID; on a `string` key |
| `generated:"auto_increment"` | the database generates a number; on an integer key |
| `generated:"auto"` | `uuid` or `auto_increment`, from the field's type |
| `size:"n"` | `varchar(n)`; on a `string` |
| `type:"name"` | an explicit PostgreSQL type: `jsonb`, `numeric`, `uuid`, `date`... |
| `default:"literal"` | a default value, checked against the field's type. Database functions like `now()` are not supported yet |
| `unique:""` | a unique column |
| `unique:"group"` | unique together with the fields of the same group |
| `index:""` / `index:"group"` | an index, alone or shared by the group |
| `fk:"table"` | a foreign key to the primary key of the model with that table. The column takes that key's type |
| `ondelete:"action"` | with `fk`: `cascade`, `restrict` or `set null` |
| `version:""` | an integer incremented on each `Save`, which refuses stale writes with a 409 |

Nullability comes from the Go type: a pointer may be null, a value may not. [Models & schema](/guide/models)

## Request bodies

| Tag | What it does |
|---|---|
| `json:"name"` | the JSON field name |
| `binding:"required"` | binding fails, with a 400, if the field is missing |

Request structs map to columns like models do, by `db` tag or by snake_cased field name.

## Configuration blocks

| Tag | What it does |
|---|---|
| `yaml:"name"` | the key of a field inside a block read with `value:"block"` |
