# MB CLI

Ferramenta de linha de comando para gerir plugins e comandos personalizados no seu ambiente.

<img title="MB CLI" alt="MB CLI" src="mb-cli.png">

## Documentação

Tudo sobre instalação, uso diário, plugins e opções está no site:

**https://carlosdorneles-mb.github.io/mb-cli/**

Sugestão de entrada: [Começar](https://carlosdorneles-mb.github.io/mb-cli/docs/getting-started).

## Desenvolvimento local

É necessário [Go](https://go.dev/dl/) (versão em `go.mod`) e `make`.

```bash
git clone https://github.com/carlosdorneles-mb/mb-cli.git
cd mb-cli
make build          # gera bin/mb
./bin/mb --help
```

Para executar sem gravar o binário: `make run-local -- --help` (ou `make run` depois de `make build`). Testes: `make test`.

Opcional — registar os plugins de exemplo do repositório e atualizar o cache:

```bash
make install-plugins-examples
./bin/mb plugins sync
```

## Contribuir

Abre um *issue* para ideias ou bugs. Para código ou documentação: *fork*, ramo com alterações focadas e *pull request* contra `main`. O CI e a revisão seguem o fluxo habitual do repositório; pormenores de versões e release estão na documentação do projeto.
