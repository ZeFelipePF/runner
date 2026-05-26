# Pendências de Conformidade com a Especificação

Documento de referência para fechar a Sprint 6 atendendo a 100% dos critérios da
[especificação do professor](../../runnerProfessor/especificacao.md), do
[design C4](../../runnerProfessor/design.md) e do [plano revisitado v2](../../runnerProfessor/docs/plano-revisitado-v2.md).

Cada item lista: **origem na spec**, **estado atual**, **plano de ação**,
**arquivos afetados** e **critérios de aceitação**. Numeração casa com o
`BACKLOG.md` (itens 6.5 a 6.14, seção 6.B).

---

## 6.5 — Porta padrão do Simulador = 8443

**Origem:** `especificacao.md` US-03, primeiro critério.
> "O CLI deve verificar se a porta padrão, empregada pelo Simulador, **8443**, está disponível antes de tentar iniciá-lo."

**Estado atual:** projeto usa **9090**:
- `simulador/internal/state/config.go:45` (`PortaPadrao: 9090`)
- `simulador/internal/processo/processo.go:143` (`inicio = 9090`)
- Testes em `processo_test.go` e `state_test.go`

**Plano de ação:**
1. Introduzir constante `PortaPadraoSimulador = 8443` em `simulador/internal/state/config.go`.
2. Substituir literais `9090` em `processo.go` e `state/config.go` pela constante.
3. Atualizar testes (`processo_test.go`, `state_test.go`, `iniciar_test.go`) — usar a constante para que mudanças futuras sejam transparentes.
4. Atualizar docs: `docs/manual-usuario.md`, `docs/tecnico.md`, `docs/exemplos.md`, `CLAUDE.md`.
5. Manter `--porta N` para override.

**Critérios de aceitação:**
- [ ] `simulador iniciar` (sem `--porta`) tenta 8443; auto-detecta na janela 8443..8463 se ocupada.
- [ ] Todos os testes existentes passam com a nova porta.
- [ ] Nenhuma referência literal a `9090` no código de produção (`grep -rn 9090 simulador/internal` retorna apenas comentários ou nada).

---

## 6.6 — Status do Simulador via `/api/info`

**Origem:** `especificacao.md` US-03.
> "O status pode ser obtido via endpoint **/api/info**."

**Estado atual:** `simulador/internal/processo/processo.go:170-171` consulta apenas `/actuator/health` e `/health`. Comando `status` (`Consultar`) não acessa `/api/info`.

**Plano de ação:**
1. Em `Consultar()` (`processo.go:275`), após confirmar PID vivo, fazer `GET http://127.0.0.1:<porta>/api/info` com timeout curto (~2s).
2. Adicionar campo `info map[string]any` no struct de retorno (`StatusSimulador` ou similar).
3. Imprimir bloco `info` no JSON do comando `status` quando disponível.
4. Manter `/actuator/health` para health check de startup (não substituir).
5. Tolerar 404 / connection refused — o endpoint pode não existir em todas as versões do `simulador.jar`.

**Critérios de aceitação:**
- [ ] `simulador status` exibe campo `info` quando `/api/info` responde 200.
- [ ] Quando endpoint indisponível, campo `info` é omitido (não causa erro).
- [ ] Teste unitário com `httptest` simulando `/api/info`.

---

## 6.7 — Provisionar JRE em vez de JDK

**Origem:** `especificacao.md` US-03 e US-04.
> "O CLI deve **baixar o JRE** caso não esteja disponível no diretório .hubsaude a partir do Eclipse Temurin (Adoptium)."

E nos exemplos de URL do `release.json`:
```
https://api.adoptium.net/v3/binary/latest/21/ga/linux/x64/jre/hotspot/normal/eclipse
```

**Estado atual:** pacote `internal/jdk/` (duplicado em assinatura e simulador) baixa **JDK 21 completo** (path `/jdk/`). Diretório-alvo é `~/.hubsaude/jdk/`.

**Plano de ação:**
1. Renomear pacote para `internal/jre/` (ou manter `jdk/` por compatibilidade, mas alterar internamente). **Recomendado manter o nome do pacote `jdk` e mudar apenas a URL** para minimizar churn — adicionar comentário explicando.
2. Alterar URL na função `Garantir`: `image_type=jre` (Adoptium API) ou path `/21/ga/<os>/<arch>/jre/hotspot/normal/eclipse`.
3. Alterar diretório de destino: `~/.hubsaude/jre/temurin-21.x/` (ou manter `jdk/` e documentar que contém JRE).
4. Função `Detectar` continua procurando `java`/`javac` — `java` está em ambos; `javac` é exclusivo do JDK. Não usar `javac` como prova de presença.
5. Atualizar `hubsaude-meta.json` para registrar `imageType: "jre"`.
6. Sobrescrever via env já existente (`HUBSAUDE_ADOPTIUM_URL`) para testes.
7. Atualizar `docs/tecnico.md`, `docs/instalacao.md`, `CLAUDE.md` (decisão técnica + tabela de decisões).
8. Refletir em ambos `assinatura/internal/jdk/` e `simulador/internal/jdk/`.

**Critérios de aceitação:**
- [ ] Download produz JRE (sem `javac` no diretório).
- [ ] Testes existentes em `jdk/` passam com novo path/imageType.
- [ ] `hubsaude-meta.json` registra `imageType: "jre"`.
- [ ] CLAUDE.md e `decisoes-tecnicas.md` atualizados.

---

## 6.8 — Auto-shutdown do servidor por inatividade (`--timeout`)

**Origem:** `especificacao.md` US-01, último critério.
> "O CLI deve permitir a requisição de interrupção programada do assinador.jar após o número de minutos fornecidos sem interação."

**Estado atual:** flag declarada com texto **"não implementado"**: `assinatura/cmd/servidor_iniciar.go:68`.

**Opções de implementação:**

### Opção A — Auto-shutdown gerenciado pelo Java (preferida)
Adicionar parâmetro `--idle-timeout-minutes N` ao `AssinadorServidor`. O servidor Javalin mantém timestamp da última requisição em interceptor; uma `ScheduledExecutorService` verifica a cada minuto e, se `agora - ultimoAcesso > timeout`, chama `app.stop()` e `System.exit(0)`.

Vantagens: o CLI só passa o parâmetro; não precisa supervisionar processo após `iniciar`.

### Opção B — Watcher no CLI Go
Goroutine no CLI que, após `iniciar`, verifica `/health` periodicamente e tempo desde última requisição (precisaria que o jar expusesse `/metrics` ou similar). Mais frágil.

**Recomendação: Opção A.**

**Plano de ação (Opção A):**
1. Em `AssinadorServidor.java`, aceitar arg `--idle-timeout-minutes N` (0 = desativado, default).
2. Registrar `before` handler que atualiza `lastRequestAt = Instant.now()`.
3. `ScheduledExecutorService.scheduleAtFixedRate` a cada 30s: se `lastRequestAt + N min < now`, log informativo + `app.stop()` + `System.exit(0)`.
4. No CLI Go (`assinatura/cmd/servidor_iniciar.go`):
   - Remover sufixo "— nao implementado" do help.
   - Quando `--timeout > 0`, repassar `--idle-timeout-minutes <N>` ao processo java.
   - `startup.Garantir` precisa propagar o valor (assinatura via método novo `GarantirComTimeout`).
5. Teste Java: subir servidor com timeout=1 min em teste de integração rápido (com `Duration` mockado ou tempo reduzido para 5s via override).
6. Teste Go: verificar que o flag é repassado ao processo.

**Critérios de aceitação:**
- [ ] `assinatura servidor iniciar --timeout 5` faz o jar desligar sozinho após 5 min sem requisições.
- [ ] `state.json` é limpo (CLI detecta servidor morto em `status`/`parar` subsequentes).
- [ ] Help do flag não menciona mais "não implementado".

---

## 6.9 — Diagramas C4 PlantUML

**Origem:** `especificacao.md` §7 entregável 5 + `design.md` (referencia diagramas C4).

**Estado atual:** nenhum diagrama no projeto (`find . -name "*.puml"` vazio).

**Plano de ação:**
1. Criar diretório `diagramas/` na raiz do projeto.
2. Criar `diagramas/contexto.puml` (Nível 1 — C4 Context):
   - Atores: Usuário.
   - Sistemas externos: Dispositivo de Assinatura (PKCS#11), HubSaúde Simulator (web), GitHub Releases (download), Adoptium API.
   - Sistema central: Sistema Runner.
3. Criar `diagramas/conteineres.puml` (Nível 2 — C4 Container):
   - Containers internos: CLI `assinatura`, CLI `simulador`, `assinador.jar` (CLI + server).
   - Relações: CLI → jar via `os/exec` (local) ou HTTP (server); CLI `simulador` → simulador.jar; jar → PKCS#11.
4. Usar inclusão do `<C4/Container>` / `<C4/Context>` (PlantUML stdlib).
5. Criar `geraimagens.sh` e `geraimagens.bat` que renderizam SVG/PNG via `plantuml`.
6. Gerar `diagramas/imagens/contexto.svg` e `conteineres.svg`, comitar.
7. Linkar do `README.md` e `docs/tecnico.md`.

**Critérios de aceitação:**
- [ ] Arquivos `.puml` válidos (renderizam sem erro).
- [ ] SVGs comitados em `diagramas/imagens/`.
- [ ] README exibe os diagramas.

---

## 6.10 — Decisão de escopo do "código-fonte do Simulador HubSaúde"

**Origem:** `especificacao.md` §7 entregável 7.
> "Código fonte do Simulador do HubSaúde — implementação completa, código bem documentado, compatível com Windows/Linux/macOS."

**Conflito:** `CLAUDE.md` declara que o simulador.jar **não é desenvolvido** neste projeto (é dependência externa baixada via GitHub Releases). Não há registro formal dessa decisão no diretório `planejamento/`.

**Plano de ação:**
1. Adicionar seção em `planejamento/decisoes-tecnicas.md` registrando formalmente: **Decisão DT-XX — Simulador HubSaúde tratado como dependência externa.**
   - Contexto: simulador.jar é parte do ecossistema HubSaúde, mantido pela equipe da plataforma. Replicar seu código no escopo deste TP seria inviável e não traria valor pedagógico (foco é integração).
   - Alternativa avaliada: criar stub mínimo Javalin com `/api/info` e `/actuator/health`.
   - **Decisão final a confirmar com o professor.**
2. Adicionar nota explícita no `README.md` na seção de escopo.
3. (Opcional, se professor exigir) criar `simulador-stub/` Maven com endpoints mínimos para testes locais — mas marcado claramente como "stub para desenvolvimento, não é o simulador oficial".

**Critérios de aceitação:**
- [ ] `planejamento/decisoes-tecnicas.md` contém DT registrando a decisão.
- [ ] README esclarece que simulador.jar é dependência externa.
- [ ] (Se aplicável) stub criado com pelo menos `/actuator/health` e `/api/info`.

---

## 6.11 — Integração PKCS#11 (US-02.5 completa)

**Origem:** consolidação de **6 menções** do professor a PKCS#11:

| Documento do professor | Exigência |
|---|---|
| `especificacao.md` US-02 | "O assinador.jar deve **suportar interação com dispositivo criptográfico (token/smart card) via interface PKCS#11**." |
| `design.md` (C4 nível 2) | Relação obrigatória `assinador.jar → Dispositivo Criptográfico — PKCS#11` no diagrama (coberto em 6.9) |
| `plano-revisitado-v2.md` US-02.5 | 4 critérios: SunPKCS11; **testes com SoftHSM2**; **mensagem clara quando indisponível**; **documentação de setup** |
| `plano-revisitado.md` Sprint 2 | "Experimentação com SunPKCS11... indicação é SoftHSM2... inclui testes de integração" |
| `plano-preliminar.md` | Conceitual — chave privada não acessível como parâmetro; SunPKCS11 é a "JDBC dos tokens" |
| `diagramas/contexto.puml` + `conteineres.puml` | Atores e relações PKCS#11 modelados |

**Estado atual:**
- `PKCS11SignatureService.java` (63 linhas) — esqueleto: construtor + validação de args + métodos lançam `UnsupportedOperationException`.
- Códigos `DISPOSITIVO_INDISPONIVEL` / `PIN_INVALIDO` existem em **Go** (`assinatura/cmd/exit.go` codes 6), mas o **Java nunca os emite** — gap de alinhamento.
- `docs/tecnico.md` já tem seção "5.1 PKCS#11 (esqueleto)" explicando o design, mas **sem setup do SoftHSM2**.
- Nenhum teste com SoftHSM2.

**Plano de ação — quatro frentes (a, b, c, d) cobrindo todos os critérios do professor:**

### (a) Implementação funcional via SunPKCS11
1. Em `PKCS11SignatureService.sign(payload, parametros)`:
   - Aceitar `parametros.materialCriptografico.tipo = "SMARTCARD"` ou `"TOKEN"` (já previsto em `planejamento/contrato-fhir.md`).
   - Construir Provider: `Security.getProvider("SunPKCS11").configure(pkcs11ConfigPath)`.
   - `KeyStore ks = KeyStore.getInstance("PKCS11", provider); ks.load(null, pin);`.
   - `PrivateKey k = (PrivateKey) ks.getKey(alias, null);`.
   - `Signature sig = Signature.getInstance("SHA256withRSA"); sig.initSign(k); sig.update(payload); byte[] s = sig.sign();`.
   - Retornar `Base64.encode(s)`.
2. Em `validate(...)`: mesmo padrão usando `PublicKey` do certificado armazenado no token (`ks.getCertificate(alias).getPublicKey()`).
3. Tornar `SignatureController` capaz de escolher entre `FakeSignatureService` e `PKCS11SignatureService` por configuração (env `ASSINADOR_PROVIDER=fake|pkcs11`).

### (b) Tratamento de erros explícito (mensagem clara)
4. Criar enum `AssinadorException.Codigo.DISPOSITIVO_INDISPONIVEL` e `PIN_INVALIDO` no Java (atualmente só existem em Go).
5. Mapear exceções nativas:
   - `ProviderException` com causa `IOException`/`UnsatisfiedLinkError` → `DISPOSITIVO_INDISPONIVEL` ("Driver PKCS#11 não pôde ser carregado: <caminho>")
   - `LoginException` / `FailedLoginException` → `PIN_INVALIDO` ("PIN incorreto para o token")
   - `KeyStoreException` com mensagem "no such alias" → `PARAM_INVALIDO` ("alias '<x>' não encontrado no dispositivo")
6. Adicionar no `MapeadorErro` (Java) → HTTP 503 para `DISPOSITIVO_INDISPONIVEL`, 401 para `PIN_INVALIDO`.
7. Adicionar no `cmd/exit.go` (Go) — já reservados nos códigos 6; remover comentário "reservado".

### (c) Testes com SoftHSM2 (opt-in)
8. Criar `assinador/src/test/java/.../PKCS11IntegrationTest.java` com `@Tag("pkcs11")`:
   - Setup: criar `pkcs11.cfg` apontando para `${SOFTHSM2_LIB}` (env).
   - Inicializar token, importar par RSA-2048 via `pkcs11-tool`.
   - Caso 1: assinar e verificar `Verified OK`.
   - Caso 2: PIN errado → `PIN_INVALIDO`.
   - Caso 3: driver inexistente → `DISPOSITIVO_INDISPONIVEL`.
9. Configurar Maven Surefire: `<excludedGroups>pkcs11</excludedGroups>` no `pom.xml` (CI passa sem SoftHSM2; localmente `mvn test -Dgroups=pkcs11`).

### (d) Documentação de setup
10. Criar nova seção em `docs/tecnico.md` § 5.2 **"Setup do SoftHSM2 para desenvolvimento"** cobrindo:
    - **Linux:** `apt-get install softhsm2 opensc; softhsm2-util --init-token --slot 0 --label "test" --pin 1234 --so-pin 1234`.
    - **macOS:** `brew install softhsm opensc`.
    - **Windows:** instalar SoftHSM2 do release oficial + abrir terminal com PATH ajustado.
    - Importar par de chaves via `pkcs11-tool --module $SOFTHSM2_LIB --login --pin 1234 --keypairgen --key-type RSA:2048 --label testkey`.
    - Exemplo de `pkcs11.cfg` para SunPKCS11.
    - Variáveis de ambiente esperadas (`SOFTHSM2_LIB`, `SOFTHSM2_CONF`).
11. Adicionar seção em `docs/manual-usuario.md` explicando os códigos 6 (`DISPOSITIVO_INDISPONIVEL`/`PIN_INVALIDO`) quando o usuário usar provider PKCS#11.
12. Atualizar `CLAUDE.md` removendo a nota "PKCS11SignatureService: esqueleto".

**Critérios de aceitação (espelhados aos 4 da US-02.5):**
- [ ] **(SunPKCS11)** `PKCS11SignatureService` assina e valida usando provider real.
- [ ] **(SoftHSM2)** Teste de integração `@Tag("pkcs11")` passa localmente com SoftHSM2 + chave RSA-2048.
- [ ] **(mensagem clara)** Os 3 cenários de erro retornam `RespostaErro` JSON estruturada com código + mensagem (exemplo: `{"error":"DISPOSITIVO_INDISPONIVEL","message":"Driver PKCS#11 nao pode ser carregado: /tmp/inexistente.so"}`).
- [ ] **(documentação)** `docs/tecnico.md` § 5.2 cobre setup nos 3 SOs com comandos copiáveis.
- [ ] CI continua verde (tag `pkcs11` é skipped por default).
- [ ] Enums Java e Go alinhados (sem mais "reservado" em `cmd/exit.go`).

---

## 6.12 — Documento de requisitos não-funcionais

**Origem:** `runnerProfessor/docs/planejamento.md` — crítica explícita à ausência de NFRs, sugestão de uso da ISO 25010 e IEEE 29148:2018.

**Estado atual:** nenhum documento de NFRs no projeto.

**Plano de ação:**
1. Criar `planejamento/requisitos-nao-funcionais.md` cobrindo as características da ISO 25010:
   - **Performance:** cold start (modo local) < 3s em hw razoável; warm start (modo HTTP) < 200ms p95.
   - **Portabilidade:** Windows/Linux/macOS amd64; sem dependência manual de JDK/JRE; sem instalação root.
   - **Segurança:** artefatos assinados via Cosign (Sigstore OIDC); chaves nunca em código; PKCS#11 isola chave privada no dispositivo.
   - **Confiabilidade:** auto-detecção de porta livre; recuperação de `state.json` corrompido; PID stale tratado.
   - **Manutenibilidade:** cobertura > 80%; CI matrix 3 SOs; logging estruturado.
   - **Observabilidade:** `slog` + OTel bridge (Go), `Logstash` + OTel appender (Java).
   - **Usabilidade:** `--help` em todos comandos; mensagens de erro com código + sugestão.
2. Cada NFR no formato SMART quando possível (métrica + alvo + horizonte).
3. Linkar do README e CLAUDE.md.

**Critérios de aceitação:**
- [ ] Documento criado com pelo menos 7 características da ISO 25010.
- [ ] Cada NFR tem critério verificável.
- [ ] README aponta para o documento.

---

## 6.13 — Estratégia `release.json` para descoberta do simulador.jar

**Origem:** `especificacao.md` US-03, exemplo de estratégia (não obrigatória).
> "Busca release.json (URL fixa e conhecida) no próprio repositório (branch main) via url estável `https://raw.githubusercontent.com/{owner}/{repo}/main/release.json`."

**Estado atual:** projeto usa a API `api.github.com/repos/.../releases` diretamente.

**Plano de ação (avaliação primeiro):**
1. Avaliar trade-offs:
   - **Pró release.json:** uma única URL estável; não depende de paginação ou rate limit da GitHub API; permite indicar JRE específico por SO.
   - **Pró API atual:** descoberta automática da última versão; metadata SHA-256 nativa.
2. **Decisão recomendada:** suportar **ambos**, com `release.json` como primário e fallback para API.
3. Em `simulador/internal/download/`, adicionar `buscarViaReleaseJson(url)` → `Metadados{Jar, Versao, JreURLs}`.
4. URL configurável via env `HUBSAUDE_RELEASE_JSON_URL`, default `https://raw.githubusercontent.com/hubsaude/simulador/main/release.json`.
5. Se falhar (404, erro de parse), cai no fluxo atual da GitHub Releases API.
6. Estender `--source` para aceitar caminho a um `release.json` local (`file://...`).

**Critérios de aceitação:**
- [ ] Teste unitário com `httptest` servindo `release.json` válido.
- [ ] Fallback testado (release.json 404 → API tradicional).
- [ ] Documento de decisão em `planejamento/decisoes-tecnicas.md`.

---

## 6.14 — Atualizar README com diagramas, NFRs e nota de escopo

**Estado atual:** README atual não inclui diagramas C4, não linka NFRs, não esclarece escopo do simulador.

**Plano de ação:**
1. Adicionar seção "Arquitetura" no README com diagramas embedados (SVG de 6.13).
2. Adicionar seção "Escopo" esclarecendo que o `simulador.jar` é dependência externa (referenciando decisão de 6.14).
3. Adicionar links na seção "Documentação":
   - `planejamento/requisitos-nao-funcionais.md` (6.16)
   - `docs/pendencias-spec.md` (este doc)
4. Conferir que comandos documentados (`criar`, `validar`, `iniciar`, `parar`, `status`, `servidor iniciar --timeout`) batem com flags reais.

**Critérios de aceitação:**
- [ ] README exibe SVGs.
- [ ] Linka NFRs e decisão de escopo.
- [ ] Todos os exemplos de comando são executáveis sem ajuste.

---

## Ordem sugerida de execução

| Ordem | Item | Justificativa |
|------:|------|---------------|
| 1 | 6.5 (porta 8443) | Mudança pontual, libera testes downstream |
| 2 | 6.7 (JRE) | Decisão atômica que cascateia em docs |
| 3 | 6.6 (`/api/info`) | Pequeno e isolado |
| 4 | 6.8 (`--timeout`) | Envolve Java + Go, mas é localizado |
| 5 | 6.10 (decisão escopo simulador) | Documental, destrava 6.14 |
| 6 | 6.9 (diagramas C4) | Documental, destrava 6.14 |
| 7 | 6.12 (NFRs) | Documental, destrava 6.14 |
| 8 | 6.11 (PKCS#11/SoftHSM2) | Maior esforço; pode ser opt-in |
| 9 | 6.13 (release.json) | Opcional; só se sobrar tempo |
| 10 | 6.14 (README) | Consolida tudo |
| 11 | 6.16, 6.17 (erros + limpeza) | Pré-release |
| 12 | 6.15 (cobertura) | Após todas as mudanças de código |
| 13 | 6.18 (release `v1.0.0`) | Última coisa |

---

## Checklist final de conformidade com a spec

Antes de fechar a tag `v1.0.0`, validar **cada critério** dos documentos do professor — agrupados por US:

### US-01 — Invocar assinador.jar via CLI (10 critérios)
- [x] CLI aceita comandos para criação e validação
- [x] CLI invoca o assinador.jar com parâmetros
- [x] CLI permite invocação direta (modo local/CLI)
- [x] CLI permite invocação via HTTP (modo servidor)
- [x] CLI exibe resultado de forma legível
- [x] CLI inicia assinador.jar no modo servidor usando porta padrão
- [x] CLI detecta instância existente e a reutiliza
- [x] CLI usa modo servidor por padrão (`--modo` default "http")
- [x] CLI interrompe execução em porta padrão ou indicada
- [x] **CLI permite interrupção programada após N minutos sem interação** (item 6.8 — `--timeout`)

### US-02 — Simular assinatura digital (5 critérios)
- [x] Valida todos os parâmetros FHIR
- [x] Simula criação de assinatura
- [x] Simula validação de assinatura
- [x] **Suporta interação com dispositivo criptográfico via PKCS#11** (item 6.11 — `PKCS11SignatureService` funcional; testes opt-in com SoftHSM2)
- [x] Retorna mensagens de erro claras para parâmetros inválidos

### US-03 — Gerenciar ciclo de vida do Simulador (7 critérios)
- [x] **Verifica porta padrão 8443 antes de iniciar** (item 6.5)
- [x] Permite iniciar o Simulador
- [x] Permite parar o Simulador (endpoint `/shutdown`)
- [x] **Exibe status via endpoint `/api/info`** (item 6.6)
- [x] simulador.jar é baixado dinamicamente do GitHub Releases (+ release.json opcional — item 6.13)
- [x] **Baixa o JRE (não o JDK) caso não disponível em `.hubsaude`** (item 6.7 — `image_type=jre`)
- [x] Não baixa simulador.jar se já estiver disponível localmente

### US-04 — Provisionar JDK/JRE (4 critérios)
- [x] Detecta JDK/JRE presente na versão exigida
- [x] **Baixa JRE compatível quando ausente** (item 6.7)
- [x] Disponibiliza para uso do Assinador e Simulador
- [x] Funciona nas três plataformas

### US-05 — Binários multiplataforma (6 critérios)
- [x] Binário Windows (amd64)
- [x] Binário Linux (amd64)
- [x] Binário macOS (amd64)
- [x] Distribuído via GitHub Releases
- [x] Checksums SHA256 para integridade
- [x] Versionamento semântico (SemVer)

### Entregáveis (§7 da especificação)
- [x] Código-fonte da aplicação assinatura
- [x] Código-fonte da aplicação assinador.jar
- [x] Testes (unit, integração, cenários de erro, aceitação)
- [x] Documentação (manual, técnica, exemplos, instalação)
- [x] **Especificação com diagramas C4** (item 6.9 — `diagramas/contexto.puml` e `conteineres.puml`; SVGs gerados via `geraimagens.sh`)
- [x] Artefatos executáveis (.exe / .AppImage / .dmg + Cosign)
- [x] **Decisão formal sobre código-fonte do Simulador HubSaúde** (item 6.10 — DT em `decisoes-tecnicas.md`)

### Cosign / Sigstore (§9)
- [x] Todos artefatos assinados via Cosign
- [x] Identidade baseada em OIDC + transparency log
- [x] Arquivos `<artefato>.sig` + `<artefato>.pem` na release
- [x] Processo de assinatura automatizado no CI/CD
- [ ] Validar manualmente que `cosign verify-blob` retorna "Verified OK"

### Críticas do `planejamento.md` do professor
- [x] **Requisitos não-funcionais documentados** (ISO 25010) — `planejamento/requisitos-nao-funcionais.md` com 7 categorias (item 6.12)
- [ ] **DoR/DoD por requisito** (avaliar incluir tabela em `BACKLOG.md`)

### Higiene do repositório
- [x] `target/` ignorado e não rastreado
- [x] `bin/`, `*.log`, JSONs de teste manual ignorados
- [ ] Limpar/remover arquivos órfãos remanescentes antes do tag

### Release final
- [ ] CI verde nas 3 plataformas (build + test + acceptance)
- [ ] Tag `v1.0.0` criada → `release.yml` executa e publica → item 6.18

> **Quando todos os checkboxes estiverem [x], o projeto cobre 100% da especificação do professor.**
