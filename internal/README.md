# Internal — Arquitetura do mb-cli

Este diretório contém todo o código interno da CLI. Nenhuma dependência externa deve importar pacotes de `internal/`.

## Visão Geral da Arquitetura

O projeto segue os princípios de **Clean Architecture** e **Hexagonal Architecture** (Ports & Adapters), com injeção de dependência via [`go.uber.org/fx`](https://pkg.go.dev/go.uber.org/fx).

```
internal/
├── bootstrap/          # Composição raiz — monta a FX App e o Root Command
├── cli/                # Camada de Apresentação (Cobra apenas)
│   ├── root/           #   Root command e flags globais (--verbose, --quiet, --env)
│   ├── envs/           #   Comandos: mb envs list, set, unset, groups, path
│   ├── plugins/        #   Comandos: mb plugins add, list, remove, update, sync
│   ├── plugincmd/      #   Comandos dinâmicos gerados por plugins instalados
│   ├── completion/     #   Autocomplete shell (install, uninstall, generate)
│   ├── run/            #   mb run <processo>
│   ├── update/         #   mb update (self-update e tools update)
│   └── runtimeflags/   #   Flags de runtime injetadas nos plugins
├── usecase/            # Casos de Uso / Regras de Aplicação
│   ├── addplugin/      #   Service para instalar plugins (remote/local)
│   ├── envs/           #   Lógica de coleta e merge de variáveis de ambiente
│   ├── plugins/        #   Sync, remove, update de plugins
│   └── update/         #   Orquestração de self-update e tools update
├── domain/             # Modelos de Domínio e Regras de Negócio Puras
│   └── plugin/         #   Plugin, PluginSource, Category, HelpGroup, scan rules
├── infra/              # Implementações Concretas (Adapters)
│   ├── browser/        #   Abrir URLs no navegador
│   ├── executor/       #   ScriptExecutor (roda plugins via bash/binário)
│   ├── fs/             #   Filesystem real (os.MkdirAll, os.Stat, ...)
│   ├── keyring/        #   SecretStore via go-keyring (macOS Keychain, libsecret)
│   ├── opcli/          #   Integração com 1Password CLI (env vars de secrets)
│   ├── plugins/        #   Git operations, scanner de manifest.yaml, hash de config
│   ├── selfupdate/     #   Self-update via GitHub Releases
│   ├── shellhelpers/   #   Scripts utilitários embarcados (instalação no shell)
│   └── sqlite/         #   Persistência SQLite (plugins, sources, categories, help groups)
├── ports/              # Interfaces (Contratos) — o centro da Clean Architecture
│   ├── exec.go         #   ScriptExecutor
│   ├── fs.go           #   Filesystem
│   ├── git.go          #   GitOperations
│   ├── layout.go       #   PluginLayoutValidator
│   ├── onepassword.go  #   OnePasswordEnv
│   ├── pluginstore.go  #   PluginCacheStore, PluginCLIStore, PluginScanner
│   ├── secret.go       #   SecretStore
│   └── sync.go         #   PluginSyncStore, ShellHelperInstaller
├── deps/               # Configuração de Runtime (paths, flags resolvidas)
├── module/             # Módulos Fx — wire de dependência
│   ├── cache/          #   SQLite store + lifecycle (close on shutdown)
│   ├── cli/            #   Root command (recebe deps + portas de infra)
│   ├── deps/           #   SecretStore, OnePasswordEnv, Dependencies
│   ├── executor/       #   ScriptExecutor (plugin runner)
│   ├── infra/          #   Implementações reais: OS, Git, ShellInstaller, LayoutValidator
│   ├── plugins/        #   PluginScanner
│   └── runtime/        #   Paths (ConfigDir, PluginsDir, CacheDBPath) + AppConfig
├── shared/             # Utilitários Transversais
│   ├── config/         #   Carregamento e validação de config.yaml
│   ├── env/            #   Merge de variáveis de ambiente (defaults + group + inline)
│   ├── envgroup/       #   Grupos de ambiente (.env.staging, .env.production)
│   ├── safepath/       #   Validação segura de paths (previne path traversal)
│   ├── system/         #   Logger (gum log), gum table, gum input, gum confirm
│   ├── ui/             #   Tema, banner, erros em PT, glamour theme
│   └── version/        #   Version injection via ldflags
└── fakes/              # Test Doubles — mocks para testes unitários
    └── ports.go        #   FakeFS, FakeGit, FakeLogger, FakeShellInstaller, ...
```

## Camadas e Responsabilidades

### 1. `cli/` — Apresentação (Cobra)

**Responsabilidade**: Parse de argumentos, bind de flags, chamada ao usecase, exibição do resultado.

**Regra**: O `RunE` de cada comando deve ter **no máximo 5-10 linhas**. Toda lógica de negócio fica no usecase.

```go
// Exemplo: thin controller
func (c *cobra.Command) RunE: func(cmd *cobra.Command, args []string) error {
    log := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, cmd.ErrOrStderr())
    return svc.Add(cmd.Context(), addplugin.Request{
        Source:   args[0],
        Package:  pkg,
        Tag:      tag,
        NoRemove: noRemove,
    }, log)
}
```

### 2. `usecase/` — Casos de Uso

**Responsabilidade**: Orquestrar entidades e portas para executar uma regra de negócio.

**Regra**: Depende apenas de **interfaces** (`ports/`), nunca de implementações concretas.

```go
type Service struct {
    rt      Runtime
    store   ports.PluginCacheStore   // interface
    scanner ports.PluginScanner     // interface
    fsys    ports.Filesystem        // interface
    git     ports.GitOperations     // interface
    // ...
}

func (s *Service) Add(ctx context.Context, req Request, log Logger) error {
    // lógica de negócio — sem Cobra, sem os.MkdirAll, sem git clone direto
}
```

### 3. `domain/` — Modelos de Domínio

**Responsabilidade**: Entidades e regras de negócio puras. Sem dependências externas.

```go
type Plugin struct {
    CommandName  string
    CommandPath  string
    Description  string
    Entrypoint   string
    PluginDir    string
    ConfigHash   string
    GroupID      string
    // ...
}
```

### 4. `infra/` — Implementações Concretas (Adapters)

**Responsabilidade**: Implementar as interfaces definidas em `ports/` usando bibliotecas reais (SQLite, os/exec, git, etc.).

```go
// OS implement ports.Filesystem usando o sistema de arquivos real.
type OS struct{}

func (OS) RemoveAll(path string) error { return os.RemoveAll(path) }
func (OS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
```

### 5. `ports/` — Interfaces (Contratos)

**Responsabilidade**: Definir contratos que o usecase espera. São a fronteira entre domínio e infra.

```go
type Filesystem interface {
    RemoveAll(path string) error
    MkdirAll(path string, perm fs.FileMode) error
    Stat(name string) (fs.FileInfo, error)
    IsNotExist(err error) bool
    ReadDir(name string) ([]fs.DirEntry, error)
    Getwd() (string, error)
}
```

### 6. `module/` — Wire de Dependência (Fx)

**Responsabilidade**: Conectar interfaces às suas implementações usando `go.uber.org/fx`.

```go
// InfraModule fornece implementações reais para as portas de infraestrutura.
var InfraModule = fx.Module("infra",
    fx.Provide(
        func() ports.Filesystem { return mbfs.OS{} },
        func() ports.GitOperations { return plugins.GitService{} },
        func() ports.ShellHelperInstaller { return shellhelpers.Installer{} },
        func() ports.PluginLayoutValidator { return plugins.LayoutValidator{} },
    ),
)
```

### 7. `fakes/` — Test Doubles

**Responsabilidade**: Implementações de teste das interfaces de `ports/`, para uso em testes unitários.

```go
func TestAddPluginService_InvalidPath(t *testing.T) {
    fsys := fakes.NewFakeFS()
    git := fakes.NewFakeGit()
    logger := fakes.NewFakeLogger()
    // ...
    err := svc.Add(t.Context(), addplugin.Request{Source: "/nonexistent"}, logger)
    // assert error
}
```

## Fluxo de Injeção de Dependência

```text
main.go
  └── bootstrap.Bootstrap()
        └── fx.New(
              PathsModule        → resolve ConfigDir, PluginsDir, CacheDBPath
              CacheModule        → SQLite Store (com lifecycle close)
              PluginsModule      → PluginScanner
              ExecutorModule     → ScriptExecutor
              DepsModule         → SecretStore, OnePasswordEnv, Dependencies
              InfraModule        → OS, GitService, ShellInstaller, LayoutValidator
              CLIModule          → Root Command (recebe deps + portas de infra)
            )
        └── fx.Populate(&rootCmd)
```

O `NewRootCmd` recebe `deps.Dependencies` + as 4 interfaces de infra (`Filesystem`, `GitOperations`, `ShellHelperInstaller`, `PluginLayoutValidator`) e constrói internamente os services de usecase. O Fx resolve tudo automaticamente.

## Como Adicionar um Novo Usecase

1. **Criar o serviço** em `internal/usecase/<nome>/service.go`:

   ```go
   package meuusecase

   type Service struct {
       store ports.PluginCacheStore
       fsys  ports.Filesystem
       // apenas interfaces!
   }

   func New(store ports.PluginCacheStore, fsys ports.Filesystem) *Service {
       return &Service{store: store, fsys: fsys}
   }

   func (s *Service) Execute(ctx context.Context, req Request, log Logger) error {
       // lógica de negócio
   }
   ```

2. **Construir no `root/command.go`** — adicione as interfaces necessárias na assinatura do `NewRootCmd` e construa o serviço internamente:

   ```go
   func NewRootCmd(
       d deps.Dependencies,
       fsys ports.Filesystem,
       git  ports.GitOperations,
       // ... adicione novas interfaces aqui
   ) RootCommand {
       meuSvc := meuusecase.New(d.Store, fsys)
       // ...
   }
   ```

3. **Passar o serviço para o sub-comando**:

   ```go
   meuCmd := meupacote.NewCmd(meuSvc, d)
   rootCmd.AddCommand(meuCmd)
   ```

4. **Reduzir o `RunE`** do Cobra a 3-5 linhas chamando `svc.Execute(...)`.

5. **Criar testes** usando `internal/fakes/`:

   ```go
   func TestMeuUsecase_Cenario(t *testing.T) {
       fsys := fakes.NewFakeFS()
       store := &fakeStore{...}
       logger := fakes.NewFakeLogger()
       svc := meuusecase.New(store, fsys)
       err := svc.Execute(t.Context(), Request{...}, logger)
       // assertions
   }
   ```

## Diretrizes

| Princípio | Descrição |
|-----------|-----------|
| **Dependência aponta para dentro** | `cli` → `usecase` → `ports` ← `infra` |
| **Interfaces no centro** | `ports/` define contratos; `infra/` implementa |
| **Cobra é thin** | `RunE` parse args, chama usecase, exibe resultado |
| **Logger por chamada** | Logger é passado por parâmetro, não injetado no constructor |
| **Testes com fakes** | `internal/fakes/` para testes unitários; SQLite real para integração |
| **Sem variáveis globais** | Tudo injetado via Fx ou parâmetro |
| **Sem `init()` para DI** | Fx resolve o grafo de dependência automaticamente |
| **Sem módulos Fx por usecase** | Modules Fx são para infraestrutura; usecases são construídos no `root/command.go` |
