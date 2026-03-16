# Tools

Ferramentas e utilitários — categoria de exemplo do MB CLI.

## Comandos

### `do`

Comando que usa **entrypoint padrão** e **flags** no mesmo manifest:

- **`mb tools do`** — executa o script padrão (`run.sh`).
- **`mb tools do --deploy`** ou **`mb tools do -d`** — executa `deploy.sh`.
- **`mb tools do --rollback`** — executa `rollback.sh`.

Útil como referência para manifests com `entrypoint` e `flags` definidos. Veja o [manifest do comando do](do/manifest.yaml) e os scripts em `do/`.

## Uso

Com os plugins de exemplo registrados (`make install-examples` na raiz do repositório):

```bash
mb tools do
mb tools do --deploy
mb tools do -d
```

Para mais detalhes sobre como criar plugins e definir entrypoints e flags, veja a documentação em [Criar um plugin](../../../docs-site/docs/creating-plugins.md).
