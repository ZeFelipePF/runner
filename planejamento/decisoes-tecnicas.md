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

## Versao JDK

**Decisao: JDK 21 (LTS) — Eclipse Temurin (Adoptium)**
**Justificativa:** Long-Term Support ate 2029, virtual threads, pattern matching e records disponiveis. Temurin e a distribuicao de referencia da comunidade, com API REST para download programatico (usada pelo provisionamento automatico).

- `source` e `target` do Maven: `21`
- Distribuicao para download automatico: Eclipse Temurin via `api.adoptium.net`

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
