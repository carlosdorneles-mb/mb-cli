---
sidebar_position: 4
---

# Flags globais

O MB CLI oferece algumas flags que valem para qualquer comando e afetam o nível de saída e o ambiente. **`-v` e `-q` podem ser usados em qualquer posição:** antes ou depois do subcomando (por exemplo `mb -v tools hello` ou `mb tools hello -v`). Em ambos os casos o plugin recebe `MB_VERBOSE`/`MB_QUIET` conforme a flag.

## --verbose / -v

**O que faz:** Ativa saída mais verbosa. O CLI pode exibir mensagens adicionais de diagnóstico ou logs que ajudam a entender o que está acontecendo.

**Quando usar:** Para depurar um problema, acompanhar o fluxo de um comando ou quando você quer mais detalhes na saída.

Ao executar um **comando de plugin**, o CLI define no ambiente do processo do plugin a variável **`MB_VERBOSE=1`**. O plugin pode usá-la para exibir mais logs (por exemplo, nível debug).

Exemplo:

```bash
mb -v plugins list
mb --verbose plugins sync
mb -v tools meu-comando
```

## --quiet / -q

**O que faz:** Reduz ou suprime mensagens informativas. O CLI evita imprimir avisos ou mensagens de progresso que não sejam estritamente necessárias.

**Quando usar:** Em scripts ou quando você quer apenas o resultado (por exemplo, saída de um plugin) sem mensagens extras do próprio CLI.

Ao executar um **comando de plugin**, o CLI define no ambiente do processo do plugin a variável **`MB_QUIET=1`**. O plugin pode usá-la para suprimir logs informativos e exibir apenas o essencial (ou apenas erros).

Exemplo:

```bash
mb -q plugins list
mb --quiet plugins sync
mb -q self update --check-only   # só código de saída (ex.: 2 = há atualização), sem texto em stdout
mb -q tools meu-comando
```

## Uso em plugins de shell

Em plugins escritos em shell, você pode ler `MB_VERBOSE` e `MB_QUIET` para decidir se imprime mensagens e em qual nível. Assim o plugin respeita a preferência do usuário ao usar `-v` ou `-q`.

O CLI disponibiliza **helpers de shell** (por exemplo a função `log`) e define no ambiente do plugin a variável **`MB_HELPERS_PATH`** (diretório `~/.config/mb/lib/shell`). Os arquivos nesse diretório são criados e atualizados ao rodar **`mb plugins sync`**. Para usá-los nos plugins, no início do script faça: `. "$MB_HELPERS_PATH/all.sh"` (todos) ou `. "$MB_HELPERS_PATH/log.sh"` (só log). Depois você pode chamar `log info "mensagem"`, `log debug "detalhe"`, etc. Veja a [Referência: Helpers de shell](../technical-reference/helpers-shell.md) para a lista de helpers e como carregar.

- **`MB_QUIET=1`** — O usuário pediu saída mínima. Evite chamar `gum log` para mensagens informativas; só mostre erros se fizer sentido.
- **`MB_VERBOSE=1`** — O usuário pediu mais detalhes. Você pode incluir logs em nível debug ou mensagens de diagnóstico.

Para a função `log` e outros helpers, veja a [Referência: Helpers de shell](../technical-reference/helpers-shell.md).

## --env-file e --env / -e

- **`--env-file <path>`** — Define um arquivo de variáveis de ambiente (formato `.env`) que será carregado e mesclado ao ambiente antes de executar um plugin. Útil para manter configurações em um arquivo separado.
- **`--env KEY=VALUE`** — Injeta uma variável no processo do plugin. Pode ser repetido várias vezes. Tem a maior precedência em relação aos outros meios de definir variáveis.

Para a ordem completa de precedência e como usar defaults com `mb self env`, veja [Variáveis de ambiente](./environment-variables.md).

## --env-group

**O que faz:** Ao executar plugins, depois de carregar `~/.config/mb/env.defaults`, mescla por cima o arquivo `~/.config/mb/.env.<nome>` (valores do grupo sobrescrevem chaves iguais do default).

**Quando usar:** Para alternar entre ambientes (ex.: `staging`, `prod`) sem trocar o conteúdo de `env.defaults`.

```bash
mb --env-group staging tools deploy
```

O nome do grupo segue as mesmas regras que em `mb self env set --group`. Detalhes em [Variáveis de ambiente](./environment-variables.md).

## --doc

**O que faz:** Abre no navegador a URL de documentação configurada (por omissão o site público do projeto). O URL base define-se em **`~/.config/mb/config.yaml`** com a chave **`docs_url`**. Encerra o CLI com código `0`.

**Uso:** Só no comando raiz, **antes** de qualquer subcomando — por exemplo `mb --doc`. Não é herdada por `mb plugins list` etc.

Detalhes e exemplos: [Configuração do CLI (config.yaml)](../technical-reference/cli-config.md).

```bash
mb --doc
```
