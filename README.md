# Sistema Runner

Trabalho prático da disciplina **Implementação e Integração** — Bacharelado em Engenharia de Software, UFG (2026-01).

O Sistema Runner facilita a execução de aplicações Java via linha de comandos, ocultando a complexidade de configuração e instalação do ambiente Java. O projeto é de interesse real da Secretaria de Estado de Saúde de Goiás (SES) e da Universidade Federal de Goiás (UFG), no contexto da plataforma **HubSaúde** de interoperabilidade de dados em saúde.

## Componentes

| Componente | Linguagem | Descrição |
|------------|-----------|-----------|
| **assinatura** | Go | CLI que invoca o `assinador.jar` para criar e validar assinaturas digitais |
| **assinador.jar** | Java 21 | Valida parâmetros FHIR e simula operações de assinatura digital |
| **simulador** | Go | CLI que gerencia o ciclo de vida do `simulador.jar` do HubSaúde |

O `simulador.jar` não é desenvolvido neste projeto — é obtido dinamicamente via GitHub Releases da disciplina.

## Arquitetura

```
Usuário
  ├── assinatura (CLI Go)
  │     ├── modo local:  java -jar assinador.jar sign/validate
  │     └── modo HTTP:   POST http://localhost:8088/sign | /validate
  │                              └── assinador.jar (servidor Javalin)
  │                                        └── SunPKCS11 -> driver PKCS#11
  │
  └── simulador (CLI Go)
        └── java -jar ~/.hubsaude/simulador/simulador.jar

Estado local: ~/.hubsaude/
  ├── jdk/          JDK 21 provisionado automaticamente (Temurin)
  ├── simulador/    simulador.jar + metadados de versão
  ├── state.json    PID e porta dos processos em execução
  └── config.json   Configurações do usuário
```

## Funcionalidades

- **Criar e validar assinaturas digitais** (simuladas) com validação rigorosa de parâmetros FHIR
- **Dois modos de invocação** do assinador: direto (`java -jar`) ou via HTTP (servidor persistente)
- **Suporte a dispositivo criptográfico** (token/smartcard PKCS#11) via SunPKCS11
- **Gerenciar o Simulador HubSaúde**: iniciar, parar e consultar status
- **Provisionamento automático do JDK 21** quando ausente na máquina
- **Startup inteligente**: detecta processos em execução, auto-seleciona porta disponível
- **Binários multiplataforma** (Windows, Linux, macOS) distribuídos via GitHub Releases
- **Artefatos assinados** com Cosign (Sigstore) para verificação de autenticidade

## Instalação

Baixe o binário para sua plataforma na página de [Releases](../../releases):

| Plataforma | assinatura | simulador |
|------------|------------|-----------|
| Windows | `assinatura-x.y.z-windows-amd64.exe` | `simulador-x.y.z-windows-amd64.exe` |
| Linux | `assinatura-x.y.z-linux-amd64.AppImage` | `simulador-x.y.z-linux-amd64.AppImage` |
| macOS | `assinatura-x.y.z-macos-amd64.dmg` | `simulador-x.y.z-macos-amd64.dmg` |

Nenhum outro pré-requisito — o JDK é provisionado automaticamente no primeiro uso.

### Verificação de integridade

Todos os artefatos são assinados com [Cosign](https://docs.sigstore.dev/cosign/overview/) (Sigstore):

```bash
cosign verify-blob \
  --certificate assinatura-x.y.z-linux-amd64.AppImage.pem \
  --signature assinatura-x.y.z-linux-amd64.AppImage.sig \
  assinatura-x.y.z-linux-amd64.AppImage
```

## Uso

### CLI assinatura

```bash
# Criar assinatura digital (modo HTTP — padrão, inicia o servidor automaticamente)
assinatura criar --payload pedido-assinatura.json

# Criar assinatura (modo local — invocação direta, sem servidor)
assinatura criar --payload pedido-assinatura.json --modo local

# Validar assinatura
assinatura validar --payload pedido-validacao.json

# Ler o payload do stdin
cat pedido-assinatura.json | assinatura criar --payload -

# Gerenciar o servidor manualmente
assinatura servidor iniciar [--porta 8088]
assinatura servidor status
assinatura servidor parar
```

O `--payload` é um arquivo JSON com os parâmetros FHIR. Veja payloads completos em
[docs/exemplos.md](docs/exemplos.md) e o contrato em [docs/tecnico.md](docs/tecnico.md).

### CLI simulador

```bash
# Iniciar o simulador (baixa automaticamente se necessário)
simulador iniciar

# Iniciar com URL alternativa para o simulador.jar
simulador iniciar --source http://servidor-local/simulador.jar

# Iniciar uma versão específica (GitHub Releases tag)
simulador iniciar --versao-simulador v1.2.0

# Forçar uma porta específica
simulador iniciar --porta 9100

# Consultar status (PID, porta, uptime, health)
simulador status

# Parar o simulador (HTTP shutdown → SIGTERM → SIGKILL)
simulador parar
```

### Códigos de saída (CLI assinatura)

O exit code reflete o código de erro retornado pelo `assinador.jar`:

| Exit | Código | Significado |
|------|--------|-------------|
| 0 | — | Sucesso |
| 1 | `CLI_ERROR` | Erro genérico do CLI |
| 2 | `PARAM_AUSENTE` | Campo FHIR obrigatório faltando |
| 3 | `PARAM_INVALIDO` | Campo FHIR mal formado |
| 4 | `ALGORITMO_NAO_SUPORTADO` | Algoritmo não suportado |
| 5 | `ASSINATURA_INVALIDA` | Validação da assinatura falhou |
| 6 | `DISPOSITIVO_INDISPONIVEL` / `PIN_INVALIDO` | PKCS#11 |
| 7 | `ERRO_INTERNO` | Falha interna do `assinador.jar` |

Erros são emitidos em stderr como JSON: `{"error": "PARAM_AUSENTE", "message": "..."}`. Saída de sucesso vai para stdout, sempre como JSON pretty-printed.

> Os códigos 5 e 6 estão **reservados** para a evolução do PKCS#11 e da validação criptográfica real; o `assinador.jar` em modo simulado não os emite hoje.

### Variáveis de ambiente

| Variável | Uso |
|----------|-----|
| `HUBSAUDE_ASSINADOR_JAR` | Caminho explícito para `assinador.jar` |
| `HUBSAUDE_ADOPTIUM_URL` | Sobrescreve o endpoint da API Adoptium (testes) |
| `HUBSAUDE_SIMULADOR_REPO` | Sobrescreve o endpoint do GitHub Releases do simulador |

## Desenvolvimento

### Pré-requisitos

- Go 1.22+
- JDK 21 (ou provisionado automaticamente ao executar os CLIs)

### Build

```bash
# CLIs
cd assinatura && go build ./...
cd simulador  && go build ./...

# assinador.jar
cd assinador && ./mvnw package
```

### Testes

```bash
# CLI (unitários)
cd assinatura && go test ./...
cd simulador  && go test ./...

# CLI (integração end-to-end — requer assinador.jar buildado e java no PATH)
cd assinatura && go test -tags=integration ./internal/assinador/...

# assinador.jar (unitários)
cd assinador && ./mvnw test

# assinador.jar (integração)
cd assinador && ./mvnw verify
```

## Estrutura do Projeto

```
runner/
├── assinatura/          # CLI Go — invocação do assinador.jar
│   ├── cmd/             # Subcomandos cobra (criar, validar, servidor) + exit codes
│   └── internal/        # assinador/ (incl. integration_test), jdk/, porta/, state/, logging/
├── simulador/           # CLI Go — gerenciamento do simulador.jar
│   ├── cmd/             # Subcomandos cobra (iniciar, parar, status)
│   └── internal/        # download/, jdk/, porta/, processo/, state/, logging/
├── assinador/           # Java 21 — Maven
│   └── src/             # servico/, servidor/, cli/, validacao/
├── .github/workflows/
│   ├── ci.yml           # Build + testes nas 3 plataformas
│   └── release.yml      # Cross-compile + Cosign + GitHub Release
├── docs/                # Documentação de uso e técnica
├── planejamento/        # Documentação de planejamento técnico
├── especificações.md    # Especificação do projeto
├── BACKLOG.md           # Backlog por sprint
└── README.md
```

## Documentação

Guias de uso (`docs/`):

- [Manual do Usuário](docs/manual-usuario.md) — referência de comandos, flags e troubleshooting
- [Guia de Instalação](docs/instalacao.md) — download, verificação com Cosign, primeiros passos
- [Documentação Técnica](docs/tecnico.md) — arquitetura, contratos HTTP/CLI, `SignatureService`, PKCS#11
- [Exemplos de Uso](docs/exemplos.md) — payloads completos prontos para copiar

Projeto:

- [Backlog de Entrega](BACKLOG.md)
- [Planejamento Técnico](planejamento/README.md)
- [Especificação](especificações.md)
