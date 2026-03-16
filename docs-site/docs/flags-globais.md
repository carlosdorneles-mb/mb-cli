---
sidebar_position: 4
---

# Flags globais

O MB CLI oferece algumas flags que valem para qualquer comando e afetam o nível de saída e o ambiente. Elas podem ser usadas antes ou depois do subcomando.

## --verbose / -v

**O que faz:** Ativa saída mais verbosa. O CLI pode exibir mensagens adicionais de diagnóstico ou logs que ajudam a entender o que está acontecendo.

**Quando usar:** Para depurar um problema, acompanhar o fluxo de um comando ou quando você quer mais detalhes na saída.

Exemplo:

```bash
mb -v plugins list
mb --verbose self sync
```

## --quiet / -q

**O que faz:** Reduz ou suprime mensagens informativas. O CLI evita imprimir avisos ou mensagens de progresso que não sejam estritamente necessárias.

**Quando usar:** Em scripts ou quando você quer apenas o resultado (por exemplo, saída de um plugin) sem mensagens extras do próprio CLI.

Exemplo:

```bash
mb -q plugins list
mb --quiet self sync
```

## --env-file e --env

- **`--env-file <path>`** — Define um arquivo de variáveis de ambiente (formato `.env`) que será carregado e mesclado ao ambiente antes de executar um plugin. Útil para manter configurações em um arquivo separado.
- **`--env KEY=VALUE`** — Injeta uma variável no processo do plugin. Pode ser repetido várias vezes. Tem a maior precedência em relação aos outros meios de definir variáveis.

Para a ordem completa de precedência e como usar defaults com `mb self env`, veja [Variáveis de ambiente](./variaveis-ambiente.md).
