# Planejamento - Sistema Runner

Documentacao de planejamento tecnico do Sistema Runner.
Cada arquivo cobre um aspecto do projeto. Use este indice para navegacao rapida.

## Indice

| Arquivo | Conteudo |
|---------|----------|
| [resumo.md](resumo.md) | Resumo executivo e checklist geral de decisoes/to-dos |
| [decisoes-tecnicas.md](decisoes-tecnicas.md) | Escolhas de linguagem, frameworks e ferramentas (com alternativas) |
| [arquitetura.md](arquitetura.md) | Estrutura de projeto, modulos e comunicacao entre componentes |
| [estado-local.md](estado-local.md) | "Banco de dados" local dos CLIs (~/.hubsaude/): state.json, config.json, metadados |
| [startup.md](startup.md) | Sequencia de inicializacao dos CLIs (JDK, state, porta, processo) |
| [entregavel-cli-assinatura.md](entregavel-cli-assinatura.md) | Planejamento do CLI `assinatura` (US-01) |
| [entregavel-assinador.md](entregavel-assinador.md) | Planejamento do `assinador.jar` — SignatureService, PKCS#11, SignatureController (US-02) |
| [entregavel-simulador.md](entregavel-simulador.md) | Planejamento do CLI `simulador`, download com --source, auto-porta (US-03) |
| [entregavel-jdk.md](entregavel-jdk.md) | Planejamento do provisionamento automatico do JDK (US-04) |
| [entregavel-distribuicao.md](entregavel-distribuicao.md) | Build multiplataforma, releases e assinatura com Cosign (US-05, S9) |
| [entregavel-testes.md](entregavel-testes.md) | Estrategia de testes (unitarios, integracao, aceitacao) |
| [entregavel-documentacao.md](entregavel-documentacao.md) | Planejamento da documentacao de uso e tecnica |
| [ci-cd.md](ci-cd.md) | Pipeline de CI/CD e automacao |

## Status

Fase atual: **Sprint 6 — entrega final** (Sprints 1-5 concluidas).

- **Documentacao (6.1-6.4):** concluida em [`../docs/`](../docs/) — manual do usuario, guia de
  instalacao, documentacao tecnica e exemplos.
- **Conformidade com a spec do professor (6.5-6.14):** concluida — porta 8443, `/api/info`,
  provisionamento de JRE, `--timeout`, diagramas C4 (com SVGs renderizados), PKCS#11 funcional,
  requisitos nao-funcionais (ISO 25010), estrategia `release.json`. Rastreabilidade item-a-item
  em [pendencias-spec.md](pendencias-spec.md).
- **Polimento (6.15-6.17):** concluido — cobertura JaCoCo **88,6%** (gate 80%), **74 testes**
  Java + suites Go verdes; enums de erro Go ↔ Java alinhados; corrigidos os defeitos de build
  (JaCoCo 0.8.13 para Java 24; acao de `/shutdown`/auto-shutdown tornada injetavel para nao
  chamar `System.exit` durante os testes).
- **Pendente (6.18):** apenas a publicacao da release `v1.0.0` (push da tag `v*` dispara o
  `release.yml` — acao manual do mantenedor).

Decisoes tecnicas e investigacao FHIR ja consolidadas (ver decisoes-tecnicas.md e contrato-fhir.md).
