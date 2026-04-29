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
│   ├── cmd/            # Comandos cobra (criar, validar, servidor iniciar/parar/status)
│   ├── internal/
│   │   ├── assinador/  # ClienteCLI (os/exec), ClienteHTTP (net/http), Garantir startup, LocalizarJar
│   │   ├── porta/      # Auto-deteccao de porta livre (+20 janela)
│   │   ├── logging/    # slog + OTel bridge
│   │   ├── jdk/        # (planejado Sprint 4) deteccao e provisionamento JDK
│   │   └── state/      # Leitura/escrita de ~/.hubsaude/
│   └── main.go
├── simulador/          # CLI Go
│   ├── cmd/            # Comandos cobra (iniciar, parar, status)
│   ├── internal/
│   │   ├── logging/    # slog + OTel bridge
│   │   ├── state/      # Leitura/escrita de ~/.hubsaude/ (duplicado)
│   │   ├── download/   # (planejado Sprint 4) Download do simulador.jar
│   │   ├── jdk/        # (planejado Sprint 4) Deteccao e provisionamento JDK
│   │   └── processo/   # (planejado Sprint 4) Gerenciamento do processo Java
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

**Fase:** Sprint 3 parcialmente concluida (3.1–3.5). Itens 3.6–3.9 pendentes.

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

### Sprint 3 — Em andamento (14/04 — 27/04)
**Concluido (3.1–3.5):**
- `SignatureController` Javalin + `AssinadorServidor` (`POST /sign`, `POST /validate`, `GET /health`, `POST /shutdown`)
  — `App.java` aceita `server` como primeiro arg; default continua sendo modo CLI
- Pacote Go `internal/assinador` com `ClienteCLI` (modo local via `os/exec`) e `ClienteHTTP` (modo HTTP)
- `Garantir()` em `startup.go`: checa `state.json` + PID + `/health`, reusa instancia existente ou inicia nova
- Pacote Go `internal/porta`: `LivreOuProxima` em janela de +20 portas
- Wiring nos comandos cobra: `criar`, `validar`, `servidor iniciar/parar/status`
- Localizacao do jar via `LocalizarJar()` (env `HUBSAUDE_ASSINADOR_JAR` → `~/.hubsaude/assinador/` → cwd → `assinador/target/`)

**Pendente (3.6–3.9):**
- Limpeza fina de `state.json` ao parar servidor
- Formatacao explicita de stdout/stderr (saida ja usa JSON puro)
- Propagacao estruturada de erros entre camadas (mapear `*RespostaErro` → exit code do CLI)
- Testes de integracao end-to-end (CLI ↔ jar real, dois modos)

### Sprint 4 — proximas acoes
- Provisionamento JDK (Adoptium API)
- Comandos `simulador iniciar/parar/status` integrados a download do `simulador.jar`
- Cache + flag `--source` + verificacao SHA256

### Status de testes (validado localmente)
- **Go:** 37 testes em `assinatura` + 17 em `simulador` = **54 passando**
- **Java:** 69 testes em `assinador` (cobertura > 80%)
- **Total: 123 testes passando**
- Tooling local: Go 1.26 (Homebrew), Java 21, Maven 3.9 via `mvnw`

---

## Cobertura de Testes

**CLI Go (54 testes — `testing` + `testify`):**
- `assinatura/cmd`: registro de subcomandos e flags (criar, validar, servidor)
- `assinatura/internal/assinador`: ClienteCLI (stdin/stdout, RespostaErro estruturada), ClienteHTTP (sucesso, erro 4xx, /health, AguardarPronto), Garantir startup (reuso, PID obsoleto)
- `assinatura/internal/porta`: Disponivel, EmUso, LivreOuProxima (range +20)
- `assinatura/internal/state`: leitura/escrita state.json + config.json, `t.TempDir()`, CleanStale
- `simulador/cmd` + `simulador/internal/state`: idem ao assinatura

**assinador.jar (69 testes — JUnit 5):**
- `FakeSignatureService` / `PKCS11SignatureService` / `SignatureService`
- `ValidadorFHIR`: campos obrigatorios, base64, algoritmo, hashes SHA-256
- `AssinadorCli` + `AcaoAssinar` + `AcaoValidar`: parsing args, payload via stdin/arquivo
- `MapeadorErro` + `RespostaErro`
- `SignatureController`: /sign, /validate, /health, /shutdown via HTTP real
- Cobertura JaCoCo > 80% (excluindo `PKCS11SignatureService`)

**Lacunas (a cobrir nas proximas sprints):**
- Testes de integracao end-to-end CLI ↔ jar real (Sprint 3, item 3.9)
- Download do JDK e simulador.jar (Sprint 4, mock de HTTP)
- Provisionamento JDK multiplataforma (Sprint 4)

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

Empacotamento `.AppImage` (Linux) e `.dmg` (macOS) esta planejado para Sprint 5.
Ver `planejamento/ci-cd.md` para os workflows prontos.

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
