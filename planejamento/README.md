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

Fase atual: **Planejamento** - decisoes tecnicas pendentes.
Investigacao de parametros FHIR pendente (ver entregavel-cli-assinatura.md e entregavel-assinador.md).
