---
sidebar_position: 5
---

# Variáveis de ambiente

O MB CLI controla o ambiente em que os plugins são executados. As variáveis são mescladas em uma ordem bem definida e só o processo do plugin recebe esse ambiente final; o CLI em si não altera o ambiente do seu shell.

## Ordem de precedência

Da **menor** para a **maior** precedência:

1. **Variáveis do sistema** — O que já está em `os.Environ()` (incluindo o que você exportou no shell).
2. **Arquivo de defaults** — `~/.config/mb/env.defaults` (ou equivalente no macOS). Opcionalmente, o arquivo passado em **`--env-file <path>`** é mesclado (e sobrescreve os defaults quando há conflito de chave).
3. **Linha de comando** — **`--env KEY=VALUE`** (pode ser repetido). Qualquer valor definido aqui sobrescreve os anteriores.

Ou seja: **`--env`** tem a última palavra para cada chave.

## Como usar

### Defaults persistentes: `mb self env`

Você pode definir variáveis que serão usadas em toda execução de plugins, sem precisar passar `--env` toda vez:

- **`mb self env list`** — Lista as variáveis definidas nos defaults.
- **`mb self env set KEY [VALUE]`** — Define uma variável nos defaults. Se VALUE não for informado, o valor pode ser lido de forma interativa (dependendo do ambiente).
- **`mb self env unset KEY`** — Remove a variável dos defaults.

Esses comandos alteram o arquivo `env.defaults` no diretório de configuração do MB.

### Arquivo de ambiente: `--env-file`

Para usar um arquivo `.env` em um caminho específico (por exemplo, por projeto):

```bash
mb --env-file .env tools meu-comando
```

O conteúdo do arquivo é mesclado ao ambiente antes de rodar o plugin.

### Variáveis na linha de comando: `--env`

Para sobrescrever ou definir variáveis em uma única execução:

```bash
mb --env API_KEY=xyz --env AMBIENTE=prod tools meu-comando
```

## Exemplo prático

No seu plugin, você pode acessar variáveis injetadas normalmente. Por exemplo, em um script `run.sh`:

```bash
#!/bin/sh
echo "API_KEY está definida? ${API_KEY:-não}"
```

Se você definiu `API_KEY` com `mb self env set API_KEY` ou com `--env API_KEY=abc`, o plugin verá o valor ao ser executado.

Para detalhes de implementação (onde no código o merge é feito e como é passado ao processo do plugin), veja a [Referência técnica](./plugins.md).
