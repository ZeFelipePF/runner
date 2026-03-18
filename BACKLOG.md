# Sistema Runner - Backlog de Entrega

**Disciplina:** Implementação e Integração — Engenharia de Software (2026-01)
**Período:** 17/03/2026 a 16/06/2026 (3 meses — 6 sprints de 2 semanas)
**Referência:** [Especificação](../runner%20(especificacoes)/especificacao.md) · [Design](../runner%20(especificacoes)/design.md)

---

## Legenda

| Símbolo | Significado |
|---------|-------------|
| 🔲 | Não iniciado |
| 🔄 | Em andamento |
| ✅ | Concluído |

---

## Sprint 1 — Fundação (17/03 – 30/03)

> **Objetivo:** estruturar os repositórios, definir stack tecnológica e entregar o esqueleto funcional do CLI.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 1.1 | Definir linguagem do CLI (ex.: Go, Rust, Python) e justificar a escolha | — | Alta | 🔲 |
| 1.2 | Criar estrutura de projeto do CLI `assinatura` (build, lint, testes) | US-01 | Alta | 🔲 |
| 1.3 | Criar estrutura de projeto Java `assinador.jar` (Maven/Gradle, testes) | US-02 | Alta | 🔲 |
| 1.4 | Implementar esqueleto do CLI com parsing de comandos (`criar`, `validar`) | US-01 | Alta | 🔲 |
| 1.5 | Configurar pipeline CI básico (build + testes em push/PR) | US-05 | Média | 🔲 |
| 1.6 | Escrever testes unitários iniciais para o parsing do CLI | US-01 | Alta | 🔲 |

**Entregável:** CLI compila e exibe `--help` com os subcomandos; CI executa build e testes automaticamente.

---

## Sprint 2 — Assinador: Validação e Simulação (31/03 – 13/04)

> **Objetivo:** implementar o `assinador.jar` com validação rigorosa de parâmetros e respostas simuladas.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 2.1 | Implementar validação de parâmetros FHIR no `assinador.jar` | US-02 | Alta | 🔲 |
| 2.2 | Implementar simulação de criação de assinatura (resposta pré-construída) | US-02 | Alta | 🔲 |
| 2.3 | Implementar simulação de validação de assinatura (resultado pré-determinado) | US-02 | Alta | 🔲 |
| 2.4 | Implementar tratamento de erros e mensagens claras no `assinador.jar` | US-02 | Alta | 🔲 |
| 2.5 | Escrever testes unitários do `assinador.jar` (parâmetros válidos e inválidos) | US-02 | Alta | 🔲 |
| 2.6 | Implementar modo CLI do `assinador.jar` (entrada via args, saída em stdout) | US-02 | Média | 🔲 |

**Entregável:** `assinador.jar` executável via `java -jar` com validação completa e respostas simuladas; cobertura de testes > 80%.

---

## Sprint 3 — Integração CLI ↔ Assinador (14/04 – 27/04)

> **Objetivo:** conectar o CLI `assinatura` ao `assinador.jar` nos dois modos de invocação.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 3.1 | Implementar invocação direta (CLI → `java -jar assinador.jar`) | US-01 | Alta | 🔲 |
| 3.2 | Implementar modo servidor HTTP no `assinador.jar` | US-01 | Alta | 🔲 |
| 3.3 | Implementar invocação via HTTP no CLI (`assinatura` → HTTP → `assinador.jar`) | US-01 | Alta | 🔲 |
| 3.4 | Implementar formatação e exibição dos resultados no CLI | US-01 | Média | 🔲 |
| 3.5 | Implementar propagação estruturada de erros entre as camadas | US-01 | Alta | 🔲 |
| 3.6 | Escrever testes de integração (CLI ↔ Assinador, ambos os modos) | US-01 | Alta | 🔲 |

**Entregável:** fluxo completo de criação e validação de assinatura funcionando nos modos CLI e HTTP; testes de integração passando.

---

## Sprint 4 — Simulador e Provisionamento JDK (28/04 – 11/05)

> **Objetivo:** implementar o gerenciamento do Simulador HubSaúde e o provisionamento automático do JDK.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 4.1 | Implementar comando `simulador iniciar` (download + start do `.jar`) | US-03 | Alta | 🔲 |
| 4.2 | Implementar verificação de portas disponíveis antes de iniciar o Simulador | US-03 | Alta | 🔲 |
| 4.3 | Implementar comando `simulador parar` | US-03 | Alta | 🔲 |
| 4.4 | Implementar comando `simulador status` | US-03 | Média | 🔲 |
| 4.5 | Implementar download dinâmico do `simulador.jar` via GitHub Releases (com cache local) | US-03 | Alta | 🔲 |
| 4.6 | Implementar detecção do JDK na máquina (versão exigida) | US-04 | Alta | 🔲 |
| 4.7 | Implementar download automático do JDK compatível (Windows, Linux, macOS) | US-04 | Alta | 🔲 |
| 4.8 | Escrever testes para provisionamento do JDK e gerenciamento do Simulador | US-03/04 | Alta | 🔲 |

**Entregável:** CLI gerencia ciclo de vida do Simulador; JDK é provisionado automaticamente quando ausente; testes passando nas 3 plataformas.

---

## Sprint 5 — Build Multiplataforma, Assinatura de Artefatos e Testes de Aceitação (12/05 – 25/05)

> **Objetivo:** gerar binários para as 3 plataformas, assinar artefatos com Cosign e validar critérios de aceitação.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 5.1 | Configurar cross-compilation do CLI para Windows, Linux e macOS (amd64) | US-05 | Alta | 🔲 |
| 5.2 | Configurar empacotamento: `.exe` (Win), `.AppImage` (Linux), `.dmg` (macOS) | US-05 | Alta | 🔲 |
| 5.3 | Gerar checksums SHA256 para cada artefato | US-05 | Alta | 🔲 |
| 5.4 | Integrar Cosign no pipeline CI/CD (assinatura OIDC + transparency log) | §9 | Alta | 🔲 |
| 5.5 | Gerar `.sig` e `.pem` para cada artefato na release | §9 | Alta | 🔲 |
| 5.6 | Configurar versionamento semântico (SemVer) e GitHub Releases automáticas | US-05 | Média | 🔲 |
| 5.7 | Escrever testes de aceitação baseados nos critérios das US-01 a US-05 | Todas | Alta | 🔲 |
| 5.8 | Executar testes de aceitação nas 3 plataformas (CI matrix) | Todas | Alta | 🔲 |

**Entregável:** release no GitHub com binários assinados para 3 plataformas; todos os testes de aceitação passando.

---

## Sprint 6 — Documentação, Polimento e Entrega Final (26/05 – 16/06)

> **Objetivo:** finalizar documentação, corrigir pendências e publicar a release 1.0.0.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 6.1 | Escrever manual do usuário para o CLI `assinatura` | §7 | Alta | 🔲 |
| 6.2 | Escrever documentação técnica da integração (CLI ↔ Assinador) | §7 | Alta | 🔲 |
| 6.3 | Escrever guia de instalação (download, verificação com Cosign, primeiros passos) | §7 | Alta | 🔲 |
| 6.4 | Incluir exemplos de uso no README e documentação | §7 | Média | 🔲 |
| 6.5 | Revisar cobertura de testes e preencher lacunas | Todas | Média | 🔲 |
| 6.6 | Revisar tratamento de erros e mensagens ao usuário | US-01/02 | Média | 🔲 |
| 6.7 | Corrigir bugs e pendências acumuladas | Todas | Alta | 🔲 |
| 6.8 | Publicar release 1.0.0 no GitHub Releases | US-05 | Alta | 🔲 |

**Entregável final:** release 1.0.0 publicada com binários assinados, documentação completa, testes passando.

---

## Matriz de Rastreabilidade — User Stories × Sprints

| User Story | Sprint 1 | Sprint 2 | Sprint 3 | Sprint 4 | Sprint 5 | Sprint 6 |
|------------|----------|----------|----------|----------|----------|----------|
| **US-01** Invocar Assinador via CLI | 1.4, 1.6 | — | 3.1–3.6 | — | 5.7, 5.8 | 6.6 |
| **US-02** Simular Assinatura Digital | — | 2.1–2.6 | — | — | 5.7, 5.8 | 6.6 |
| **US-03** Gerenciar Simulador HubSaúde | — | — | — | 4.1–4.5, 4.8 | 5.7, 5.8 | — |
| **US-04** Provisionar JDK | — | — | — | 4.6–4.8 | 5.7, 5.8 | — |
| **US-05** Binários multiplataforma | 1.5 | — | — | — | 5.1–5.6 | 6.8 |

---

## Riscos e Mitigações

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| Empacotamento multiplataforma complexo (`.dmg`, `.AppImage`) | Atraso na Sprint 5 | Investigar ferramentas de empacotamento na Sprint 1; ter fallback para binários simples |
| Integração com Cosign/Sigstore desconhecida pela equipe | Atraso na Sprint 5 | Fazer spike técnico na Sprint 3 |
| Validação de parâmetros FHIR exige domínio específico | Implementação incorreta | Consultar especificações FHIR durante Sprint 2; validar com professor |
| Provisionamento de JDK varia por SO | Bugs em plataformas específicas | Testar em CI matrix (Windows, Linux, macOS) desde a Sprint 4 |
| Escopo subestimado | Entrega incompleta | Priorizar US-01 e US-02 (core); US-03/04 são secundárias |

---

## Definição de Pronto (Definition of Done)

Um item do backlog é considerado **pronto** quando:

1. Código implementado e revisado (code review)
2. Testes unitários escritos e passando
3. Testes de integração passando (quando aplicável)
4. CI pipeline verde
5. Documentação atualizada (quando aplicável)
6. Sem regressões introduzidas
