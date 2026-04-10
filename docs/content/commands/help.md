---
sidebar_position: 6
---

# `mb help`

Ajuda sobre qualquer comando do MB CLI.

## Uso básico

```bash
mb help                  # ajuda geral (igual a mb --help)
mb help plugins          # ajuda do subcomando plugins
mb help plugins add      # ajuda do subcomando add
mb help envs list        # ajuda do subcomando list
```

## Equivalências

| Forma | Resultado |
|---|---|
| `mb help <cmd>` | Ajuda formatada do comando |
| `mb <cmd> --help` | Ajuda inline do comando |
| `mb help <cmd1> <cmd2>` | Ajuda do subcomando aninhado |

> **Nota:** Para `mb run`, prefira `mb help run` em vez de `mb run --help`, pois o `--help` pode ser repassado ao programa filho.

## Flag `-h` / `--help`

O mesmo efeito pode ser obtido com a flag `-h` ou `--help` em qualquer comando:

```bash
mb -h
mb plugins -h
mb plugins add --help
```

A flag `-h` está disponível em **todos os comandos e subcomandos**, incluindo os gerados dinamicamente a partir de plugins instalados.

## O que a ajuda mostra

Para cada comando, o help exibe:

- **USAGE** — sintaxe do comando com argumentos obrigatórios e opcionais
- **COMANDOS** / **PLUGINS** — subcomandos disponíveis (agrupados por categoria)
- **FLAGS** — flags disponíveis com descrição, valor por defeito e atalhos (`-e`, `-J`, etc.)
- **FLAGS GLOBAIS** — flags herdadas do comando raiz (`--env`, `--env-vault`, `-v`, `-q`)
- **EXAMPLES** — exemplos de uso (quando o plugin os declara no manifest)

## Plugins na ajuda

Comandos de plugins instalados aparecem na secção **PLUGINS** do help raiz e nos helps de categoria:

```bash
mb help
mb help tools
mb help tools meu-comando
```

Plugins locais exibem a indicação **(local)** ao lado da descrição.

## Documentação no navegador

Para abrir a documentação online no navegador:

```bash
mb --doc
```

O URL base configura-se em `config.yaml` no diretório de configuração do MB (ex.: `~/.config/mb/config.yaml` no Linux) com a chave `docs_url`.
