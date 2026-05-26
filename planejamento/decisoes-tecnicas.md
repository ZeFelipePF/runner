# Decisoes Tecnicas

Decisoes tomadas para o projeto. Alternativas sao mantidas como referencia.

---

## Linguagem do CLI

**Decisao: Go**
**Justificativa:** Cross-compilation nativa (`GOOS`/`GOARCH`) sem dependencias externas, binario estatico unico, ecossistema CLI maduro (`cobra`), biblioteca padrao suficiente para HTTP e execucao de processos. Padrao de mercado para ferramentas CLI (Docker, kubectl, gh).

| Aspecto | Detalhe |
|---------|---------|
| Cross-compilation | `GOOS`/`GOARCH` — nativa, sem ferramentas extras |
| CLI parsing | `cobra` |
| HTTP client | `net/http` (stdlib) |
| Execucao de processos | `os/exec` (stdlib) |
| Binario | Estatico, ~10MB, zero dependencias de runtime |
| Empacotamento | Binario direto; tooling extra para .AppImage/.dmg |

<details>
<summary>Alternativas consideradas</summary>

**Rust:** binario menor e seguranca de memoria, mas curva de aprendizado alta e cross-compilation mais trabalhosa.

**Python:** desenvolvimento rapido, mas binario pesado (~50MB), empacotamento multiplataforma fragil, sem binario nativo real.
</details>

---

## Build System Java

**Decisao: Maven com Maven Wrapper (`mvnw`)**
**Justificativa:** Padrao da industria com ampla documentacao, convencao sobre configuracao reduz decisoes desnecessarias, `mvnw` garante reproducibilidade sem Maven instalado na maquina do desenvolvedor ou CI.

<details>
<summary>Alternativas consideradas</summary>

**Gradle:** mais flexivel e cache incremental mais rapido, mas DSL mais complexo e overhead desnecessario para este escopo.

**Maven sem wrapper:** menos reproducivel, depende de instalacao local do Maven.
</details>

---

## Framework HTTP Java

**Decisao: Javalin**
**Justificativa:** Micro-framework leve (~1MB no jar final), API simples e moderna, minimo de boilerplate — ideal para uma API pequena com dois endpoints (`/sign`, `/validate`). Evita o overhead de Spring Boot (~30MB) desnecessario para este escopo.

Dependencia Maven:
```xml
<dependency>
    <groupId>io.javalin</groupId>
    <artifactId>javalin</artifactId>
    <version>6.x.x</version>
</dependency>
```

<details>
<summary>Alternativas consideradas</summary>

**Spring Boot:** framework completo, otimo se a equipe ja conhece, mas pesado (~30MB) para uma API com dois endpoints.

**com.sun.net.httpserver:** zero dependencias externas, mas API de baixo nivel sem roteamento, exige muito boilerplate.
</details>

---

## Versao do runtime Java

**Decisao: JDK 21 para build, JRE 21 para provisionamento automatico — Eclipse Temurin (Adoptium)**
**Justificativa:** Long-Term Support ate 2029, virtual threads, pattern matching e records disponiveis. Temurin e a distribuicao de referencia da comunidade, com API REST para download programatico.

- `source` e `target` do Maven: `21` (build do `assinador.jar` exige JDK)
- **Provisionamento automatico (US-03 e US-04 da spec): JRE 21** (`image_type=jre` na Adoptium API). A spec exige explicitamente "baixar o JRE" — JRE e mais leve (~50MB vs ~180MB do JDK), suficiente para executar `assinador.jar` e `simulador.jar`.
- Diretorio destino: `~/.hubsaude/jdk/` (nome mantido por compatibilidade com layouts existentes; conteudo e JRE).
- Deteccao (`Detectar`) procura apenas `java` no PATH/JAVA_HOME/hubsaude — funciona com JDK ou JRE.

<details>
<summary>Alternativas consideradas</summary>

**JDK 17 LTS:** mais conservador, sem virtual threads. LTS ate 2027.

**JDK 23+:** nao e LTS, suporte curto — inadequado para projeto com vida util de anos.
</details>

---

## Formato de Comunicacao CLI <-> assinador.jar

**Decisao: JSON**
**Justificativa:** Padrao universal para APIs REST, alinhado com FHIR (que usa JSON), parsing nativo disponivel em Go (`encoding/json`) e Java (`jackson` ou `gson`), facil de debugar com `curl` e logs.

- Modo CLI: saida JSON em stdout, erros JSON em stderr
- Modo HTTP: `Content-Type: application/json` em requests e responses

<details>
<summary>Alternativas consideradas</summary>

**YAML:** mais legivel para humanos, mas nao e padrao para APIs e parsing mais complexo.

**Protocol Buffers:** compacto e tipado, mas overhead de definicao de schemas `.proto` desnecessario para este escopo.
</details>

---

## Porta Padrao do Modo Servidor

**Decisao: `8088` (assinador) — porta configuravel, com auto-deteccao se ocupada**

- Porta padrao do assinador.jar: `8088`
- Se ocupada: CLI tenta `8089`, `8090`, ... ate encontrar livre (range maximo: +20)
- Configuravel via `--porta` ou `config.json`
- Porta efetiva gravada em `~/.hubsaude/state.json` apos inicializacao

---

## Escopo do Simulador HubSaude (simulador.jar)

**Decisao: Simulador HubSaude tratado como dependencia externa, nao desenvolvido neste TP**

A especificacao (§7 entregavel 7) lista "Codigo fonte do Simulador do HubSaude" como entregavel. Apos analise do escopo e contexto:

- **Contexto:** o `simulador.jar` e parte do ecossistema HubSaude, mantido pela equipe da plataforma (Secretaria de Estado de Saude de Goias / UFG). Replicar seu codigo no escopo deste TP seria inviavel (sem acesso aos modelos FHIR completos) e nao traria valor pedagogico, ja que o foco do trabalho e a **integracao** (CLI <-> jar) e nao o conteudo do simulador.
- **Alternativa avaliada:** criar stub minimo Javalin com endpoints `/actuator/health` e `/api/info` para desenvolvimento local. Descartada por nao adicionar valor — os testes existentes ja cobrem o gerenciamento de processo via mocks (`httptest`).
- **Decisao final:** o `simulador.jar` e baixado dinamicamente da GitHub Releases da disciplina (US-03 da spec). O CLI `simulador` cobre integralmente o ciclo de vida (download, cache, iniciar, parar, status). Esta abordagem alinha-se com a propria expectativa da spec ("obtido dinamicamente pelo CLI, baixando a versao mais recente disponivel no repositorio da disciplina").
- **Onde a decisao esta refletida:** README seccao "Escopo"; `CLAUDE.md` topo do arquivo.

---

## Descoberta do simulador.jar — release.json + fallback GitHub API

**Decisao: Suportar ambas as estrategias, com `release.json` como primaria opcional**

A especificacao US-03 sugere (mas nao obriga) a estrategia de buscar um
`release.json` em URL fixa do repositorio da disciplina:

```
https://raw.githubusercontent.com/{owner}/{repo}/main/release.json
```

- **Pro release.json:** URL estavel, sem paginacao/rate limit da GitHub API, permite anotar URLs especificas de JRE por SO.
- **Pro GitHub API atual:** descoberta automatica da ultima release.

Decisao final implementada em `simulador/internal/download/release_json.go`:

1. Se `Opcoes.ReleaseJsonURL` (CLI flag) ou env `HUBSAUDE_RELEASE_JSON_URL` definida → tenta `release.json` primeiro.
2. Em sucesso, baixa o `jar.url` indicado.
3. Em falha (404, JSON invalido, sem `jar.url`), cai automaticamente na GitHub Releases API tradicional.
4. `--source` continua sobrescrevendo tudo (URL direta).

Formato esperado:

```json
{
  "jar": {
    "url": "https://github.com/.../assinador.jar",
    "version": "1.2.0",
    "sha256": "<opcional>"
  },
  "jre": {
    "linux_x64": "...",
    "windows_x64": "...",
    "mac_x64": "..."
  }
}
```
