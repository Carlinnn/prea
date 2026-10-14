# Roadmap do Preá

Este documento lista os **gaps** do que já foi implementado (levantados lendo o código em 2026-10-10). Os **próximos passos** estão no fim, um por gap aberto.

Severidade dos gaps: **Crítica** = quebra o isolamento ou a correção do framework; **Alta** = risco de segurança ou perda de dados; **Média** = afeta produção ou uso real; **Baixa** = qualidade ou ergonomia.

---

## Gaps

### Segurança

| Sev. | Gap | Onde |
|---|---|---|
| Baixa | Sem refresh token: quando o token expira, é preciso fazer login de novo. | `auth` |

**Resolvidos:**
- O segredo do JWT (`security.jwt.secret`) é obrigatório, com pelo menos 32 bytes. Sem ele, a app não sobe (`security.SetSignKey` falha no `main`) e `CreateToken`/`DecodeToken` devolvem erro. Os testes definem `security.SignKey`.
- `DecodeToken` aceita só HS256 (`jwt.WithValidMethods`).
- CORS, CSRF (via `net/http.CrossOriginProtection`), headers de segurança e proxies confiáveis são configurados em `core.Security` no `preas.go`. Por padrão, nenhum proxy é confiável.
- Existe o 403 (`core.NewError(http.StatusForbidden, ...)`). O exemplo tem o `middleware.Owner`: só o dono altera ou apaga o próprio usuário.
- Revogação: o `jti` do token fica em `users.token` e o `middleware.Auth` confere com o banco. Um novo login invalida o token anterior, e existe o `POST /auth/logout`.
- O modo do gin vem de `server.mode`, que o `main` aplica depois de carregar a config. Antes, o gin lia o `GIN_MODE` antes de o `.env` ser carregado.

**Decisões (não são gaps):**
- O cadastro com `profile`, a listagem aberta para quem está logado e o login sem rate limit fazem parte do exemplo, que é propositalmente simples.
- Quem decide se o token tem `exp` é quem codifica o token.
- As rotas geradas são públicas; para fechá-las, use um middleware.

### Servidor (`core`)

| Sev. | Gap | Onde |
|---|---|---|
| Baixa | Erros de validação devolvem a mensagem crua do validator (`Key: 'CreateRequest.Email' Error:Field validation...`). Ver o passo 8. | handlers |

**Resolvidos:**
- Graceful shutdown (`core/server.go`): no SIGINT/SIGTERM o servidor para de aceitar conexões, espera as requisições em andamento (`ShutdownTimeout`) e só então retorna, então o pool do banco fecha depois. Se o tempo acaba, as conexões são fechadas e os contextos das requisições são cancelados. Testado com `-race`. O `prea run` roda o binário direto, porque o `go run` não repassa SIGTERM.
- Os timeouts do `http.Server` são configuráveis em `preas.Server` e têm padrões seguros.
- O banco recebe um `Ping` na subida, com timeout de 5s.
- Respostas: o envelope `{text, code, data}` é montado com `core.Reply` e partes (`Text`, `Code`, `Data`); o `core.Response` serve para formatos próprios. O JSON só sai indentado em modo debug.
- Logs: o pacote `logger` (slog) tem os níveis debug, info, success e error, configurados por `logging.level` e `logging.format`. O `core.RequestLogger` escolhe o nível pelo status e loga o erro real dos 500, e o `core.Recovery` loga panics com stack. A saída do gin vira debug.
- A config de segurança (headers, origens) é copiada quando o middleware é criado, então alterá-la depois não gera data race.

**Decisões (não são gaps):** health check, request id e métricas dependem da aplicação, e o usuário cria como módulo ou middleware.

### Configuração

| Sev. | Gap | Onde |
|---|---|---|
| — | Nenhum gap aberto. | |

**Resolvido (passo 1):** as variáveis de ambiente eram lidas em lugares diferentes, e o `.env` era carregado por um import com efeito colateral (`common/env.go`, removido). Agora tudo passa pelo pacote `config`, carregado uma vez no `main`; o `logger` e o `security` recebem a config na subida (`logger.Configure`, `security.SetSignKey`).

### Injeção de dependência (`di`)

| Sev. | Gap | Onde |
|---|---|---|
| — | Nenhum gap aberto. | |

**Resolvidos:**
- **Middleware como componente:**
  - `Module.Middlewares` aceita funções e componentes (`middleware.Auth{}`); o componente é criado, injetado e passa por `Init`/`Close` como qualquer outro, e é a mesma instância se também estiver em `Providers`;
  - `preas.Middlewares` e `Security.Middlewares` também aceitam componentes: eles vivem num módulo implícito `app` que importa todos os módulos, então injetam os `Globals` e os exports de qualquer módulo;
  - erros (ponteiro, struct sem `Handle`, dependência faltando) aparecem na subida;
  - o `prea g m` gera componente por padrão, e função com `--func`.
- **A tag curta `inject` foi removida.** Ela não era uma tag Go válida (o `go vet` acusava) e agora dá erro na subida com a correção; só `inject:""` é aceita.
- **Ciclo de vida:**
  - hooks opcionais `Init(ctx) error` e `Close(ctx) error`/`io.Closer`;
  - o `Init` roda na ordem das dependências, inclusive entre módulos, e o `Close` na ordem inversa, no desligamento gracioso e antes do pool do banco fechar;
  - se um `Init` falha, a app não sobe e o que já tinha iniciado é fechado;
  - o `Init` cobre o caso de "provider por construtor": lógica de criação que pode falhar;
  - para testes, existe o `core.App.Build(ctx)`, que devolve o engine e o `closer`.
- **Componentes nomeados**, no estilo dos tokens do Angular:
  - `di.Named("analytics", x)` em providers e globals, e `inject:"analytics"` no campo;
  - `inject:""` prefere o componente sem nome;
  - **comportamento corrigido:** antes, com dois candidatos do mesmo tipo, o primeiro era escolhido em silêncio. Agora isso é erro na subida, e a mensagem pede um nome.

**Resolvido (Crítica):**
- **O problema:** os módulos são variáveis de pacote, e os providers eram instâncias criadas uma vez por processo (`new(Repository)`). Um segundo `App` reaproveitava as instâncias, e como campos já preenchidos são mantidos, continuava com as dependências do primeiro container, sem erro nenhum. Dois containers em paralelo escreviam nos mesmos campos, o que é uma data race.
- **A solução:** os providers passaram a ser declarados como valor (`Repository{}`) e funcionam como molde. A cada `Load`, o container cria a instância com `reflect.New` e copia o molde. O módulo vira só declaração e nunca é alterado. Declarar um ponteiro dá erro com a correção na mensagem.
- **Testes:**
  - `di.TestIsolation`: 8 containers em paralelo com `-race`, cada um com o seu banco, nenhum componente compartilhado e a declaração intacta;
  - `core.TestAppsAreIsolated`: dois `App` montados a partir do mesmo módulo. Esse teste falha no código antigo (`app with "test" answered "prod"`).

**Plano futuro:** memória por requisição, como o `SecurityContextHolder.getContext()` e o `@RequestScope` do Spring, que guardam o dado num `ThreadLocal` acessível de qualquer lugar. Por enquanto o framework só tem singletons por container, e o estado de requisição vai explicitamente no `context.Context`/`gin.Context`, como as claims do `Auth`. Em Go não existe `ThreadLocal`.

### Banco (`database`)

| Sev. | Gap | Onde |
|---|---|---|
| Média | Só Postgres, e os tipos do pgx aparecem na API pública (`Conn`, `BaseRepository.DB *pgxpool.Pool`). **Decisão:** resolvido pela Prea Persistence (passo 4): o banco vem do `application.yaml` e cada dialect é um módulo. | `database` |
| Média | `GetAll` não tem paginação nem limite, então a resposta cresce sem teto. **Decisão:** resolvido pelo specification util com paginação (passo 9). | `database/repository.go` |
| Baixa | Tabela e rota usam o nome do pacote no singular. | templates |

**Resolvidos:**
- **Migrations (passo 2):** antes só havia um `.sql` solto, e depois o `schema: update`, sem histórico. Agora há migrations versionadas (semver ou timestamp) rodadas pelo goose, com `-- prea:down` opcional, geradas a partir dos models na subida (`from_models`) e aplicadas na subida (`on_start`) ou pelo `prea migrate`.
- **Tags de schema (passo 3):** o model descreve o schema com uma tag por opção (`pk`, `generated`, `size`, `type`, `default`, `unique`, `index`, `fk`, `ondelete`, `version`). Antes:
  - a PK era sempre `id`; agora qualquer campo pode ser PK, inclusive composta, e o `generated` diz se o banco a gera (`auto_increment`, `uuid`, `auto`);
  - o mapeamento *lax* escondia um campo sem coluna; agora o `database.schema: validate` impede a subida quando a tabela não bate com o model. O mapeamento continua *lax* de propósito, porque o estrito quebraria o `select *` com colunas extras;
  - o schema era escrito à mão (`user.sql`, removido); agora `database.schema` aplica os models na subida (`update`, `create`, `create-drop`), como o `ddl-auto` do Hibernate;
  - não havia lock otimista; agora o `Save` com `version` devolve `ErrConflict` para uma entidade desatualizada.
- **Update que perdia zeros (Baixa → Alta):** um `Update` com struct ignorava campos zerados, então `Active = false` ou `Count = 0` sumiam sem erro, uma perda de dados silenciosa. Agora:
  - `Save(ctx, entidade)`, no estilo do `save` do JPA: grava todas as colunas, zeros inclusive; insere com PK zerada, senão atualiza (`ErrNotFound` se a linha não existe). O fluxo é carregar, alterar e salvar, dentro de uma `Transaction` quando precisa ser atômico;
  - o `Update` continua sendo *patch*, mas o DTO pode usar ponteiros: `nil` não mexe, e um ponteiro para zero grava o zero (testado). O template do `UpdateRequest` explica isso;
  - lock otimista com `version:""` (passo 3).
- **`Count` e `Exists`:** com filtro `Args` (`nil` = tabela toda), usando o mesmo `where` do `Where`/`First`.
- **Seleção de campos:** os métodos prontos usam `select *`; para escolher as colunas, use o `Prepare`, que mapeia para o model ou para um DTO de projeção (`Prepare[Summary](r.DB, "select id, name ...")`).
- **Transação (`@Transactional`):** `database.Transactional` embutido no componente marca quais métodos são atômicos com `s.Transaction(ctx, func(ctx) error {...})`.
  - A transação viaja no `ctx`: todo `Statement` do mesmo pool que roda com esse `ctx` usa a transação, inclusive o `BaseRepository` e as queries de `r.Prepare`, sem mudar os repositórios.
  - Commit quando a função devolve `nil`; rollback em erro ou panic (o panic continua subindo).
  - Chamada dentro de outra transação, abre um savepoint: o erro desfaz só o trabalho interno e a externa decide se segue.
  - A chave da transação no `ctx` é o pool, então com mais de um banco uma transação não captura queries de outro.
  - Exemplo em `example/internal/preas/auth`: o `Login` grava a sessão (que revoga a anterior) e assina o token na mesma transação; se a assinatura falha, a sessão anterior continua valendo. Isso corrigia um bug real do exemplo.
  - Ficou para depois: nível de isolamento e transação só de leitura (`TransactionWith(ctx, opts, fn)`), quando houver necessidade.

### CLI

| Sev. | Gap | Onde |
|---|---|---|
| Baixa | Faltam `prea new` (scaffold de projeto) e `prea dev` (hot reload); o `prea run` e o `prea migrate` dependem de `./cmd/server`. **Decisão:** vêm junto do `prea.go` (passo 6), agora que os módulos estão separados. | `cmd/prea` |

**Resolvidos:**
- **Nomes:** letras em camel case ou separadas por `_` (`order_item`, `OrderItem`); números continuam recusados (`order123`). Cada parte recebe a forma da convenção: pacote `orderitem`, tipo `OrderItem`, tabela `order_item`, rota `/order-item` e arquivo de middleware `rate_limit.go`.
- **Validação antes de gerar:** palavra-chave do Go, pacote com o nome de um import do `preas.go`, preá ou middleware existente e identificador já declarado no pacote de middlewares. O `preas.go` novo é calculado e validado **antes** de criar qualquer arquivo; um comando que falha não altera o projeto.
- **Testes pesados do gerador** (`cmd/prea`), em três camadas:
  - unidade: nomes e a edição do `preas.go` em vários formatos;
  - comportamento: arquivos, conteúdo e registro, e o projeto intacto em todo erro;
  - projeto gerado: um projeto novo, ligado aos módulos locais do framework por `replace`, passa por `go vet` e `go test -race`; o teste de dentro confere as rotas e, com banco, faz o CRUD pelo HTTP. Mutações nos templates (tirar o `TrimSpace`, trocar a rota) são pegas.

**Decisão (não é gap):** o caminho fixo `frameworkModule = "github.com/prea-framework/prea"` está certo: cada pacote do framework é importado de `<frameworkModule>/<pacote>`.

### Estrutura e qualidade

| Sev. | Gap | Onde |
|---|---|---|
| Baixa | O teste de integração do banco só roda com env; não há docker-compose para Postgres local (o CI sobe o seu). | — |

**Resolvidos:**
- **Código morto em `common`:** o pacote foi removido, junto com o `Stoi64` e o `ThrowException`, que não eram usados. Os erros genéricos (`ErrNotFound`, `ErrUnauthorized`...) foram substituídos: erros de HTTP usam `core.NewError(http.StatusX, msg)`, com as constantes do `net/http`, e cada pacote declara os próprios erros (`database.ErrNotFound`, `database.ErrConflict`, `security.ErrInvalidToken`), que dizem o próprio status (`HTTPStatus()`) ao `core.Error`.
- **Um módulo só (passo 5):** o framework era um módulo chamado `prea`, sem domínio, e o `core` importava o `database` e o `security`, levando o pgx, o goose e o JWT a todo binário. Agora cada pacote é um módulo em `github.com/prea-framework/prea/<pacote>`, ligados por um `go.work` commitado. O `core` só depende de `di`, `config` e `logger`: os erros dizem o próprio status, e o banco entra pelo gancho `core.App.Setup`. A app de exemplo foi para `example/`, e um CI testa todos os módulos e confere os requires.

**Decisões (não são gaps):**
- A app de exemplo (`example/`) sai do repositório quando virar o template do `prea new` (passo 6).
- O acoplamento do `middleware.Auth` com `auth/dtos` é do exemplo. Os preás `user` e `auth` serão refeitos do zero, então eles não ganham testes de service e handler agora.

---

## Próximos passos

Cada gap aberto acima já tem um passo:
- só Postgres → Prea Persistence (passo 4);
- `GetAll` sem limite → specification e paginação (passo 9, dentro da persistence);
- erros de validação crus → validação de structs (passo 8);
- `prea new` e `prea dev` → `prea.go` (passo 6);
- refresh token → lib de segurança (passo 7).
