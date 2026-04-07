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
│   ├── cmd/            # Comandos cobra (criar, validar, servidor)
│   ├── internal/
│   │   ├── assinador/  # Cliente CLI e HTTP para o assinador.jar
│   │   ├── jdk/        # Deteccao e provisionamento JDK
│   │   └── state/      # Leitura/escrita de ~/.hubsaude/
│   └── main.go
├── simulador/          # CLI Go
│   ├── cmd/            # Comandos cobra (iniciar, parar, status)
│   ├── internal/
│   │   ├── download/   # Download do simulador.jar
│   │   ├── jdk/        # Deteccao e provisionamento JDK (duplicado)
│   │   └── processo/   # Gerenciamento do processo Java
│   └── main.go
├── assinador/          # Java — Maven
│   └── src/main/java/br/gov/saude/assinador/
│       ├── servico/    # SignatureService, FakeSignatureService
│       ├── servidor/   # SignatureController (Javalin)
│       ├── cli/        # Modo CLI (args -> JSON stdout)
│       └── validacao/  # Validacao de parametros FHIR
├── .github/workflows/
│   ├── ci.yml          # Build + testes (matrix 3 SOs)
│   └── release.yml     # Cross-compile + Cosign + GitHub Release
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

**Fase:** Sprint 1 concluida. Sprint 2 proxima.

### Sprint 1 — Concluida (2026-03-31)
- [x] Parametros FHIR investigados e documentados (`planejamento/contrato-fhir.md`)
- [x] CLI `assinatura` com subcomandos `criar`, `validar`, `servidor` (cobra)
- [x] CLI `simulador` com subcomandos `iniciar`, `parar`, `status` (cobra)
- [x] Projeto Java `assinador` (Maven + mvnw, JUnit 5, Javalin, SignatureService interface)
- [x] `~/.hubsaude/` com `state.json` e `config.json` (leitura, escrita, PID check)
- [x] Pipeline CI (`.github/workflows/ci.yml`, matrix 3 SOs)
- [x] Logging estruturado OpenTelemetry (slog + OTel bridge em Go; SLF4J + Logback JSON + OTel appender em Java)
- [x] 25 testes passando (22 Go + 3 Java)

### Decisoes tomadas na Sprint 1
- Modulos Go independentes (sem workspace compartilhado)
- Pacote `state` duplicado em ambos CLIs (decisao intencional)
- PID check via `windows.OpenProcess` (Windows) / `Signal(0)` (Unix)
- OTel: multiHandler (JSON stderr + OTel bridge) em Go; Logstash encoder + OTel appender em Java

### Sprint 2 — proximas acoes
- Implementar `FakeSignatureService` (resposta simulada com formato FHIR)
- Implementar validacao de parametros FHIR (campos obrigatorios, base64, algoritmo)
- Implementar tratamento de erros estruturado (codigos FHIR)
- Implementar modo CLI do `assinador.jar` (args -> JSON stdout/stderr)
- Testes unitarios com cobertura > 80%

---

## Testes Pendentes

Nenhum teste implementado. Cenarios a cobrir:

**CLI Go (`testing` + `testify`):**
- Parsing de flags e subcomandos
- Auto-deteccao de porta (livre vs ocupada)
- Leitura/escrita de `state.json` (`t.TempDir()`)
- Validacao de PID obsoleto em `state.json`
- Invocacao do assinador.jar (mock de `os/exec`)
- Requisicao HTTP ao assinador (mock de `net/http`)
- Download do JDK e simulador.jar (mock de HTTP)

**assinador.jar (JUnit 5 + Mockito):**
- `FakeSignatureService.sign` → retorna valor base64 fixo
- `FakeSignatureService.validate` → assinatura correta = true; incorreta = false
- Validacao: campo obrigatorio ausente → `PARAM_AUSENTE` (400)
- Validacao: base64 invalido → `PARAM_INVALIDO` (400)
- Validacao: algoritmo nao suportado → `ALGORITMO_NAO_SUPORTADO` (400)
- `POST /sign` payload valido → 200 com assinatura
- `POST /validate` → 200 com `{ "valid": true/false }`
- `POST /sign` payload invalido → 400

**Meta:** 80% de cobertura de linhas (foco em `validacao/` e `servico/`).

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
