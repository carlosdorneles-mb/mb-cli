---
sidebar_position: 2
---

# Desenvolvimento Local

Se você for alterar o código ou contribuir para o MB CLI.

## Pré-requisitos

- Go 1.26.1+ (ou a versão em `go.mod`)
- `make`

## Build e instalação

```bash
make build          # binário em bin/mb
make install        # instala em $GOPATH/bin
```

## Executar localmente

Para rodar o CLI sem instalar (a partir do código-fonte):

```bash
make run-local              # go run . (ajuda: make run-local -- --help)
make run-local plugins sync # argumentos podem ser passados direto
make run plugins sync       # build + ./bin/mb
```

## Plugins de exemplo

Para usar os plugins de exemplo do repositório:

```bash
make install-plugins-examples   # regista cada plugin com mb plugins add
make run plugins sync           # ou mb plugins sync
```

## Testes

```bash
make test
```

## Arquitetura do Código

O projeto segue **Clean Architecture** com injeção de dependência via [`go.uber.org/fx`](https://pkg.go.dev/go.uber.org/fx).

```
internal/
├── bootstrap/      # Composição raiz — monta a FX App
├── cli/            # Camada de apresentação (Cobra apenas)
├── usecase/        # Casos de uso — lógica de negócio pura
├── domain/         # Modelos de domínio
├── infra/          # Adapters concretos (SQLite, Git, FS, etc.)
├── ports/          # Interfaces (contratos entre camadas)
├── module/         # Módulos Fx — wire de dependência
├── fakes/          # Test doubles para testes unitários
└── shared/         # Utilitários transversais (config, UI, version)
```

**Fluxo de dependência** (aponta para o centro):

```
cli → usecase → domain
  ↘            ↗
   infra → ports
```

Para detalhes, veja a [Referência de Arquitetura](../technical-reference/architecture.md).

Próximo passo: [Criar um plugin](../plugin-authoring/create-a-plugin.md) para montar seu primeiro plugin e rodá-lo com o MB.
