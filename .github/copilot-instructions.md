# svcomp — Instruções do Projeto

## Visão Geral

`svcomp` é uma ferramenta CLI em Go para comparar schemas MySQL entre dois bancos de dados (source e target) e gerar as instruções SQL necessárias para sincronizar o target com o source.

## Linguagem e Runtime

- **Go** (versão mínima: 1.23)
- Compilado para binário estático; sem dependência de runtime externo
- Módulo: `github.com/simplesvet/svcomp`

## Estrutura de Diretórios

```
svcomp/
├── cmd/svcomp/         # Ponto de entrada CLI (main.go + comandos cobra)
├── internal/
│   ├── config/         # Parsing de DSN e flags de conexão
│   ├── connector/      # Abertura e gerenciamento de conexões MySQL
│   ├── lister/         # Listagem de objetos do banco (tabelas, funções, etc.)
│   ├── comparator/     # Comparação source vs target; produz listas de create/update/delete
│   ├── differ/         # Geração de SQL: schemadiff p/ tabelas; DROP/CREATE p/ demais objetos
│   └── output/         # Formatação e escrita do SQL gerado (stdout ou arquivo)
├── testdata/integration/schemas/  # Schemas SQL de apoio aos testes de integração
├── docker-compose.yaml   # Ambiente local de integração com 2 MySQL
├── pkg/types/          # Tipos compartilhados (DBObject, DiffResult, ObjectKind…)
├── Makefile
├── go.mod
└── go.sum
```

## Dependências Principais

| Pacote | Uso |
|--------|-----|
| `github.com/go-sql-driver/mysql` | Driver MySQL |
| `github.com/spf13/cobra` | Framework CLI |

Dependência externa opcional:

- `schemadiff` CLI no PATH para geração de `ALTER TABLE` precisos em diffs de tabela (com fallback para `DROP+CREATE` quando indisponível)

## Objetos Gerenciados

`tables`, `views`, `functions`, `procedures`, `triggers`, `events`

## Regras de Comparação e Geração de SQL

1. **Tabelas** — Usar `schemadiff` para gerar `ALTER TABLE` precisos.
2. **Demais objetos** (view, function, procedure, trigger, event):
   - Objeto só existe no target → gerar `DROP IF EXISTS`
   - Objeto só existe no source → gerar `DROP IF EXISTS` + `CREATE`
   - Objeto existe nos dois → gerar `DROP IF EXISTS` + `CREATE` (substituição completa)
3. A ordem de geração deve respeitar dependências: tabelas primeiro, depois views, functions, procedures, triggers, events.

## Convenções de Código

- Pacotes `internal/` são privados; nunca importar diretamente de fora do módulo.
- Erros devem ser propagados com contexto: `fmt.Errorf("lister: %w", err)`.
- Nenhuma variável global de estado; passar dependências por injeção (structs com interfaces).
- Interfaces definidas no pacote consumidor, não no pacote produtor.
- Testes unitários em `*_test.go` no mesmo pacote.
- Testes de integração usam `docker compose` com dois serviços MySQL e restore dos schemas em `testdata/integration/schemas/`.

## CLI — Flags Esperadas

```
svcomp compare \
  --source "user:pass@tcp(host:3306)/db_source" \
  --target "user:pass@tcp(host:3306)/db_target" \
  [--output arquivo.sql] \
  [--dry-run]
```

## Build

```bash
make build          # compila para o OS atual
make build-all      # cross-compile linux/amd64, linux/arm64, darwin/amd64, windows/amd64
make test           # executa testes unitários
make test-integration  # executa testes com container MySQL
make clean-integration # remove containers, rede, volumes e órfãos da integração
```

## Saída

- SQL gerado deve ser idempotente e executável diretamente no MySQL.
- Cada bloco de SQL deve ser precedido por um comentário indicando o objeto e a operação.
- Em caso de `--dry-run`, apenas imprime o SQL; não executa nada.
