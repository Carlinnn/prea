# Preá Framework

Framework web modular para Go, inspirado no Spring e no Angular. Ele traz módulos com injeção de dependência por tag, handlers HTTP sobre o [gin](https://github.com/gin-gonic/gin), um repositório base para Postgres com queries no estilo PDO, middlewares em três níveis e uma CLI que gera o código de cada módulo.

> **Status:** pré-alfa (`v0.0.1`). A API ainda muda. Os gaps conhecidos e os próximos passos estão no [ROADMAP.md](ROADMAP.md).

## Sumário

- [Requisitos](#requisitos)
- [Começando](#começando)
- [Estrutura do projeto](#estrutura-do-projeto)
  - [Módulos Go](#módulos-go)
- [Conceitos](#conceitos)
  - [Módulos (preás)](#módulos-preás)
  - [Injeção de dependência](#injeção-de-dependência)
  - [Handler, Service, Repository, Model e DTOs](#handler-service-repository-model-e-dtos)
  - [Banco de dados](#banco-de-dados)
  - [Middlewares](#middlewares)
  - [Logs](#logs)
  - [Erros e respostas](#erros-e-respostas)
  - [Segurança HTTP](#segurança-http)
  - [Autenticação](#autenticação)
- [CLI](#cli)
- [Configuração](#configuração)
- [Testes](#testes)

## Requisitos

- Go **1.26+**
- PostgreSQL, se a app usar banco (com `gen_random_uuid()`, nativo a partir do 13)

## Começando

```bash
# 1. CLI: instala com o go install e põe no PATH (zsh, bash, fish ou sh)
curl -fsSL https://raw.githubusercontent.com/prea-framework/prea/main/install.sh | sh
# uma versão: ... | PREA_VERSION=v0.0.1 sh; num clone: go install ./cmd/prea

# 2. Um projeto novo
prea new loja --with-db     # sem --with-db: só o preá health, sem banco
cd loja
cp .env.example .env        # preencha DATABASE_URL e TEST_DATABASE_URL

# 3. Rodar com hot reload: o profile dev gera e aplica as migrations a partir dos models
prea dev
```

Para testar:

```bash
curl localhost:8000/health
curl -X POST localhost:8000/item -H 'Content-Type: application/json' -d '{"name": "caneta"}'

prea tests                  # os testes, com o profile tests
prea build                  # o binário em bin/loja; leve config/ e migrations/ junto no deploy
```

## Estrutura do projeto

Um projeto criado pelo `prea new`:

```
loja/
├── main.go                  # prea.Run(prea.Config{Modules: preas.Modules, ...})
├── config/                  # application.yaml e os profiles dev, prod e tests
├── migrations/              # 0.0.1_models.sql...
├── middleware/              # middlewares da app (prea g m)
├── preas/
│   ├── preas.go             # Modules, Middlewares, Security e Logging: o que a app declara
│   ├── health/              # GET /health, com o teste ao lado (health_test.go)
│   └── item/                # com --with-db: o CRUD gerado pelo prea g p
└── tests/                   # os testes de ponta a ponta, da app inteira
```

```go
// main.go: só liga as variáveis do preas/preas.go ao prea.Run
func main() {
	prea.Run(prea.Config{
		Modules:     preas.Modules,
		Middlewares: preas.Middlewares,
		Security:    preas.Security,
		Logging:     preas.Logging,
	})
}
```

O `prea.Run`, em ordem:
1. carrega `config/` com o profile ativo;
2. confere as chaves obrigatórias e configura os logs;
3. abre o banco de `database.url`, se houver;
4. com `PREA_MIGRATE` (o `prea migrate`), roda o comando e sai;
5. aplica o segredo JWT, se houver;
6. monta os módulos, gera e aplica as migrations, confere o schema e serve até SIGINT/SIGTERM.

Este repositório:

```
prea/
├── prea.go                  # módulo da raiz: prea.Run e prea.Test, que ligam os outros
├── go.work                  # liga os módulos no desenvolvimento
├── cmd/prea/                # módulo da CLI (instalável; não depende do framework)
│   └── templates/           # o projeto do prea new (new/) e os preás do prea g p (prea/)
├── core/                    # módulo: Module, App, Router, Middleware, Security, respostas e `core.NewError`
├── di/                      # módulo: container de injeção
├── database/                # módulo: BaseRepository, Prepare, schema e migrations
├── security/                # módulo: JWT e hash de senha (argon2id)
├── logger/                  # módulo: logs por nível (slog)
└── config/                  # módulo: carrega a config (koanf) e os profiles
```

Direção das dependências: **CLI → nada**, **app → framework**, **framework → nada da app**.

### Módulos Go

Cada pasta do framework é um módulo, instalado e versionado à parte. O da raiz (`prea.Run`) liga todos e é o que um projeto do `prea new` usa; quem monta o `core.App` à mão e usa só o `core` não leva o driver do banco nem o JWT:

| Módulo | Depende de (framework) | Dependências externas |
|---|---|---|
| `github.com/prea-framework/prea` | `core`, `database`, `security`, `config`, `logger`, `di` | gin, pgx |
| `.../di` | — | — |
| `.../logger` | — | — |
| `.../config` | — | koanf, mapstructure, godotenv |
| `.../security` | — | golang-jwt, x/crypto |
| `.../database` | `logger` (e `di` nos testes) | pgx, goose |
| `.../core` | `di`, `config`, `logger` | gin |
| `.../cmd/prea` | — | urfave/cli |

```bash
go get github.com/prea-framework/prea                  # tudo, com o prea.Run
go get github.com/prea-framework/prea/core             # só HTTP e DI
```

- **O `core` não conhece o banco nem a segurança.** Os erros do `database` e do `security` dizem o próprio status (`HTTPStatus()`, veja [Erros e respostas](#erros-e-respostas)), e o banco entra na subida pelo gancho `core.App.Setup`:
  ```go
  core.App{..., Globals: []any{db}, Setup: database.Setup(db, opts)}
  ```
- **Versões:** uma tag por módulo, com o prefixo da pasta (`core/v0.0.1`, `cmd/prea/v0.0.1`); a do módulo da raiz não tem prefixo (`v0.0.2`). Todos os módulos ficam na mesma major; as minors e os patches andam separados, conforme cada um muda. A versão atual é a pré-alfa `v0.0.1`.
- **No repositório:** o `go.work` commitado liga os módulos pelo disco, então um clone compila sem `replace`. Um módulo que passa a exigir uma versão nova de um irmão só funciona fora do workspace depois que a tag dessa versão é publicada. Na raiz, o `./...` cobre só o módulo da raiz, porque os outros são módulos aninhados; os comandos usam o caminho: `go test github.com/prea-framework/prea/...`.
- **O CI** confere que cada módulo importa exatamente os módulos do framework que o seu `go.mod` exige. No workspace, um import sem `require` compila, e é essa conferência que impede o `core` de voltar a importar o `database`.

## Conceitos

### Módulos (preás)

Cada funcionalidade é um módulo em `preas/<nome>/`:

```
user/
├── module.go       # declara os componentes, exports, imports e middlewares
├── handler.go      # HTTP: rotas, leitura do corpo da requisição, respostas
├── service.go      # regras de negócio
├── repository.go   # acesso ao banco
├── model.go        # entidade (mapeia a tabela)
└── dtos/           # um arquivo por DTO
```

```go
var Module = &core.Module{
	Name: "user",
	Providers: []any{
		Repository{},
		di.Export(Service{}), // visível para módulos que importarem este
		Handler{},
	},
}
```

Um módulo que usa outro o declara em `Imports`:

```go
var Module = &core.Module{
	Name:      "auth",
	Imports:   []*core.Module{user.Module},
	Providers: []any{Service{}, Handler{}},
}
```

Os módulos são ligados à aplicação na `var Modules` do `preas/preas.go`. O `prea g p` registra o módulo novo ali automaticamente.

### Injeção de dependência

Campos com a tag `inject:""` são preenchidos na subida:

```go
type Service struct {
	Repo  *Repository   `inject:""`
	Users *user.Service `inject:""` // vem do export de um módulo importado
}
```

- **Ordem de busca:** os providers do próprio módulo, depois os exports dos módulos importados, depois os globais (`core.App.Globals`, por exemplo o pool do banco).
- **Interfaces:** um campo do tipo interface é injetado se existir exatamente um provider que a implementa.
- **Mais de um candidato** para o mesmo tipo dá erro na subida, e o erro pede um nome. Nunca é escolhido um em silêncio.
- **Structs embutidas** também são injetadas; é assim que o `BaseRepository` recebe a conexão.
- **Providers são declarados como valor** (`Repository{}`), e não como ponteiro. O módulo é só uma declaração: cada container (cada `core.App`) cria as **próprias instâncias** a partir dela e nunca altera a declaração.
  - Dois `App` no mesmo processo, como testes com bancos diferentes ou uma API e um admin, não compartilham nada.
  - Vários containers podem carregar o mesmo módulo ao mesmo tempo sem data race.
  - Declarar com `new(Repository)` dá erro na subida: `provider *user.Repository must be declared as a value, user.Repository{}`.
- **Campo preenchido na declaração** é mantido, o que serve para mocks e configuração: `Service{Repo: mockRepo}` e `routes{path: "/x"}`. A cópia é rasa, então ponteiros preenchidos ali são compartilhados por quem usar a mesma declaração.
- Os componentes são **singletons por container**: dentro de um `App`, todos que injetam `*Service` recebem a mesma instância. Os campos injetados precisam ser exportados.
- Instâncias prontas, como o pool do banco, entram em `core.App.Globals`, e essas sim são compartilhadas.
- **Concorrência:** um mesmo componente atende todas as requisições ao mesmo tempo, cada uma em sua goroutine. Não guarde estado de requisição em campos da struct; use variáveis locais ou o `gin.Context`. Estado compartilhado mutável precisa de `sync.Mutex` ou `atomic`.
- **Erros aparecem na subida**, com o caminho do campo, por exemplo `di: module auth: Service.Users: no provider for *user.Service`. Isso vale para dependência faltando, provider duplicado, interface ambígua e ciclo de imports.

Use sempre `inject:""`. A forma curta `inject` sem valor não é uma tag válida em Go (o `go vet` acusa) e dá erro na subida.

**Componentes nomeados.** O nome é opcional, como um token do Angular. Ele serve para quando há mais de um componente do mesmo tipo:

```go
// num core.App montado à mão (o prea.Run abre só o pool de database.url)
Globals: []any{mainDB, di.Named("analytics", analyticsDB)}

// em um módulo; Named e Export podem ser combinados em qualquer ordem
Providers: []any{di.Named("stripe", Gateway{}), di.Export(di.Named("paypal", Gateway{}))}

type Reports struct {
	DB        *pgxpool.Pool `inject:""`          // o sem nome (mainDB)
	Analytics *pgxpool.Pool `inject:"analytics"` // o nomeado
}
```

- `inject:"nome"` procura só os componentes com esse nome.
- `inject:""` prefere o componente sem nome. Se só existir um, mesmo nomeado, ele é usado.

**Ciclo de vida.** Um componente pode ter, opcionalmente, hooks de subida e de desligamento:

```go
type Mailer struct {
	Config *Config `inject:""`
	client *smtp.Client
}

// roda na subida, depois da injeção e antes de o servidor aceitar requisições
func (m *Mailer) Init(ctx context.Context) error {
	c, err := smtp.Dial(m.Config.SMTPAddr)
	if err != nil {
		return err // a app não sobe
	}
	m.client = c
	return nil
}

// roda no desligamento, depois que as requisições em andamento terminam
func (m *Mailer) Close(ctx context.Context) error { return m.client.Quit() }
```

- **`Init`** faz o papel de construtor: lê config, valida, abre conexões. Se der erro, a app não sobe, e os componentes que já tinham iniciado são fechados.
- **Ordem:** o `Init` segue as dependências. Se A injeta B, o `B.Init` roda antes do `A.Init`, inclusive entre módulos. O `Close` roda na ordem inversa.
- **`Close`** pode ser `Close(ctx) error` ou `Close() error` (`io.Closer`). Ele roda no desligamento gracioso, antes de o pool do banco fechar.
- **Globals não passam pelo ciclo de vida:** quem cria, como o `main` com o pool, é quem fecha.
- **Em testes:** `engine, closer, err := app.Build(ctx)` monta a app com o `Init` e devolve o `closer`, que roda o `Close`.

### Handler, Service, Repository, Model e DTOs

**Handler.** Só cuida de HTTP. Todo componente que implementa `core.Router` é montado automaticamente:

```go
type Handler struct {
	Service *Service         `inject:""`
	Auth    *middleware.Auth `inject:""`
}

var _ core.Router = (*Handler)(nil) // falha na compilação se Register estiver errado

func (h *Handler) Register(r gin.IRouter) {
	g := r.Group("/users")
	g.POST("", h.create)
	g.GET("", h.Auth.Handle, h.list)
}
```

**Service.** Guarda as regras de negócio e devolve erros com status (veja [Erros e respostas](#erros-e-respostas)).

**Model.** A struct da tabela, com o método `Table()`. As tags descrevem o schema (veja [Model e schema](#model-e-schema)):

```go
type User struct {
	ID        string     `json:"id" pk:"" generated:"uuid"`
	Email     string     `json:"email" unique:""`
	LastLogin *time.Time `json:"lastlogin" db:"lastlogin"`
}

func (User) Table() string { return "users" }
```

**DTOs.** Ficam em `dtos/`, um arquivo por DTO. Os campos viram colunas no `Create`/`Update` pela mesma regra de nomes do model.

### Banco de dados

#### Model e schema

O model é a fonte da verdade do schema. Cada opção é uma tag própria, como o `inject:""`:

```go
type Player struct {
	ID        string     `pk:"" generated:"uuid"`
	Email     string     `size:"255" unique:""`
	First     string     `unique:"full_name"`                   // unique composto: mesmo nome de grupo
	Last      string     `unique:"full_name"`
	TeamID    *int64     `fk:"teams" ondelete:"cascade" index:""`
	Profile   int16      `default:"1"`
	Settings  []byte     `type:"jsonb"`
	Version   int64      `version:""`
	LastLogin *time.Time `db:"lastlogin"`
	Notes     string     `db:"-"`
}
```

| Tag | O que faz |
|---|---|
| `db:"nome"` / `db:"-"` | nome da coluna (o padrão é o campo em snake_case: `TeamID` → `team_id`), ou ignora o campo |
| `pk:""` | chave primária; em mais de um campo, é composta. Sem nenhum `pk`, a coluna `id` é a PK |
| `generated:"auto_increment\|uuid\|auto"` | só na PK. `auto_increment`: inteiro gerado pelo banco (`generated by default as identity`). `uuid`: `uuid default gen_random_uuid()`. `auto`: escolhe pelo tipo Go (inteiro ou string) |
| `size:"n"` | `varchar(n)`; sem ele, uma string vira `text` |
| `type:"x"` | tipo explícito, no nome interno do Postgres (`jsonb`, `int4`, `numeric`) |
| `default:"lit"` | valor padrão **literal**, conferido contra o tipo do campo: `default:"1"`, `default:"true"`, `default:"pt-BR"`. Funções do banco (`now()`) não entram por enquanto |
| `unique:""` / `unique:"grupo"` | unique da coluna, ou composto com os campos do mesmo grupo |
| `index:""` / `index:"grupo"` | índice, com a mesma regra de grupo |
| `fk:"tabela"` + `ondelete:"cascade\|restrict\|set null"` | chave estrangeira para a PK do model daquela tabela |
| `version:""` | lock otimista, como o `@Version` do JPA (veja `Save`) |

- **Nulidade vem do tipo Go:** um ponteiro aceita `null`; um valor é `not null`.
- **Tipos:** `string` → `text`, `int64`/`int` → `int8`, `int32` → `int4`, `int16`/`int8` → `int2`, `bool` → `bool`, `float64` → `float8`, `float32` → `float4`, `time.Time` → `timestamptz`, `[]byte` → `bytea`. Qualquer outro tipo precisa de `type`.
- **Erros na subida:** tag com valor inválido, `generated` fora da PK, `size` fora de string, `default` que não converte, `version` que não é inteiro, model sem PK, `fk` para uma tabela sem model e ciclo de FKs. Todos de uma vez, com o caminho do campo: `database: user.User.Email: size:"x" needs a string field and a positive number`.
- **Limitação:** um typo no nome da tag (`uniqe:""`) não é detectado, porque o Go não diferencia uma tag desconhecida das de outras libs (`json`, `binding`).

**Schema na subida.** A chave `database.schema` funciona como o `ddl-auto` do Hibernate. Os models são os dos repositórios do container (todo componente que embute um `BaseRepository[T]`), sem registro à parte:

| Modo | O que faz |
|---|---|
| `none` (padrão) | só confere as tags |
| `validate` | compara as tabelas com os models. Tabela ou coluna faltando, tipo diferente e nulidade diferente impedem a subida, todas listadas de uma vez |
| `update` | cria o que falta: schemas, tabelas, colunas e índices. Nunca apaga nem altera uma coluna |
| `create` | apaga as tabelas dos models e as cria de novo |
| `create-drop` | `create`, e apaga as tabelas no desligamento; bom para testes |

O schema é aplicado numa transação, depois da injeção e antes dos `Init`, então um `Init` já encontra as tabelas. Para manter histórico e revisar as mudanças, use as [migrations](#migrations): num projeto do `prea new --with-db`, a base só valida (`schema: validate`) e o dev gera as migrations a partir dos models.

#### Migrations

Uma pasta de arquivos SQL versionados, rodados pelo [goose](https://github.com/pressly/goose). **O que está no arquivo roda**, como está escrito, inclusive funções com `$$`:

```
migrations/
├── 0.0.1_models.sql          # gerada a partir dos models
├── 0.0.2_add_active.sql      # prea migrate create add_active
└── 0.1.0_backfill.sql        # a versão é sua: pode pular
```

```sql
alter table users add column active bool not null default true;

-- prea:down
alter table users drop column active;
```

- **Nome:** `<versão>_<nome>.sql`. A versão é semver (`0.0.1`, o padrão) ou um timestamp (`20261013120000`, com `version: timestamp`), sem misturar os dois na mesma pasta. O `create` sobe o patch: `0.0.1`, `0.0.2`...
- **`-- prea:down`** (opcional): o que vem abaixo é o que o `down` roda. Sem essa linha, o `down` recusa a migration em vez de desfazer pela metade.
- **`-- prea:no-transaction`** (opcional, na primeira linha): roda fora de transação, para `create index concurrently`. Sem ela, cada migration roda numa transação: se falha, nada dela fica, e as seguintes não rodam.
- **Controle:** o que já rodou fica na tabela `prea_migrations`, criada antes da primeira migration. Um advisory lock impede duas instâncias de migrar ao mesmo tempo. Uma migration pendente mais antiga que a última aplicada (depois de um merge) também roda.

```yaml
database:
  schema: validate
  migrations:
    dir: migrations       # padrão
    version: semver       # semver (padrão) | timestamp
    on_start: false       # aplica as pendentes na subida
    from_models: false    # gera, na subida, uma migration com o diff entre os models e o banco
```

**Na subida**, a ordem é: `from_models` gera → `on_start` aplica → `schema` confere. Quem faz isso é o `database.Setup(pool, opts)`, ligado ao `core.App.Setup` no `main`: ele roda depois de os módulos carregarem e antes de qualquer `Init`. No profile dev da app de exemplo, os dois ficam ligados: mudou um model, reiniciou, e a migration é gerada e aplicada.

```sql
-- migrations/0.0.2_models.sql, gerada depois de adicionar Color *string ao model
-- Generated by prea from the models: review it before committing.

alter table "notes" add column "color" text;

-- prea:down
alter table "notes" drop column "color";
```

- **O diff** compara os models com o banco do profile: tabela, coluna e índice faltando, tipo e nulidade. Uma coluna que existe no banco e não no model sai **comentada**, porque apagá-la perde dados; quem revisa decide.
- **Com migrations pendentes**, nada é gerado: elas podem já ter as mudanças. Aplique-as antes (`on_start` ou `prea migrate up`).
- **`from_models` com `schema: update`/`create`** dá erro na subida: um aplicaria direto o que o outro registra.
- Em produção, deixe o `from_models` desligado e aplique as migrations no deploy (`prea migrate up`) ou com `on_start`.

#### Repositório

O `database.BaseRepository[T]` traz o CRUD pronto:

| Método | O que faz |
|---|---|
| `GetAll(ctx)` | todas as linhas, ordenadas pela PK |
| `GetByID(ctx, id)` | uma linha ou `ErrNotFound`. Com PK composta, o `id` é um `database.Args` com todas as colunas da chave (vale para `Update` e `Delete`) |
| `Where(ctx, database.Args{"owner_id": 1})` | as linhas que batem com o filtro (`nil` vira `IS NULL`) |
| `First(ctx, args)` | a primeira linha que bate com o filtro |
| `Create(ctx, dto)` | insere **todos** os campos do DTO e devolve a linha criada |
| `Update(ctx, id, dto)` | *patch*: atualiza só os campos **não zerados**; para gravar um zero, use ponteiro no DTO (`nil` = não mexe) ou `database.Args`. Incrementa o `version` |
| `Save(ctx, entidade)` | como o `save` do JPA: grava **todas** as colunas, zeros inclusive. Com PK zerada, insere; com PK `generated`, atualiza (`ErrNotFound` se não existe); com PK natural (sem `generated`, composta inclusive), insere ou atualiza (upsert). Com `version`, só grava se a versão ainda é a da entidade, e a incrementa; senão, `ErrConflict` |
| `Count(ctx, args)` | quantas linhas batem com o filtro (`nil` conta todas) |
| `Exists(ctx, args)` | se alguma linha bate com o filtro |
| `Delete(ctx, id)` | apaga a linha, ou devolve `ErrNotFound` |

**Atualizar no estilo JPA:** carregue, altere e salve. Use uma transação quando a leitura e a escrita precisam ser atômicas:

```go
err := s.Transaction(ctx, func(ctx context.Context) error {
	u, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	u.Active = false // um zero que o Update ignoraria
	_, err = s.Repo.Save(ctx, u)
	return err
})
```

No insert, o `Save` grava todos os campos, então só a PK recebe o default do banco. Para usar os defaults das outras colunas, use o `Create` com um DTO.

**Queries próprias, no estilo PDO**, com parâmetros nomeados `@nome`:

```go
type Repository struct {
	database.BaseRepository[User]
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (User, error) {
	return r.Prepare("select * from users where username = @username").
		Bind("username", username).
		One(ctx)
}
```

- `All(ctx)`, `One(ctx)` e `Exec(ctx)` correspondem a fetchAll, fetch e rowCount.
- **Escolher os campos:** os métodos prontos usam `select *`; para trazer só algumas colunas, escreva-as no `Prepare`. `database.Prepare[T](conn, sql)` mapeia o resultado para qualquer tipo: um DTO com parte das colunas (`Prepare[Summary](r.DB, "select id, name from users")`), ou um valor simples como `Prepare[int64](r.DB, "select max(profile) from users")`.
- `conn` pode ser o pool ou uma transação (`pgx.Tx`). Com o pool, a transação do `ctx` (veja abaixo) tem prioridade.

**Transações, no estilo `@Transactional`:** embuta `database.Transactional` no componente e envolva os métodos que precisam ser atômicos:

```go
type Service struct {
	database.Transactional // injeta o pool, como um inject:""
	Users  *Repository `inject:""`
	Audits *audit.Service `inject:""`
}

func (s *Service) Register(ctx context.Context, req dtos.CreateRequest) (User, error) {
	var u User
	err := s.Transaction(ctx, func(ctx context.Context) error {
		var err error
		if u, err = s.Users.Create(ctx, req); err != nil {
			return err
		}
		return s.Audits.Log(ctx, "user created", u.ID)
	})
	return u, err
}
```

- **A transação viaja no `ctx`.** Tudo que roda com o `ctx` recebido pela função usa a transação: o `BaseRepository`, o `r.Prepare` e os services de outros preás. Os repositórios não mudam.
- **Commit** quando a função devolve `nil`. **Rollback** em erro ou panic, e o panic continua subindo.
- **Aninhada**, uma transação dentro de outra abre um savepoint: o erro desfaz só o trabalho interno, e a externa decide se segue ou devolve o erro.
- **Cuidados:**
  - use o `ctx` da função, não o de fora, senão a query roda fora da transação;
  - uma transação é uma conexão, então não use o `ctx` dela em várias goroutines ao mesmo tempo;
  - acesso direto ao pool (`r.DB.Query`) fica fora da transação: use `r.Prepare`.

**Nomes das colunas:** o campo vira snake_case (`OwnerID` → `owner_id`). A tag `db:"nome"` sobrescreve e `db:"-"` ignora o campo. Nomes de tabelas e colunas são sempre escapados.

**Erros do Postgres viram erros do framework:**

| Erro do Postgres | Erro do framework | HTTP |
|---|---|---|
| unique violation | `ErrConflict` | 409 |
| valor malformado (ex.: uuid inválido), not null, foreign key, check | `ErrInvalidInput` | 400 |
| nenhuma linha | `ErrNotFound` | 404 |

### Middlewares

Existem dois tipos de middleware:

- **Componente** (padrão do `prea g m <Nome>`): uma struct com campos `inject` e um método `Handle(c *gin.Context)`. O container gerencia como qualquer componente: cria, injeta e chama `Init`/`Close`. Use quando o middleware precisa de um service, do banco ou de config.
- **Função** (`prea g m <Nome> --func`): um `core.Middleware`, que é o mesmo que `gin.HandlerFunc`, sem injeção.

```go
// componente: middleware/auth.go
type Auth struct {
	Sessions Sessions `inject:""` // interface; user.Service a implementa
}

func (a *Auth) Handle(c *gin.Context) { ... }
```

Onde cada um pode ser usado:

| Nível | Função | Componente |
|---|---|---|
| **global** (`preas.Middlewares` e `Security.Middlewares`) | `core.RequestLogger` | `middleware.RateLimit{}` |
| **módulo** (`Module.Middlewares`, vale para todas as rotas do módulo) | `gin.BasicAuth(contas)` | `middleware.Auth{}` |
| **rota** | `g.GET("", middleware.Owner, h.list)` | injetado no handler: `g.GET("", h.Auth.Handle, h.list)` |

```go
var Module = &core.Module{
	Name:        "admin",
	Providers:   []any{Handler{}},
	Middlewares: []any{core.Middleware(audit), middleware.Auth{}}, // funções e componentes, na ordem
}
```

- **Por módulo**, o componente é criado automaticamente: não precisa repetir em `Providers`.
- **Global**, o componente fica num módulo implícito `app` que importa todos os módulos: injeta os `Globals` e os exports (`di.Export`) de qualquer módulo, e passa por `Init`/`Close`. Se aparece nas duas listas globais, é a mesma instância.
- **Por rota**, declare `middleware.Auth{}` em `Providers` e injete no handler (``Auth *middleware.Auth `inject:""` ``).
- **Nos dois ao mesmo tempo**, é a mesma instância.
- **Erros na subida, não em produção:** declarar `&middleware.Auth{}` ou uma struct sem `Handle`.
- **Cuidado com o atalho** `(&middleware.Auth{}).Handle`. Ele compila, mas cria um `Auth` fora do container e sem injeção, e dá panic na primeira requisição. Passe sempre o valor `middleware.Auth{}`.

Para depender de outro preá sem ciclo de imports, use uma interface: o `auth` declara um `Sessions`, e o `user.Service` o implementa.

A ordem de execução é **segurança → global → módulo → rota → handler**. Chame `c.Next()` para continuar ou `c.Abort()` para parar a cadeia.

### Logs

O pacote `logger` tem quatro níveis, em ordem: **debug < info < success < error**.

```go
logger.Debug("cache miss", "key", k)
logger.Info("enviando e-mail", "to", u.Email)
logger.Success("usuário criado", "id", u.ID)
logger.Error("falha no pagamento", "error", err)
```

```
time=2026-10-10T03:09:47.989-03:00 level=SUCCESS msg=request method=POST path=/users status=201 latency=18.6ms ip=::1
```

- **Argumentos** são pares chave/valor. A saída é estruturada (`log/slog`), em texto ou JSON.
- **`logging.level`** (veja [Configuração](#configuração)) define o nível mínimo, e o padrão é `info`. Com `debug`, aparece também a saída do gin (rotas registradas, avisos).
- **`logging.format: json`** é a opção para produção ou para coletores de log; o profile `prod` já usa.
- **`logger.Configure(level, format)`** aplica a config na subida; antes disso, o logger usa `info` em texto.
- **`logger.Log`** é o `*slog.Logger` por trás e também o default do `slog`. Assim, o `log.Printf` e o `slog.Info` de qualquer lib caem no mesmo lugar. Para um logger próprio, como em testes ou outro destino, use `logger.New(w, level, json)`.
- **`core.RequestLogger`** gera uma linha por requisição: **success** abaixo de 400, **info** para 4xx e **error** para 5xx. Numa linha de 500 vem junto o erro real que foi passado ao `core.Error`; o cliente só vê "internal server error".
- **`core.Recovery`** transforma um panic em 500 e loga o erro com a stack trace.

### Erros e respostas

Erros de HTTP usam as constantes do `net/http`: o service ou o handler devolve um `core.NewError(status, mensagem)`, e o handler chama `core.Error(c, err)`, que responde com esse status e a mensagem:

```go
if u.ID != claims.Payload.ID {
	return core.NewError(http.StatusForbidden, "not the owner")
}
```

O `core.Error` responde com o status do primeiro erro da cadeia que tem o método `HTTPStatus() int`, e com a mensagem desse erro. Assim os pacotes do framework declaram os próprios erros sem que o `core` os importe:

| Erro | HTTP |
|---|---|
| `*core.HTTPError` (`core.NewError`) | o status do erro |
| `database.ErrInvalidInput` | 400 |
| `database.ErrNotFound` | 404 |
| `database.ErrConflict` | 409 |
| `security.ErrInvalidToken` | 401 |
| o seu, com `HTTPStatus() int` | o que o método devolver |
| qualquer outro | 500, com mensagem genérica |

O detalhe embrulhado em volta (`fmt.Errorf("%w: ...", database.ErrInvalidInput)`, a mensagem do Postgres) não chega ao cliente, que recebe só `"invalid input"`. O erro inteiro vai para o log da requisição.

Os erros chegam ao cliente como `{"text": "...", "code": 2}`.

**Respostas.** Todas usam o mesmo envelope `{"text", "code", "data"}`, montado com qualquer combinação de partes. Go não tem overload de métodos; o equivalente aqui é uma função que aceita partes variádicas:

```go
core.Reply(c, 200, core.Text("salvo"))                           // {"text":"salvo","code":1}
core.Reply(c, 200, core.Code(core.WarningCode))                  // {"code":2}
core.Reply(c, 200, core.Text("salvo"), core.Code(core.WarningCode))
core.Reply(c, 201, core.Text("criado"), core.Data(user))         // {"text":"criado","code":1,"data":{...}}
core.Reply(c, 200, core.Data(users))                             // {"code":1,"data":[...]}

core.Response(c, 200, qualquerCoisa)                             // formato próprio, sem envelope
```

Sem `Code`, ele vem do status HTTP: abaixo de 400 é `SuccessCode` (1), abaixo de 500 é `WarningCode` (2) e o resto é `DangerCode` (3). O JSON sai indentado só em modo debug.

### Servidor

O servidor vem da config. `server.port` e `server.mode` (`debug`, `release` ou `test`) são obrigatórias; os timeouts têm padrão, e o yaml só sobrescreve o que define:

```yaml
server:
  port: 8000
  mode: debug
  read_header_timeout: 5s   # padrão 5s (proteção contra slowloris)
  read_timeout: 30s         # padrão 30s, corpo incluído
  write_timeout: 30s        # padrão 30s; aumente para streaming e downloads longos
  idle_timeout: 120s        # padrão 120s, conexões keep-alive
  shutdown_timeout: 15s     # padrão 15s, tempo para as requisições em andamento terminarem
  max_header_bytes: 1048576 # padrão 1 MB
  trusted_proxies: []       # padrão nenhum; o Security.TrustedProxies do código tem precedência
```

**Subida e desligamento:**
1. Na subida, o banco recebe um `Ping`, então um DSN errado falha na hora, e não na primeira requisição.
2. Em SIGINT/SIGTERM, o servidor para de aceitar conexões e espera as requisições em andamento por até `ShutdownTimeout`.
3. Só depois o pool do banco é fechado. Nenhuma query em andamento perde a conexão.
4. Se o tempo acaba, as conexões são fechadas e o contexto das requisições é cancelado. Por isso as queries devem usar `c.Request.Context()`.
5. Um segundo Ctrl+C encerra na hora.

O tamanho do pool de conexões vai no próprio DSN: `?pool_max_conns=20` (o padrão é o maior entre 4 e o número de CPUs).

### Segurança HTTP

A configuração fica na `var Security` do `preas/preas.go` e roda antes de qualquer middleware. As origens do CORS ficam aqui, no código, e não no profile:

```go
var Security = core.Security{
	CORS: &core.CORS{
		AllowOrigins: []string{"http://localhost:3000"}, // "*" libera qualquer origem
		// AllowMethods, AllowHeaders, ExposeHeaders, AllowCredentials, MaxAge
	},
	CSRF: &core.CSRF{
		// TrustedOrigins: []string{"https://admin.exemplo.com"},
		// Bypass:         []string{"POST /webhooks/"},
	},
	Headers: core.DefaultHeaders, // nosniff, X-Frame-Options, CSP, HSTS...
	// TrustedProxies: []string{"10.0.0.0/8"},
}
```

| Campo | O que faz | `nil` |
|---|---|---|
| `CORS` | responde o preflight (`OPTIONS`) e coloca os headers `Access-Control-*` nas origens permitidas | sem CORS: o navegador bloqueia chamadas de outras origens |
| `CSRF` | bloqueia com 403 requisições não seguras (POST, PATCH...) vindas de outra origem pelo navegador, usando o `net/http.CrossOriginProtection`; as origens do CORS já são confiáveis; curl e chamadas servidor a servidor passam | sem proteção |
| `Headers` | headers colocados em toda resposta; use `maps.Clone(core.DefaultHeaders)` para ajustar | nenhum |
| `TrustedProxies` | proxies cujo `X-Forwarded-For` é aceito no `c.ClientIP()` | nenhum proxy é confiável |

`AllowCredentials` com origem `"*"` é recusado na subida.

**Configurações próprias.** Para casos que o framework não cobre:

```go
var Security = core.Security{
	// regra própria de origem (subdomínios, consulta ao banco...); o CSRF também confia nela
	CORS: &core.CORS{AllowOriginFunc: func(o string) bool { return strings.HasSuffix(o, ".exemplo.com") }},

	// resposta própria quando o CSRF bloqueia (precisa chamar c.Abort)
	CSRF: &core.CSRF{Deny: func(c *gin.Context) { c.AbortWithStatusJSON(403, gin.H{"erro": "origem"}) }},

	// suas proteções: rodam depois de Headers, CORS e CSRF e antes dos middlewares globais
	Middlewares: []any{middleware.RateLimit{}, middleware.CSRFToken}, // componentes e funções
}
```

- **Peças avulsas:** `core.Headers(map)`, `(&core.CORS{...}).Middleware()` e `(&core.CSRF{...}).Middleware()` funcionam por módulo ou por rota, por exemplo um CORS diferente só em `/public`.
- **Desligar uma peça:** deixe o campo `nil` e coloque a sua versão em `Middlewares`, ou use outra lib, como o `gin-contrib/cors`.

### Autenticação

O pacote `security` traz as peças:
- **JWT:** `security.Jwt[M]` cria e valida tokens HS256 (só esse algoritmo é aceito), com o payload `M` nas claims. A chave vem de `security.jwt.secret` e precisa ter 32 bytes ou mais. Sem ela, a app sobe normalmente, e só criar ou validar um token dá erro.
- **Senhas:** `security.HashPassword` e `security.ComparePassword`, com argon2id.
- **Erros:** um token inválido responde 401 pelo `core.Error`.

O middleware de autenticação, as sessões revogáveis e os papéis entram com a lib de segurança (veja o [ROADMAP](ROADMAP.md)).

## CLI

| Comando | O que faz |
|---|---|
| `prea new <nome> [--module] [--with-db]` | cria um projeto (veja [Estrutura do projeto](#estrutura-do-projeto)). `--module` é o caminho do módulo Go, que vai no `go.mod` e nos imports (padrão: o nome); `--with-db` liga o banco nos profiles e gera um preá de CRUD; `--local <dir>` usa um clone do framework em vez de baixá-lo |
| `prea dev [--profile]` | compila e sobe o projeto (profile `dev` por padrão) e o reinicia quando um `.go`, o `config/` ou o `.env` muda. Um erro de compilação mantém a versão que está rodando |
| `prea run [--profile]` | compila e sobe o projeto, repassando os sinais para o desligamento gracioso |
| `prea build [flags]` | gera o binário em `bin/<nome>`; as flags vão para o `go build` (`-o`, `-trimpath`...), e o `GOOS`/`GOARCH` vêm do ambiente |
| `prea tests [flags] [pacotes]` | `go test` com o profile `tests`; sem pacotes, `./...` |
| `prea g p <Nome>` | cria `preas/<nome>/` (module, handler, service, repository, model e dtos) e registra o módulo em `preas/preas.go` |
| `prea g m <Nome> [--func]` | cria `middleware/<nome>.go`: um componente (com injeção), ou uma função com `--func` |
| `prea migrate create <nome>` | cria `migrations/<próxima versão>_<nome>.sql`, para escrever à mão |
| `prea migrate up [versão]` | aplica as pendentes, ou até a versão |
| `prea migrate down [n]` | desfaz as últimas `n` aplicadas (padrão 1) |
| `prea migrate status` | lista as migrations: aplicadas (com a data) ou pendentes |
| `prea migrate validate` | confere os arquivos sem tocar no banco; para CI |

Os `prea migrate` aceitam `--profile` (`prea migrate --profile dev up`). Eles compilam o projeto e o rodam em **modo de inspeção** (`PREA_MIGRATE`, a `database.MigrateEnv`): o `prea.Run` vê a variável e, em vez de subir o servidor, roda o `database.Command` com a própria config e o próprio banco. Assim a CLI não precisa saber onde estão a config e a conexão. Isso vem antes da checagem do segredo JWT, então uma migration não o exige.

O `g` é atalho de `generate`, o `p` de `prea` e o `m` de `middleware`. Rode a CLI na raiz do projeto: ela lê o `go.mod` para descobrir o caminho dos imports.

**Nomes:** letras, em camel case ou separadas por `_`. `OrderItem` e `order_item` são a mesma coisa; números não são aceitos (`order123`). Cada parte do código recebe a forma que a convenção pede:

| `prea g p order_item` | |
|---|---|
| pacote e pasta | `orderitem` (pacote Go não usa `_`) |
| tipos | `OrderItem` |
| tabela | `order_item` (que você cria no banco) |
| rotas | `/order-item` (GET, GET `/:id`, POST, PATCH `/:id`, DELETE `/:id`) |

No middleware, o arquivo segue a tabela: `prea g m rate_limit` cria `middleware/rate_limit.go` com o tipo `RateLimit`.

**Antes de gerar, a CLI confere tudo; se algo falha, o projeto não é alterado:**
- o nome não pode ser uma palavra-chave do Go (`type`);
- o pacote não pode ter o nome de um import do `preas.go` (`core`, `time`);
- o preá ou o arquivo do middleware não pode existir;
- o middleware não pode ter o nome de algo já declarado no pacote (`Owner`);
- o `preas.go` precisa ter o bloco de imports entre parênteses e o `var Modules = []*core.Module{...}`.

## Configuração

No estilo do Spring: `application.yaml` é a base, e `application-<profile>.yaml` a completa para o profile ativo. Os arquivos são da aplicação e ficam em `config/`, na raiz do projeto. O carregamento usa o [koanf](https://github.com/knadh/koanf).

```yaml
# config/application.yaml
server:
  port: 8000
database:
  url: ${DATABASE_URL}     # vem do ambiente ou do .env
security:
  jwt:
    secret: ${JWT_SECRET}
```

- **Profile ativo:** um só, vindo de `PREA_PROFILE=dev` ou de `prea run --profile dev`. Sem profile, só a base é carregada; um arquivo que não existe é ignorado.
- **Ambiente:** só entra onde o yaml o referencia, com `${VAR}`. O `.env` da pasta em que a app roda é lido antes (veja o `.env.example` do projeto) e não substitui o que já está no ambiente. Segredos ficam no ambiente, nunca no yaml commitado.
- **Chave ausente:** uma chave que não está no yaml, ou um `${VAR}` não definido, deixa o valor zerado (`nil` em ponteiros). Quem precisa do valor valida, como o `security.SetSignKey`, que impede a subida com um segredo curto.

As chaves que o `prea.Run` lê:

| Chave | Obrigatória | Padrão |
|---|---|---|
| `server.port`, `server.mode` | **sim** | — |
| `server.*` (timeouts, `max_header_bytes`, `trusted_proxies`) | não | veja [Servidor](#servidor) |
| `logging.level`, `logging.format` | não | `info`, `text`; o `prea.Config.Logging` tem precedência |
| `database.url` | não | sem ela, a app sobe sem banco |
| `database.schema`, `database.migrations.*` | não | veja [Model e schema](#model-e-schema) |
| `security.jwt.secret` | não | aplicada só se existir |

**Nos componentes**, a tag `value:"chave"` preenche o campo na subida, como o `@Value` do Spring:

```go
type Mailer struct {
	Addr    string        `value:"mail.addr"`
	Timeout time.Duration `value:"mail.timeout"` // "30s" no yaml
	Mail    MailConfig    `value:"mail"`         // o bloco inteiro, como o @ConfigurationProperties
}

type MailConfig struct {
	Addr string `yaml:"addr"`
	From string `yaml:"from"`
}
```

- Os valores são convertidos para o tipo do campo: `"true"` → bool, `"8000"` → int, `"30s"` → `time.Duration`.
- Como no `inject`, um campo preenchido na declaração é mantido (`Mailer{Addr: "localhost:25"}` em testes).
- O `*config.Config` também é injetável (``Config *config.Config `inject:""` ``), por exemplo para ler `cfg.Profile()`.

**Módulos por profile:** `Profiles` restringe o módulo a esses profiles. Vale para os módulos listados em `core.App.Modules`:

```go
var Module = &core.Module{Name: "debug", Providers: []any{Handler{}}, Profiles: []string{"dev"}}
```

## Testes

### De uma app

Os testes de cada preá ficam **ao lado do código** (`preas/item/item_test.go`, como `package item`): enxergam o que não é exportado e montam só o próprio módulo. Os de ponta a ponta, com a app inteira, ficam em `tests/`. O `prea tests` roda os dois com o profile `tests` (`config/application-tests.yaml`, com o banco de testes e `schema: create-drop`).

```go
func TestHealth(t *testing.T) {
	app := prea.Test(t, prea.Config{Modules: []*core.Module{Module}})
	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /health: %d %s", w.Code, w.Body)
	}
}
```

O `prea.Test` monta a app como o `prea.Run`, mas com o profile `tests`, e devolve o handler para o `httptest`. Ele roda a partir da raiz do projeto, o primeiro `go.mod` acima da pasta do teste, e fecha a app e o banco no fim do teste.

### Do framework

O `./...` na raiz cobre só o módulo da raiz; use o caminho, que cobre todos os módulos do `go.work`:

```bash
go test github.com/prea-framework/prea/...

# sem os testes que compilam projetos gerados
go test -short github.com/prea-framework/prea/...

# inclui os testes de integração do banco (criam as próprias tabelas com o database.schema)
PREA_TEST_DATABASE_URL="postgres://user:pass@localhost:5432/prea_test?sslmode=disable" go test github.com/prea-framework/prea/...
```

A CLI é testada em camadas (`cmd/prea`):
- **unidade:** formas dos nomes, nomes inválidos e a edição do `preas.go` em vários formatos (lista vazia, numa linha, com comentários, em grupo `var (...)`, import com alias);
- **comportamento:** gera num projeto temporário e confere arquivos, registro e conteúdo; todo comando que falha deixa o projeto intacto. O `build` e o `tests` são conferidos com um `go` falso, que grava os argumentos e o profile;
- **projeto gerado:**
  - `prea new --local` cria projetos ligados a este clone, com e sem `--with-db`;
  - o teste gera preás e middlewares neles e roda `go vet` e `prea tests`; com o banco, o CRUD inteiro pelo HTTP;
  - `prea build`, e o binário respondendo `/health`;
- **`prea dev`:** num projeto novo, ele sobe, reinicia quando um handler muda, mantém a versão antiga num erro de compilação e para o servidor no fim.

## Licença

Veja [LICENSE](LICENSE).
