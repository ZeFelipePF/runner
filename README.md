# Sistema Runner

Trabalho prático da disciplina **Implementação e Integração** — Bacharelado em Engenharia de Software, UFG (2026-01).

O Sistema Runner facilita a execução de aplicações Java via linha de comandos, ocultando a complexidade de configuração e instalação do ambiente Java. O projeto é de interesse real da Secretaria de Estado de Saúde de Goiás (SES) e da Universidade Federal de Goiás (UFG), no contexto de uma plataforma de interoperabilidade de dados em saúde.

## Arquitetura

O sistema é composto por três aplicações integradas:

```
Usuário
  ├──> assinatura (CLI) ──> assinador.jar (Java)
  │         │                     │
  │         │ (CLI ou HTTP)       └──> Dispositivo Criptográfico (PKCS#11)
  │         │
  └──> simulador  (CLI) ──> simulador.jar (HubSaúde)
                                  │
                                  └──> HTTP
```

| Aplicação | Descrição |
|-----------|-----------|
| **assinatura** | CLI multiplataforma que permite ao usuário invocar operações de assinatura digital |
| **assinador.jar** | Aplicação Java que valida parâmetros e simula criação/validação de assinaturas digitais |
| **simulador** | CLI multiplataforma que gerencia o ciclo de vida do Simulador do HubSaúde |

## Funcionalidades

- **Criar assinatura digital** (simulada) com validação rigorosa de parâmetros FHIR
- **Validar assinatura digital** (simulada) com resultado pré-determinado
- **Dois modos de invocação** do assinador: direto (CLI) e via HTTP (servidor)
- **Gerenciar o Simulador HubSaúde**: iniciar, parar e consultar status
- **Provisionamento automático do JDK** quando ausente na máquina
- **Binários multiplataforma** (Windows, Linux, macOS) distribuídos via GitHub Releases

## Instalação

### Via GitHub Releases (recomendado)

Baixe o binário correspondente à sua plataforma na página de [Releases](../../releases):

| Plataforma | Arquivo |
|------------|---------|
| Windows | `assinatura-x.y.z-windows-amd64.exe` |
| Linux | `assinatura-x.y.z-linux-amd64.AppImage` |
| macOS | `assinatura-x.y.z-macos-amd64.dmg` |

### Verificação de integridade

Os artefatos são assinados com [Cosign](https://docs.sigstore.dev/cosign/overview/) (Sigstore). Para verificar:

```bash
cosign verify-blob \
  --certificate assinatura-x.y.z-linux-amd64.AppImage.pem \
  --signature assinatura-x.y.z-linux-amd64.AppImage.sig \
  assinatura-x.y.z-linux-amd64.AppImage
```

## Uso

### Assinatura digital

```bash
# Criar assinatura (modo CLI — invocação direta)
assinatura criar --arquivo documento.json

# Criar assinatura (modo HTTP — assinador como servidor)
assinatura criar --arquivo documento.json --modo http

# Validar assinatura
assinatura validar --arquivo documento-assinado.json
```

### Simulador do HubSaude

```bash
# Iniciar o simulador
simulador iniciar

# Consultar status
simulador status

# Parar o simulador
simulador parar
```

## Desenvolvimento

### Pré-requisitos

- JDK 21+ (ou será provisionado automaticamente pelo CLI)
- Ferramenta de build do CLI (conforme linguagem adotada)

### Build

```bash
# Compilar o CLI
# (comando conforme linguagem adotada)

# Compilar o assinador.jar
cd assinador
./mvnw package
```

### Testes

```bash
# Testes unitários
./mvnw test

# Testes de integração
./mvnw verify
```

## Estrutura do Projeto

```
runner/
├── assinatura/          # CLI multiplataforma (interface do usuário)
├── assinador/           # Aplicação Java (simulação de assinatura)
├── simulador/           # CLI para gerenciamento do HubSaúde
├── docs/                # Documentação de uso e técnica
├── BACKLOG.md           # Backlog de entrega (sprints)
└── README.md            # Este arquivo
```

## Documentação

- [Backlog de Entrega](BACKLOG.md)
