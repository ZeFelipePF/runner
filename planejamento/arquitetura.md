# Arquitetura do Sistema

## Visao Geral

```
+------------------+         +------------------+
|                  |  CLI    |                  |
|   assinatura     +-------->+  assinador.jar   |
|   (CLI nativo)   |  HTTP   |  (Java)          |
|                  +-------->+                  |
+--------+---------+         +------------------+
         |
         | (usa)
         v
+------------------+
|   JDK Provider   |  <-- Provisiona JDK automaticamente
+------------------+

+------------------+         +------------------+
|                  |  exec   |                  |
|   simulador      +-------->+  simulador.jar   |
|   (CLI nativo)   |         |  (HubSaude)      |
+--------+---------+         +------------------+
         |
         | (usa)
         v
+------------------+
|   JDK Provider   |
+------------------+
```

## Estrutura de Pastas Proposta

```
runner/
├── assinatura/              # CLI multiplataforma (Go/Rust/etc)
│   ├── cmd/                 # Ponto de entrada e definicao de comandos
│   │   ├── root.go          # Comando raiz
│   │   ├── criar.go         # Subcomando: criar assinatura
│   │   ├── validar.go       # Subcomando: validar assinatura
│   │   ├── servidor.go      # Subcomando: gerenciar servidor (iniciar/parar)
│   │   └── versao.go        # Subcomando: exibir versao
│   ├── internal/            # Logica interna
│   │   ├── assinador/       # Cliente para comunicacao com assinador.jar
│   │   │   ├── cli.go       # Invocacao direta (java -jar)
│   │   │   └── http.go      # Invocacao via HTTP
│   │   ├── jdk/             # Deteccao e provisionamento do JDK
│   │   └── config/          # Configuracao (porta, caminhos, etc)
│   ├── go.mod
│   ├── go.sum
│   └── main.go
│
├── simulador/               # CLI multiplataforma (mesma linguagem)
│   ├── cmd/
│   │   ├── root.go
│   │   ├── iniciar.go
│   │   ├── parar.go
│   │   └── status.go
│   ├── internal/
│   │   ├── download/        # Download do simulador.jar via GitHub Releases
│   │   ├── processo/        # Gerenciamento do processo Java
│   │   └── jdk/             # Compartilhado ou reutilizado do assinatura
│   ├── go.mod
│   └── main.go
│
├── assinador/               # Aplicacao Java
│   ├── src/
│   │   ├── main/java/
│   │   │   └── br/gov/saude/assinador/
│   │   │       ├── App.java              # Ponto de entrada
│   │   │       ├── cli/                  # Modo CLI (args -> stdout)
│   │   │       ├── servidor/             # Modo servidor HTTP
│   │   │       ├── servico/              # Logica de negocio
│   │   │       │   ├── AssinaturaService.java
│   │   │       │   └── ValidacaoService.java
│   │   │       ├── validacao/            # Validacao de parametros FHIR
│   │   │       └── modelo/              # DTOs e modelos de dados
│   │   └── test/java/
│   │       └── br/gov/saude/assinador/
│   │           ├── servico/
│   │           ├── validacao/
│   │           └── integracao/
│   ├── pom.xml (ou build.gradle)
│   └── mvnw / gradlew
│
├── shared/                  # Codigo compartilhado entre assinatura e simulador
│   └── jdk/                 # Modulo de provisionamento JDK (se linguagem permitir)
│
├── .github/
│   └── workflows/
│       ├── ci.yml           # Build + testes em push/PR
│       └── release.yml      # Build multiplataforma + Cosign + GitHub Release
│
├── docs/                    # Documentacao de uso
├── planejamento/            # Esta pasta
├── especificacoes.md
├── BACKLOG.md
└── README.md
```

**Nota:** Se `assinatura` e `simulador` compartilham codigo significativo (ex: provisionamento JDK), considerar um monorepo Go com modulos compartilhados ou um workspace.

## Comunicacao entre Componentes

### CLI -> Assinador (Modo Direto)

```
assinatura criar --arquivo doc.json
       |
       v
exec: java -jar assinador.jar criar --documento <base64> --certificado <...>
       |
       v
stdout: JSON com resultado
       |
       v
assinatura: parse JSON, formata e exibe
```

### CLI -> Assinador (Modo HTTP)

```
assinatura criar --arquivo doc.json --modo http
       |
       v
POST http://localhost:8088/assinatura/criar
Content-Type: application/json
Body: { "documento": "<base64>", "certificado": "..." }
       |
       v
Response: 200 OK
Body: { "status": "sucesso", "assinatura": "..." }
       |
       v
assinatura: parse JSON, formata e exibe
```

### Formato de Resposta (proposta)

**Sucesso:**
```json
{
  "status": "sucesso",
  "dados": {
    "assinatura": "MIIBxjCB...",
    "algoritmo": "SHA256withRSA",
    "timestamp": "2026-03-24T10:00:00Z"
  }
}
```

**Erro:**
```json
{
  "status": "erro",
  "codigo": "PARAM_INVALIDO",
  "mensagem": "O parametro 'certificado' e obrigatorio",
  "detalhes": {
    "parametro": "certificado",
    "tipo": "obrigatorio"
  }
}
```

## Decisoes de Arquitetura

- [x] `assinatura` e `simulador` sao dois binarios Go separados
- [x] Modulo JDK duplicado em cada CLI (sem biblioteca compartilhada)
- [ ] Definir contrato exato da API HTTP (endpoints, campos, codigos HTTP)
- [x] Logging: `stdout` para resultado final (JSON); `stderr` para progresso e erros operacionais; `--verbose` para mais detalhe; `--quiet` para suprimir progresso; sem arquivo de log
