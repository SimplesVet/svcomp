# svcompare (svcomp)

CLI em Go para comparar dois schemas MySQL e gerar o SQL para sincronização.

Observação: o nome do binario/comando atual do projeto é `svcomp`.

## Requisitos

- Go 1.23+
- MySQL acessível para source e target
- Docker + Docker Compose (para testes de integracao)

## Build

```bash
make build
```

Binario gerado em:

```bash
./bin/svcomp
```

## Uso da linha de comando

Comando principal:

```bash
./bin/svcomp compare --source "<DSN_SOURCE>" --target "<DSN_TARGET>" [--output arquivo.sql] [--dry-run]
```

Flags:

- `--source` (obrigatoria): DSN do banco source
- `--target` (obrigatoria): DSN do banco target
- `--output` (opcional): arquivo para salvar o SQL gerado
- `--dry-run` (opcional): imprime o SQL no stdout

Ajuda:

```bash
./bin/svcomp --help
./bin/svcomp help compare
```

## Formato DSN (MySQL)

Exemplo de DSN valido:

```bash
user:senha@tcp(host:3306)/nome_do_schema?parseTime=true
```

Exemplo local usando ambiente de integracao:

- Source: `root:root@tcp(127.0.0.1:3307)/svcomp_source?parseTime=true`
- Target: `root:root@tcp(127.0.0.1:3308)/svcomp_target?parseTime=true`

## Exemplos de uso

Gerar SQL e imprimir no terminal:

```bash
./bin/svcomp compare \
  --source "root:root@tcp(127.0.0.1:3307)/svcomp_source?parseTime=true" \
  --target "root:root@tcp(127.0.0.1:3308)/svcomp_target?parseTime=true" \
  --dry-run
```

Gerar SQL e salvar em arquivo:

```bash
./bin/svcomp compare \
  --source "root:root@tcp(127.0.0.1:3307)/svcomp_source?parseTime=true" \
  --target "root:root@tcp(127.0.0.1:3308)/svcomp_target?parseTime=true" \
  --output sync.sql
```

## Ambiente de testes de integracao

O projeto possui um ambiente Docker com dois MySQL e schemas de exemplo:

- `mysql-source` na porta `3307`
- `mysql-target` na porta `3308`
- Schemas SQL em `testdata/integration/schemas/`

Comandos uteis:

```bash
make integration-up          # sobe os containers
make integration-reset       # remove os schemas
make integration-restore     # remove e restaura schema1/schema2
make test-integration        # sempre remove, restaura e roda os testes
make integration-down        # derruba containers mantendo volumes
make clean-integration       # derruba containers e remove volumes/orfaos
```

## Testes

Rodar todos os testes:

```bash
make test
```

Rodar testes de integracao (com preparo automatico dos schemas):

```bash
make test-integration
```
