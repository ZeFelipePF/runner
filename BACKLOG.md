# Sistema Runner - Backlog de Entrega

**Disciplina:** Implementação e Integração — Engenharia de Software (2026-01)
**Período:** 17/03/2026 a 16/06/2026 (3 meses — 6 sprints de 2 semanas)
**Referência:** [Especificação](especificações.md) · [Planejamento](planejamento/README.md)

---

## Legenda

| Símbolo | Significado |
|---------|-------------|
| 🔲 | Não iniciado |
| 🔄 | Em andamento |
| ✅ | Concluído |

---

## Sprint 1 — Fundação (17/03 – 30/03)

> **Objetivo:** investigar parâmetros FHIR, estruturar os projetos e entregar o esqueleto funcional dos dois CLIs.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 1.1 | Investigar parâmetros FHIR e definir contrato de `/sign` e `/validate` | US-02 | Alta | ✅ |
| 1.2 | Criar estrutura de projeto do CLI `assinatura` (Go, cobra, build, lint, testes) | US-01 | Alta | ✅ |
| 1.3 | Criar estrutura de projeto do CLI `simulador` (Go, cobra, build, lint, testes) | US-03 | Alta | ✅ |
| 1.4 | Criar estrutura de projeto Java `assinador` (Maven + mvnw, JUnit 5, Javalin) | US-02 | Alta | ✅ |
| 1.5 | Implementar esqueleto do CLI `assinatura` com subcomandos (`criar`, `validar`, `servidor`) | US-01 | Alta | ✅ |
| 1.6 | Implementar esqueleto do CLI `simulador` com subcomandos (`iniciar`, `parar`, `status`) | US-03 | Alta | ✅ |
| 1.7 | Implementar `~/.hubsaude/` com `state.json` e `config.json` (leitura, escrita, inicialização) | US-01/03 | Alta | ✅ |
| 1.8 | Configurar pipeline CI básico (build + testes, matrix 3 SOs) | US-05 | Média | ✅ |
| 1.9 | Escrever testes unitários do parsing de comandos e do `state.json` | US-01 | Alta | ✅ |

**Entregável:** ambos os CLIs compilam e exibem `--help`; `~/.hubsaude/` inicializado corretamente; CI verde nas 3 plataformas.

---

## Sprint 2 — Assinador: Validação e Simulação (31/03 – 13/04)

> **Objetivo:** implementar o `assinador.jar` com a interface `SignatureService`, validação rigorosa de parâmetros FHIR e respostas simuladas.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 2.1 | Definir interface `SignatureService` (`sign`, `validate`) | US-02 | Alta | ✅ |
| 2.2 | Implementar `FakeSignatureService` (resposta simulada pré-construída) | US-02 | Alta | ✅ |
| 2.3 | Projetar esqueleto de `PKCS11SignatureService` (SunPKCS11, sem dispositivo físico) | US-02 | Média | ✅ |
| 2.4 | Implementar validação de parâmetros FHIR (campos obrigatórios, base64, algoritmo) | US-02 | Alta | ✅ |
| 2.5 | Implementar tratamento de erros estruturado (códigos `PARAM_AUSENTE`, `PARAM_INVALIDO`, etc.) | US-02 | Alta | ✅ |
| 2.6 | Implementar modo CLI do `assinador.jar` (args → JSON stdout/stderr) | US-02 | Alta | ✅ |
| 2.7 | Escrever testes unitários do `assinador.jar` (parâmetros válidos e inválidos, cobertura > 80%) | US-02 | Alta | ✅ |

**Entregável:** `assinador.jar` executável via `java -jar` com interface `SignatureService`, validação completa e `FakeSignatureService`; cobertura > 80%.

---

## Sprint 3 — Integração CLI ↔ Assinador (14/04 – 27/04)

> **Objetivo:** conectar o CLI `assinatura` ao `assinador.jar` nos dois modos, implementar startup inteligente e gerenciamento de servidor.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 3.1 | Implementar `SignatureController` com Javalin (endpoints `/sign`, `/validate`, `/health`, `/shutdown`) | US-01 | Alta | ✅ |
| 3.2 | Implementar invocação direta no CLI (`os/exec` → `java -jar assinador.jar`) | US-01 | Alta | ✅ |
| 3.3 | Implementar invocação via HTTP no CLI (`POST /sign`, `POST /validate`) | US-01 | Alta | ✅ |
| 3.4 | Implementar startup do CLI: detecção de instância existente via `state.json` + PID | US-01 | Alta | ✅ |
| 3.5 | Implementar auto-detecção de porta disponível (8088, 8089... +20) | US-01 | Alta | ✅ |
| 3.6 | Gravar PID e porta em `state.json` ao iniciar servidor; limpar ao parar | US-01 | Alta | ✅ |
| 3.7 | Implementar formatação e exibição dos resultados no CLI (`stdout`/`stderr`) | US-01 | Média | ✅ |
| 3.8 | Implementar propagação estruturada de erros entre as camadas | US-01 | Alta | ✅ |
| 3.9 | Escrever testes de integração (CLI ↔ Assinador, modo direto e HTTP) | US-01 | Alta | ✅ |

**Entregável:** fluxo completo de criação e validação de assinatura nos modos local e HTTP; startup detecta instância existente; porta auto-selecionada; testes de integração passando.

---

## Sprint 4 — Simulador e Provisionamento JDK (28/04 – 11/05)

> **Objetivo:** implementar o gerenciamento do Simulador HubSaúde e o provisionamento automático do JDK via Adoptium.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 4.1 | Implementar detecção do JDK na máquina (JAVA_HOME, PATH, `~/.hubsaude/jdk/`) | US-04 | Alta | ✅ |
| 4.2 | Implementar download automático do JDK 21 Temurin via Adoptium API (Windows, Linux, macOS) | US-04 | Alta | ✅ |
| 4.3 | Implementar comando `simulador iniciar` com auto-detecção de porta | US-03 | Alta | ✅ |
| 4.4 | Implementar download dinâmico do `simulador.jar` via GitHub Releases (com cache local e flag `--source`) | US-03 | Alta | ✅ |
| 4.5 | Implementar verificação de versão e integridade (SHA256) do `simulador.jar` | US-03 | Alta | ✅ |
| 4.6 | Implementar startup do CLI `simulador`: `state.json`, PID, porta | US-03 | Alta | ✅ |
| 4.7 | Implementar comando `simulador parar` (HTTP shutdown + fallback SIGTERM) | US-03 | Alta | ✅ |
| 4.8 | Implementar comando `simulador status` | US-03 | Média | ✅ |
| 4.9 | Escrever testes para provisionamento JDK e gerenciamento do Simulador | US-03/04 | Alta | ✅ |

**Entregável:** CLI gerencia ciclo de vida do Simulador; JDK provisionado automaticamente; `--source` funciona; testes passando nas 3 plataformas.

---

## Sprint 5 — Build Multiplataforma, Assinatura de Artefatos e Testes de Aceitação (12/05 – 25/05)

> **Objetivo:** gerar binários para as 3 plataformas, assinar artefatos com Cosign, incluir `assinador.jar` na release e validar critérios de aceitação.

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 5.1 | Configurar cross-compilation Go para Windows, Linux e macOS (amd64) | US-05 | Alta | ✅ |
| 5.2 | Configurar empacotamento: `.exe` (Win), `.AppImage` (Linux), `.dmg` (macOS) | US-05 | Alta | ✅ |
| 5.3 | Incluir `assinador.jar` como artefato da release | US-05 | Alta | ✅ |
| 5.4 | Gerar checksums SHA256 para todos os artefatos | US-05 | Alta | ✅ |
| 5.5 | Integrar Cosign no pipeline de release (OIDC + `.sig` + `.pem` por artefato) | §9 | Alta | ✅ |
| 5.6 | Configurar versionamento SemVer e GitHub Releases automáticas (tag `v*`) | US-05 | Média | ✅ |
| 5.7 | Escrever testes de aceitação baseados nos critérios das US-01 a US-05 | Todas | Alta | ✅ |
| 5.8 | Executar testes de aceitação nas 3 plataformas (CI matrix) | Todas | Alta | ✅ |

**Entregável:** release no GitHub com binários assinados (`.exe`, `.AppImage`, `.dmg`) + `assinador.jar` para 3 plataformas; todos os testes de aceitação passando.

---

## Sprint 6 — Documentação, Conformidade com a Spec e Entrega Final (26/05 – 16/06)

> **Objetivo:** finalizar documentação, fechar lacunas frente à especificação do professor (`runnerProfessor`), corrigir pendências e publicar a release final.
> Detalhamento técnico dos itens 6.5–6.14 está em [`planejamento/pendencias-spec.md`](planejamento/pendencias-spec.md).

### 6.A — Documentação (já concluída)

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 6.1 | Escrever manual do usuário (`assinatura` e `simulador`) | §7 | Alta | ✅ |
| 6.2 | Escrever documentação técnica da integração (CLI ↔ Assinador, `SignatureService`, PKCS#11) | §7 | Alta | ✅ |
| 6.3 | Escrever guia de instalação (download, verificação com Cosign, primeiros passos) | §7 | Alta | ✅ |
| 6.4 | Incluir exemplos de uso no README e documentação | §7 | Média | ✅ |

### 6.B — Conformidade com a especificação do professor (gaps detectados)

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 6.5 | **Porta padrão do Simulador = 8443** (alterar `config.go`, `processo.go`, testes, docs) | US-03 | Alta | ✅ |
| 6.6 | **Status do Simulador via `/api/info`** (mantendo `/actuator/health` como fallback) | US-03 | Alta | ✅ |
| 6.7 | **Provisionar JRE em vez de JDK** (Adoptium path `/jre/`, ajustar `Garantir`/`Detectar`/metadata) | US-04 | Alta | ✅ |
| 6.8 | **Implementar `--timeout <minutos>` do `servidor iniciar`** (auto-shutdown por inatividade) | US-01.9 | Alta | ✅ |
| 6.9 | **Diagramas C4 PlantUML** (`diagramas/contexto.puml`, `diagramas/conteineres.puml`, script de geração, SVG/PNG) | §7.5 | Alta | ✅ |
| 6.10 | **Decisão de escopo do "código-fonte do Simulador HubSaúde"** — registrar formalmente em `planejamento/` e `CLAUDE.md` que é dependência externa (com justificativa), OU criar stub mínimo | §7.7 | Alta | ✅ |
| 6.11 | **Integração PKCS#11 (US-02.5 completa):** (a) implementar `PKCS11SignatureService` via `SunPKCS11`; (b) erros `DISPOSITIVO_INDISPONIVEL`/`PIN_INVALIDO` claros no Java (alinhar com `cmd/exit.go`); (c) teste opt-in `-tags=pkcs11` com SoftHSM2; (d) documentar setup do SoftHSM2 (Linux/macOS/Windows) em `docs/tecnico.md` | US-02.5 | Alta | ✅ |
| 6.12 | **Documento de requisitos não-funcionais** (`planejamento/requisitos-nao-funcionais.md`) — desempenho, portabilidade, segurança, observabilidade conforme ISO 25010 | §planejamento | Média | ✅ |
| 6.13 | **Estratégia opcional `release.json`** — avaliar implementar URL fixa `raw.githubusercontent.com/.../release.json` para descoberta do simulador.jar; documentar decisão | US-03 | Baixa | ✅ |
| 6.14 | **Atualizar README** com diagramas C4, links para requisitos não-funcionais e nota sobre escopo do simulador.jar | §7 | Média | ✅ |

### 6.C — Polimento e entrega final (executar nesta ordem, após 6.B)

| # | Item | US | Prioridade | Status |
|---|------|----|------------|--------|
| 6.15 | Revisar cobertura de testes e preencher lacunas | Todas | Média | ✅ |
| 6.16 | Revisar tratamento de erros e mensagens ao usuário (alinhar enums Go ↔ Java) | US-01/02 | Média | ✅ |
| 6.17 | Corrigir bugs e pendências acumuladas; limpar arquivos órfãos da raiz | Todas | Alta | ✅ |
| 6.18 | **Publicar release final no GitHub Releases (tag `v1.0.0`)** — última ação (depende do push da tag pelo mantenedor) | US-05 | Alta | 🔲 |

**Detalhamento da execução de 6.15–6.17 (2026-05-26):**

- **6.15 (cobertura):** `mvnw clean verify` agora passa o gate JaCoCo (**88,6%** de linhas, exclusão correta do `PKCS11SignatureService`). `AssinadorServidor` subiu de 25%→61% com novo teste do auto-shutdown por inatividade. Suíte Java: **74 testes** verdes; suíte Go: ambos os módulos verdes.
- **6.16 (enums):** `assinatura/cmd/exit.go` agora reconhece explicitamente `PAYLOAD_MUITO_GRANDE` (→ exit 1), espelhando `MapeadorErro` (Java). Enums Go ↔ Java totalmente alinhados — não há mais código "reservado".
- **6.17 (bugs/limpeza):** corrigidos dois defeitos reais que quebravam `mvnw verify`:
  1. **JaCoCo 0.8.12 → 0.8.13** — 0.8.12 não instrumenta bytecode Java 24 (major 68); derrubava a JVM do surefire quando o build roda sob JDK > 21.
  2. **`/shutdown` e auto-shutdown chamavam `System.exit(0)` durante os testes**, derrubando o runner de testes. A ação de desligamento foi tornada **injetável** (`SignatureController`/`AssinadorServidor`): produção mantém `System.exit(0)`, testes usam no-op/latch.
  - Raiz já estava limpa (`target/`, `dependency-reduced-pom.xml`, `.claude/` ignorados pelo `.gitignore`).

**Entregável final:** release `v1.0.0` publicada com binários assinados, documentação completa (incluindo diagramas C4 e requisitos não-funcionais), todos os critérios de aceitação da especificação atendidos, testes passando nas 3 plataformas.

> **Pendência única (6.18):** todo o código, testes e documentação estão prontos. Falta apenas o mantenedor enviar a tag `v1.0.0` (`git tag v1.0.0 && git push origin v1.0.0`), o que dispara o `release.yml` (binários + Cosign + GitHub Releases). Esta é uma ação externa/irreversível e deve ser feita manualmente.

---

## Matriz de Rastreabilidade — User Stories × Sprints

| User Story | Sprint 1 | Sprint 2 | Sprint 3 | Sprint 4 | Sprint 5 | Sprint 6 |
|------------|----------|----------|----------|----------|----------|----------|
| **US-01** Invocar Assinador via CLI | 1.2, 1.5, 1.7, 1.9 | — | 3.1–3.9 | — | 5.7, 5.8 | 6.8, 6.16 |
| **US-02** Simular Assinatura Digital | 1.1, 1.4 | 2.1–2.7 | — | — | 5.7, 5.8 | 6.11, 6.16 |
| **US-03** Gerenciar Simulador HubSaúde | 1.3, 1.6, 1.7 | — | — | 4.3–4.9 | 5.7, 5.8 | 6.5, 6.6, 6.10, 6.13 |
| **US-04** Provisionar JDK/JRE | — | — | — | 4.1, 4.2, 4.9 | 5.7, 5.8 | 6.7 |
| **US-05** Binários multiplataforma | 1.8 | — | — | — | 5.1–5.6 | 6.18 |
| **§7.5 Diagramas C4** | — | — | — | — | — | 6.9 |
| **§planejamento NFRs** | — | — | — | — | — | 6.12 |

---

## Riscos e Mitigações

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| Parâmetros FHIR não investigados antes da Sprint 2 | Interface incorreta, retrabalho | Investigar na Sprint 1 (item 1.1) antes de qualquer implementação do assinador |
| Empacotamento `.AppImage` e `.dmg` complexo | Atraso na Sprint 5 | Spike técnico na Sprint 3; fallback para binários simples renomeados |
| Integração com Cosign desconhecida | Atraso na Sprint 5 | Spike técnico junto ao empacotamento na Sprint 3 |
| PKCS#11 sem dispositivo físico disponível | Implementação incompleta | Entregar esqueleto documentado; `FakeSignatureService` cobre o escopo da simulação |
| Provisionamento de JDK varia por SO | Bugs em plataformas específicas | CI matrix (Windows, Linux, macOS) ativo desde a Sprint 1 |
| Escopo subestimado | Entrega incompleta | Priorizar US-01 e US-02 (core); US-03/04 são secundárias |
| Gaps frente à spec descobertos tarde (6.5–6.14) | Não atendimento de critérios obrigatórios | Lista exaustiva em `planejamento/pendencias-spec.md`; itens priorizados Alta executados antes de 6.15–6.18 |
| Mudança da porta 9090 → 8443 pode quebrar deploys/testes existentes | Regressões | Migrar via constante única `PortaPadraoSimulador`; rodar suíte completa antes de mergear |
| SoftHSM2 indisponível no CI | Teste PKCS#11 falha em runners | Marcar teste como opt-in com build tag `pkcs11`; executar localmente |

---

## Definição de Pronto (Definition of Done)

Um item do backlog é considerado **pronto** quando:

1. Código implementado e revisado
2. Testes unitários escritos e passando
3. Testes de integração passando (quando aplicável)
4. CI pipeline verde nas 3 plataformas
5. Documentação atualizada (quando aplicável)
6. Sem regressões introduzidas
