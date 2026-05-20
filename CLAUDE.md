# CLAUDE.md — Sistema Runner

Trabalho pratico da disciplina Implementacao e Integracao — Engenharia de Software, UFG (2026-01).
Periodo: 17/03/2026 a 16/06/2026 (6 sprints de 2 semanas).

---

## Visao Geral

Tres componentes integrados:

| Componente | Tipo | Descricao |
|------------|------|-----------|
| `assinatura` | CLI Go | Invoca o assinador.jar (modo local ou HTTP) |
| `simulador` | CLI Go | Gerencia o ciclo de vida do simulador.jar do HubSaude |
| `assinador.jar` | Java | Valida parametros FHIR e simula assinatura digital |

O `simulador.jar` (HubSaude) **nao e desenvolvido** neste projeto — e baixado dinamicamente via GitHub Releases da disciplina.

---

## Stack Tecnologica

### CLIs (assinatura + simulador) — binarios separados
- **Linguagem:** Go
- **CLI parsing:** `cobra`
- **HTTP client:** `net/http` (stdlib)
- **Execucao de processos:** `os/exec` (stdlib)
- **Testes:** `testing` (stdlib) + `testify`
- **Release:** `goreleaser`

### assinador.jar
- **Linguagem:** Java 21 LTS (Temurin)
- **Build:** Maven com `mvnw`
- **Framework HTTP:** Javalin
- **Testes:** JUnit 5 + Mockito
- **Formato de I/O:** JSON (`stdout`/`stderr` no modo CLI; `application/json` no modo HTTP)

### CI/CD
- **Plataforma:** GitHub Actions
- **Assinatura de artefatos:** Cosign (`sigstore/cosign-installer@v3`, OIDC)
- **Releases:** `softprops/action-gh-release@v2`, acionado por tag `v*`

---

## Arquitetura

```
assinatura (CLI Go)
  ├── modo local:  os/exec -> java -jar assinador.jar sign/validate
  └── modo HTTP:   net/http -> POST localhost:8088/sign | /validate

simulador (CLI Go)
  └── os/exec -> java -jar ~/.hubsaude/simulador/simulador.jar

Estado local: ~/.hubsaude/
  ├── jdk/temurin-21.x/     # JDK provisionado automaticamente
  ├── simulador/             # simulador.jar + simulador-meta.json
  ├── state.json             # PID + porta dos processos em execucao
  └── config.json            # Configuracoes do usuario
```

### Endpoints do assinador.jar (modo servidor, porta 8088)
| Endpoint | Metodo | Descricao |
|----------|--------|-----------|
| `/sign` | POST | Criar assinatura (simulada) |
| `/validate` | POST | Validar assinatura |
| `/health` | GET | Health check |
| `/shutdown` | POST | Desligar servidor |

Se a porta 8088 estiver ocupada, o CLI auto-detecta a proxima livre (8089, 8090... +20) e grava em `state.json`.

### Design Java (assinador.jar)
```
SignatureService (interface)
  ├── FakeSignatureService       # Simulacao — implementacao principal entregue
  └── PKCS11SignatureService     # Esqueleto — dispositivo fisico nao disponivel

SignatureController              # Javalin — expoe /sign e /validate
```

**PKCS#11:** SunPKCS11 (JDK built-in) e a ponte entre Java e o driver nativo do token/smartcard. A chave privada **nao e acessivel como parametro** no modo PKCS#11 — fica no dispositivo. `FakeSignatureService` implementa `sign(message, privateKey)` normalmente; `PKCS11SignatureService` autentica via PIN e KeyStore.

---

## Estrutura de Pastas

```
runner/
├── assinatura/         # CLI Go
│   ├── cmd/            # Comandos cobra (criar, validar, servidor iniciar/parar/status); exit.go mapeia codigos
│   ├── internal/
│   │   ├── assinador/  # ClienteCLI (os/exec), ClienteHTTP (net/http), Garantir startup, LocalizarJar, integration_test (tag)
│   │   ├── porta/      # Auto-deteccao de porta livre (+20 janela)
│   │   ├── logging/    # slog + OTel bridge
│   │   ├── jdk/        # Deteccao (JAVA_HOME/PATH/hubsaude) + download Adoptium (Sprint 4)
│   │   └── state/      # Leitura/escrita de ~/.hubsaude/
│   └── main.go
├── simulador/          # CLI Go
│   ├── cmd/            # Comandos cobra (iniciar, parar, status); exit.go
│   ├── internal/
│   │   ├── logging/    # slog + OTel bridge
│   │   ├── state/      # Leitura/escrita de ~/.hubsaude/ (duplicado)
│   │   ├── porta/      # Auto-deteccao (duplicado de assinatura)
│   │   ├── download/   # Download do simulador.jar via GitHub Releases + cache + --source + SHA-256
│   │   ├── jdk/        # Deteccao e provisionamento JDK (duplicado de assinatura)
│   │   └── processo/   # Iniciar/Parar/Consultar (HTTP shutdown -> SIGTERM -> Kill)
│   └── main.go
├── assinador/          # Java — Maven
│   └── src/main/java/br/gov/saude/assinador/
│       ├── servico/    # SignatureService, Fake/PKCS11SignatureService, AssinadorException
│       ├── servidor/   # SignatureController + AssinadorServidor (Javalin)
│       ├── cli/        # Modo CLI (AssinadorCli, AcaoAssinar, AcaoValidar)
│       ├── validacao/  # ValidadorFHIR, AlgoritmoSuportado
│       └── erro/       # MapeadorErro (HTTP/exit code), RespostaErro
├── .github/workflows/
│   └── ci.yml          # Build + testes (matrix 3 SOs)
├── planejamento/
└── docs/
```

---

## Convencoes de Codigo

### Go
- Pacotes: minusculos, sem underscores (`state`, `processo`, `download`)
- Erros sempre propagados com contexto: `fmt.Errorf("ao iniciar servidor: %w", err)`
- Interfaces definidas no pacote consumidor, nao no produtor
- Testes em `_test.go`; usar `t.TempDir()` para operacoes de filesystem
- Logging: `stdout` → resultado final (JSON); `stderr` → progresso e erros operacionais

### Java
- Pacote base: `br.gov.saude.assinador`
- `SignatureService` como interface central; implementacoes injetadas (sem `new` direto nos controllers)
- Erros retornam JSON: `{ "error": "<CODIGO>", "message": "<descricao legivel>" }`
- Testes: `@ExtendWith(MockitoExtension.class)`

---

## Decisoes Tomadas

| Decisao | Escolha |
|---------|---------|
| Linguagem CLI | Go |
| Build Java | Maven + `mvnw` |
| Framework HTTP Java | Javalin |
| JDK | 21 LTS — Eclipse Temurin |
| JDK download | Adoptium API (`api.adoptium.net/v3`) |
| Formato I/O | JSON |
| Porta padrao assinador | 8088 (auto-detecta se ocupada) |
| Estado local | `~/.hubsaude/` |
| CLIs | Dois binarios separados |
| Modulo JDK | Duplicado em cada CLI |
| Logging | `stdout` resultado; `stderr` progresso/erros; `--verbose`/`--quiet` |
| CI/CD | GitHub Actions — `ci.yml` + `release.yml` |
| Assinatura artefatos | Cosign OIDC (obrigatorio pela spec secao 9.4) |

---

## Status Atual

**Fase:** Sprints 3, 4 e 5 concluidas. Sprint 6 em andamento (docs 6.1-6.4 concluidas; faltam 6.5-6.8).

### Sprint 1 — Concluida (2026-03-31)
- CLI `assinatura` e `simulador` com cobra (subcomandos esqueleto)
- Projeto Java `assinador` (Maven + mvnw, JUnit 5, Javalin)
- `~/.hubsaude/` com `state.json` + `config.json` (PID check via `windows.OpenProcess` / `Signal(0)`)
- Pipeline CI matrix 3 SOs
- Logging OTel (slog multiHandler em Go; Logstash + OTel appender em Java)

### Sprint 2 — Concluida (2026-04-13)
- Interface `SignatureService` + `FakeSignatureService` + esqueleto `PKCS11SignatureService`
- Validacao de parametros FHIR (`ValidadorFHIR`) com codigos `PARAM_AUSENTE`/`PARAM_INVALIDO`/`ALGORITMO_NAO_SUPORTADO`
- Modo CLI do `assinador.jar` (`AssinadorCli`, `AcaoAssinar`, `AcaoValidar`) — payload JSON via `--input <arq>` ou stdin
- Mapeador de erros (`MapeadorErro`) → HTTP status + exit code
- Cobertura > 80% (excluindo `PKCS11SignatureService`)

### Sprint 3 — Concluida (2026-04-27)
- `SignatureController` Javalin + `AssinadorServidor` (`POST /sign`, `POST /validate`, `GET /health`, `POST /shutdown`)
  — `App.java` aceita `server` como primeiro arg; default continua sendo modo CLI
- Pacote Go `internal/assinador` com `ClienteCLI` (modo local via `os/exec`) e `ClienteHTTP` (modo HTTP)
- `Garantir()` em `startup.go`: checa `state.json` + PID + `/health`, reusa instancia existente ou inicia nova
- Pacote Go `internal/porta`: `LivreOuProxima` em janela de +20 portas
- Wiring nos comandos cobra: `criar`, `validar`, `servidor iniciar/parar/status`
- Localizacao do jar via `LocalizarJar()` (env `HUBSAUDE_ASSINADOR_JAR` → `~/.hubsaude/assinador/` → cwd → `assinador/target/`)
- **3.6** Comando `servidor parar`: tenta `/shutdown` HTTP, fallback `Kill()`, limpa `state.json`, reporta `metodo` (http_shutdown | sigkill | stale_pid)
- **3.7** Saida JSON pretty-printed em stdout (`enc.SetIndent("", "  ")`); progresso/erros via stderr; flag `--quiet` redireciona progresso para `io.Discard`
- **3.8** Pacote `cmd/exit.go`: mapeia `*assinador.RespostaErro.Codigo` -> exit codes (`PARAM_AUSENTE=2`, `PARAM_INVALIDO=3`, `ALGORITMO_NAO_SUPORTADO=4`, `ASSINATURA_INVALIDA=5`, `DISPOSITIVO_INDISPONIVEL/PIN_INVALIDO=6`, `ERRO_INTERNO=7`, generico=1). `ExecuteOrExit()` emite erro JSON estruturado em stderr antes de `os.Exit`.
- **3.9** Testes de integracao em `internal/assinador/integration_test.go` (build tag `integration`): sobem o jar real, validam modo local + HTTP + reuso de instancia. 5 testes passando com `go test -tags=integration`.

### Sprint 4 — Concluida (2026-05-11)
- **4.1 + 4.2** Pacote `internal/jdk/` (duplicado em assinatura e simulador):
  - `Detectar()` busca em ordem: `JAVA_HOME` -> `java` no PATH -> `~/.hubsaude/jdk/*/bin/java`
  - `parseMajor()` aceita `21.0.3`, `1.8.0_421` (legado), `22-ea`
  - `Garantir()` baixa via Adoptium API quando ausente; suporta `linux/mac/windows` x `amd64/arm64`; extrai `.tar.gz` (Unix) ou `.zip` (Windows); valida SHA-256; persiste `hubsaude-meta.json`. Sobrescritivel em testes via `HUBSAUDE_ADOPTIUM_URL`.
- **4.3 + 4.6** `processo.Iniciar()` (em `simulador/internal/processo`): tenta reusar via `state.json` + PID + health (`/actuator/health` ou `/health`); auto-detecta porta a partir do default 9090 (config) com janela +20; usa `--server.port=N` no jar; aguarda `aguardarPorta()` antes de gravar `state.json`.
- **4.4 + 4.5** Pacote `simulador/internal/download/`:
  - Cache em `~/.hubsaude/simulador/{simulador.jar, simulador-meta.json}`
  - GitHub Releases via `https://api.github.com/repos/hubsaude/simulador/releases` (sobrescritivel via `HUBSAUDE_SIMULADOR_REPO`)
  - Versao `latest` ou tag especifica via `--versao-simulador`
  - Flag `--source <url>` aceita `http(s)://` e `file://` (versao = "custom", ignora GitHub Releases)
  - Re-valida SHA-256 do cache contra metadados a cada chamada (deteccao de corrupcao)
  - `ForcarRedownload` para ignorar cache
- **4.7** `processo.Parar()`: tenta `/actuator/shutdown` -> `/shutdown` HTTP; se nao parar em 3s, envia SIGTERM (Unix); se nao parar em 5s, `Kill()`. Limpa `state.json` ao final. Reporta `metodo` (`http_shutdown` | `sigterm` | `sigkill` | `stale_pid` | `nao_estava_rodando`).
- **4.8** `processo.Consultar()` + comando `status`: retorna `{registrado, running, pid, porta, iniciadoEm, pidVivo, versao, uptimeSegundos}`.
- **4.9** Cobertura completa: 8 testes em `jdk/` (parseMajor, deteccao, download Adoptium com httptest + tar.gz fake + checksum divergente), 10 testes em `download/` (cache hit, versao especifica, --source http/file, --force, tag inexistente, sem assets), 9 testes em `processo/` (reuso, pid obsoleto, status, parar nao-rodando, parar stale, escolha de porta).

### Sprint 5 — Concluida (2026-05-19)
- **5.1** `.goreleaser.yaml` na raiz cross-compila ambos os CLIs para `linux/windows/darwin x amd64`. Multi-modulo via `dir:` por build; ldflags injetam `Version` e `GitCommit` em `cmd.Version`/`cmd.GitCommit`. Nome final: `<bin>-<versao>-<so>-<arch>[.exe]` (darwin renomeado para `macos`).
- **5.2** Scripts em `scripts/`:
  - `build-appimage.sh` — empacota binario Linux como `.AppImage` (cria AppDir, AppRun, .desktop, icone PNG placeholder; usa `appimagetool` extraido sem FUSE)
  - `build-dmg.sh` — empacota binario macOS como `.dmg` via `hdiutil create -format UDZO -fs HFS+`
  - `rename-windows.sh` — renomeia binario Windows para o padrao da spec (.exe ja produzido pelo Go)
- **5.3** `release.yml` constroi `assinador.jar` via Maven e renomeia para `assinador-<versao>.jar`.
- **5.4** Checksums SHA-256 agregados em `checksums-sha256.txt` (gerado pelo `sha256sum * | sort` no job `release`).
- **5.5** Cosign keyless OIDC via `sigstore/cosign-installer@v3` — gera `<artefato>.sig` + `<artefato>.pem` para cada item (exceto `.sig`/`.pem`/checksums). Permissoes `id-token: write` + `contents: write` no job `release`.
- **5.6** `release.yml` acionado por tag `v*` (versao derivada de `${GITHUB_REF_NAME#v}`). `workflow_dispatch` permite dry-run sem publicar. `softprops/action-gh-release@v2` publica com release notes auto-geradas e instrucoes Cosign no body.
- **5.7** Testes de aceitacao opt-in (build tag `acceptance`):
  - `assinatura/cmd/acceptance_test.go` — 8 testes que constroem o binario via `TestMain` e validam US-01 + US-02 (help, versao, criar/validar modo local com jar real, payload invalido com codigo de saida estruturado, servidor status/parar, payload inexistente)
  - `simulador/cmd/acceptance_test.go` — 5 testes que validam US-03 (help, versao, status sem nada rodando, parar sem nada rodando, flag `--source` aceita)
  - US-04 (JDK) ja coberta pelos testes unitarios profundos em `internal/jdk/`
  - US-05 validada pela execucao end-to-end do `release.yml`
- **5.8** Job `acceptance` no `ci.yml` em matrix 3 SOs (Ubuntu/Windows/macOS): builda `assinador.jar`, exporta `HUBSAUDE_ASSINADOR_JAR`, roda `go test -tags=acceptance ./cmd -run Aceitacao`. Job adicional `release-config` valida o `.goreleaser.yaml` com `goreleaser check`.

### Sprint 6 — em andamento
- **6.1** Manual do usuario (`docs/manual-usuario.md`): conceitos local vs http, flags globais, referencia de `assinatura criar/validar/servidor` e `simulador iniciar/parar/status`, exit codes, troubleshooting, layout de `~/.hubsaude/`.
- **6.2** Documentacao tecnica (`docs/tecnico.md`): componentes, fluxo de dados, contrato HTTP (`/sign`, `/validate`, `/health`, `/shutdown`) + contrato CLI do jar, design `SignatureService`/PKCS#11, startup inteligente, provisionamento JDK, download do simulador, estado local, distribuicao/Cosign.
- **6.3** Guia de instalacao (`docs/instalacao.md`): download por plataforma, checksums, verificacao Cosign (com `--certificate-identity-regexp`/`--certificate-oidc-issuer`), instalacao Windows/Linux/macOS, modo offline, verificacao pos-instalacao.
- **6.4** Exemplos (`docs/exemplos.md`) com payloads validos completos; README corrigido (flags reais `--payload`/`--modo`/`--porta`/`--jar`; antes documentava `--message-file`/`--private-key` inexistentes) + links para `docs/`.
- **Pendentes:** 6.5 (cobertura), 6.6 (revisao de erros/mensagens), 6.7 (bugs), 6.8 (publicacao final via tag `v1.0.0`).
- **Lacuna conhecida:** divergencia de enums de erro entre Go (`cmd/exit.go`: codigos 5/6 `ASSINATURA_INVALIDA`/`DISPOSITIVO_INDISPONIVEL`) e Java (`AssinadorException.Codigo`: `PAYLOAD_MUITO_GRANDE`); 5/6 documentados como reservados. Avaliar alinhamento em 6.6.

### Status de testes (validado localmente)
- **Go:** 50 testes em `assinatura` + 49 em `simulador` = **99 passando**
- **Java:** 69 testes em `assinador` (cobertura > 80%)
- **Integracao (opt-in, `-tags=integration`):** 5 testes CLI ↔ jar real
- **Aceitacao (opt-in, `-tags=acceptance`):** 8 em `assinatura` + 5 em `simulador` = **13 testes** validando US-01..US-03 end-to-end via binario
- **Total: 168 testes passando** (173 com integracao, 186 com aceitacao)
- Tooling local: Go 1.26 (Homebrew), Java 21, Maven 3.9 via `mvnw`

---

## Cobertura de Testes

**CLI Go (99 testes — `testing` + `testify`):**
- `assinatura/cmd`: registro de subcomandos e flags + mapeamento exit codes (`exit_test.go`)
- `assinatura/internal/assinador`: ClienteCLI (stdin/stdout, RespostaErro estruturada), ClienteHTTP (sucesso, erro 4xx, /health, AguardarPronto), Garantir startup (reuso, PID obsoleto)
- `assinatura/internal/jdk`: parseMajor, Detectar (PATH/JAVA_HOME/local), Garantir Adoptium com httptest + tar.gz fake, checksum divergente
- `assinatura/internal/porta`: Disponivel, EmUso, LivreOuProxima (range +20)
- `assinatura/internal/state`: leitura/escrita state.json + config.json, `t.TempDir()`, CleanStale
- `simulador/cmd`: registro de comandos + flags
- `simulador/internal/jdk`: idem assinatura (duplicado)
- `simulador/internal/porta`: idem assinatura (duplicado)
- `simulador/internal/download`: cache hit/miss, --source http/file, versao especifica, --force, SHA-256 esperado, tag inexistente
- `simulador/internal/processo`: Iniciar (reuso, PID obsoleto, jar obrigatorio), Parar (nao-rodando, stale), Consultar, escolherPorta
- `simulador/internal/state`: idem assinatura

**Testes de integracao (opt-in, `-tags=integration`):**
- `assinatura/internal/assinador/integration_test.go`: sobe `assinador.jar` real (5 testes)
  - Modo local: sign/validate sucesso + payload invalido
  - Modo HTTP: sign via Garantir + reuso de instancia

**Testes de aceitacao (opt-in, `-tags=acceptance`):**
- `assinatura/cmd/acceptance_test.go` (8 testes): constroi binario via `TestMain`, exec-uta e valida US-01/02 — help lista subcomandos, `versao` imprime versao injetada via ldflags, `criar`/`validar` modo local com jar real, payload invalido retorna codigo de saida estruturado + erro em stderr, `servidor status`/`parar` sem nada rodando, payload inexistente.
- `simulador/cmd/acceptance_test.go` (5 testes): mesmo padrao, valida US-03 — help, `versao`, `status`/`parar` sem nada rodando, `--source` com URL custom (verifica que flag e parseada antes de erro de download).

**assinador.jar (69 testes — JUnit 5):**
- `FakeSignatureService` / `PKCS11SignatureService` / `SignatureService`
- `ValidadorFHIR`: campos obrigatorios, base64, algoritmo, hashes SHA-256
- `AssinadorCli` + `AcaoAssinar` + `AcaoValidar`: parsing args, payload via stdin/arquivo
- `MapeadorErro` + `RespostaErro`
- `SignatureController`: /sign, /validate, /health, /shutdown via HTTP real
- Cobertura JaCoCo > 80% (excluindo `PKCS11SignatureService`)

**Lacunas (a cobrir na Sprint 6):**
- Manual do usuario (`assinatura` + `simulador`) e guia de instalacao com Cosign verify
- README + exemplos de uso
- Documentacao tecnica de integracao + PKCS#11
- Publicacao final via tag `v1.0.0` (executa o `release.yml` end-to-end)

---

## Artefatos da Release

Formato exigido pela especificacao para cada versao (ex: `v1.0.0`):

```
assinatura-1.0.0-linux-amd64.AppImage   + .sig + .pem
assinatura-1.0.0-windows-amd64.exe      + .sig + .pem
assinatura-1.0.0-macos-amd64.dmg        + .sig + .pem
simulador-1.0.0-linux-amd64.AppImage    + .sig + .pem
simulador-1.0.0-windows-amd64.exe       + .sig + .pem
simulador-1.0.0-macos-amd64.dmg         + .sig + .pem
assinador-1.0.0.jar
checksums-sha256.txt
```

Empacotamento, assinatura e checksums sao gerados automaticamente pelo
`.github/workflows/release.yml` quando uma tag `v*` e enviada (ver scripts
auxiliares em `scripts/build-appimage.sh`, `scripts/build-dmg.sh`,
`scripts/rename-windows.sh` e o `.goreleaser.yaml` na raiz).

---

## Planejamento Detalhado

`planejamento/` — nao repetir aqui, consultar quando necessario:

| Arquivo | Quando consultar |
|---------|-----------------|
| `decisoes-tecnicas.md` | Justificativas das decisoes tomadas |
| `arquitetura.md` | Fluxos de comunicacao e formatos de dados |
| `estado-local.md` | Estrutura de `~/.hubsaude/` |
| `startup.md` | Sequencia de inicializacao dos CLIs |
| `entregavel-assinador.md` | SignatureService, PKCS#11, contrato da API |
| `entregavel-cli-assinatura.md` | Comandos e flags do CLI assinatura |
| `entregavel-simulador.md` | `--source`, auto-porta, download/cache |
| `entregavel-jdk.md` | Adoptium API, fluxo de download |
| `entregavel-distribuicao.md` | Cross-compile, Cosign, goreleaser |
| `entregavel-testes.md` | Criterios de aceitacao por US |
| `ci-cd.md` | Workflows GitHub Actions prontos para uso |
| `contrato-fhir.md` | Parametros FHIR investigados para /sign e /validate |
