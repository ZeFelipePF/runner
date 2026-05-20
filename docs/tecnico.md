# Documentação Técnica — Sistema Runner

Público: desenvolvedores e integradores. Descreve a arquitetura, os contratos de
comunicação entre os componentes, o design da assinatura (`SignatureService` / PKCS#11)
e os mecanismos de provisionamento (JDK e `simulador.jar`).

> Para uso prático, veja o [Manual do Usuário](manual-usuario.md).
> O contrato FHIR completo investigado está em [`planejamento/contrato-fhir.md`](../planejamento/contrato-fhir.md).

---

## 1. Visão de componentes

```
┌──────────────┐        os/exec (modo local)        ┌────────────────────┐
│  assinatura  │ ─────────────────────────────────► │   assinador.jar    │
│   (CLI Go)   │        net/http (modo http)         │   (Java/Javalin)   │
└──────┬───────┘ ─────────────────────────────────► └─────────┬──────────┘
       │                                                       │
       │ lê/escreve                              SunPKCS11 ────┘ (esqueleto)
       ▼                                              ▼
┌──────────────┐                              driver nativo do token/smartcard
│ ~/.hubsaude/ │  state.json · config.json · jdk/ · simulador/
└──────▲───────┘
       │ lê/escreve
┌──────┴───────┐        os/exec                ┌────────────────────┐
│  simulador   │ ───────────────────────────► │   simulador.jar    │
│   (CLI Go)   │                              │   (HubSaúde)        │
└──────────────┘                              └────────────────────┘
```

| Componente | Linguagem | Responsabilidade |
|------------|-----------|------------------|
| `assinatura` | Go (cobra) | Traduzir comandos do usuário em chamadas ao `assinador.jar`. |
| `assinador.jar` | Java 21 (Javalin) | Validar parâmetros FHIR e simular assinatura/validação. |
| `simulador` | Go (cobra) | Gerenciar o ciclo de vida do `simulador.jar` (download, start, stop, status). |
| `simulador.jar` | Java (externo) | Aplicação do HubSaúde, **não desenvolvida** aqui; baixada via GitHub Releases. |

O `~/.hubsaude/` é o "banco de dados" local compartilhado entre os CLIs.

---

## 2. Fluxo de dados

### 2.1. Criar assinatura (modo HTTP)

```
usuário ─► assinatura criar --payload p.json
            │
            ├─ 1. LerPayload(p.json)                      (lê bytes do arquivo/stdin)
            ├─ 2. LocalizarJar()                          (env → ~/.hubsaude → cwd → target)
            ├─ 3. Garantir(startup):
            │        state.json + PID + GET /health  ──► reusa instância saudável
            │                       senão            ──► escolhe porta livre e sobe a JVM
            ├─ 4. POST /sign  (corpo = payload JSON)
            │        assinador.jar:
            │          ValidadorFHIR.validarCriacao(payload)
            │          AcaoAssinar → SignatureService.sign(msg, chave)
            │          monta recurso Signature (FHIR R4)
            └─ 5. imprime Signature no stdout (JSON formatado)
```

No **modo local** os passos 3–4 viram um `os/exec` de
`java -jar assinador.jar sign --input -` com o payload no stdin; a saída JSON do processo
é repassada.

### 2.2. Validar assinatura

Idêntico, trocando `criar`→`validar`, `/sign`→`/validate`, `validarCriacao`→`validarVerificacao`
e `AcaoAssinar`→`AcaoValidar` (que monta um `OperationOutcome`).

---

## 3. Contrato HTTP do `assinador.jar`

Servidor Javalin, porta padrão **8088** (auto-detecta a próxima livre, janela de +20,
se ocupada). Modo iniciado por `java -jar assinador.jar server [--porta N]`.

| Endpoint | Método | Corpo | Descrição |
|----------|--------|-------|-----------|
| `/sign` | POST | payload FHIR de criação (JSON) | Cria assinatura simulada. |
| `/validate` | POST | payload FHIR de validação (JSON) | Valida assinatura. |
| `/health` | GET | — | Health check. |
| `/shutdown` | POST | — | Encerra o servidor (assíncrono, exit 0). |

Header de resposta: `Content-Type: application/json`.

### 3.1. `POST /sign`

**Request** (campos obrigatórios validados por `ValidadorFHIR.validarCriacao`):

| Campo | Tipo | Regra |
|-------|------|-------|
| `bundle` | string | não vazio |
| `provenance` | string | não vazio |
| `materialCriptografico` | object | não vazio (subtipos: `PEM`, `PKCS12`, `SMARTCARD`/`TOKEN`, `REMOTE`) |
| `certificados` | array[string] | não vazio; cada item base64 válido |
| `timestampReferencia` | integer | número (Unix UTC) |
| `estrategiaTimestamp` | string | `iat` ou `tsa` |
| `politicaAssinatura` | string | não vazio (URI da política) |
| `algoritmo` | string | opcional; se presente, `RS256` ou `ES256` |

**Response 200** — recurso `Signature` (FHIR R4):

```json
{
  "resourceType": "Signature",
  "type": [ { "system": "urn:iso-astm:E1762-95:2013", "code": "1.2.840.10065.1.12.1.1" } ],
  "when": "<ISO-8601 derivado de timestampReferencia>",
  "sigFormat": "application/jose",
  "targetFormat": "application/octet-stream",
  "data": "QVNTSU5BVFVSQS1TSU1VTEFEQS12MQ=="
}
```

O `data` é a constante `FakeSignatureService.ASSINATURA_SIMULADA` — base64 de
`"ASSINATURA-SIMULADA-v1"`. A mensagem canônica assinada é
`base64(bundle + "|" + provenance)`.

### 3.2. `POST /validate`

**Request** (validado por `ValidadorFHIR.validarVerificacao`):

| Campo | Tipo | Regra |
|-------|------|-------|
| `jws` | string | não vazio e base64 válido |
| `trustStore` | array[string] | não vazio; cada item hash SHA-256 hex (64 chars) |
| `minCertIssueDate` | integer | número (Unix UTC) |
| `referenceTimestamp` | integer | número (Unix UTC) |
| `signaturePolicyId` | string | não vazio |

**Response 200** — recurso `OperationOutcome`. `severity: information` quando o `jws`
corresponde ao valor simulado; `severity: error` (code `VALIDATION.SIGNATURE-VERIFICATION-FAILED`)
caso contrário.

### 3.3. `GET /health`

```json
{ "status": "UP", "iniciadoEm": "<ISO-8601>", "uptimeSegundos": 12 }
```

### 3.4. Erros

Qualquer falha vira `AssinadorException` → JSON `RespostaErro`:

```json
{ "error": "PARAM_AUSENTE", "message": "campo 'bundle' obrigatorio" }
```

`MapeadorErro` define o status HTTP por código (`Codigo` de `AssinadorException`):

| Código (`error`) | HTTP | Exit code (modo CLI) |
|------------------|------|----------------------|
| `PARAM_AUSENTE` | 400 | 2 |
| `PARAM_INVALIDO` | 400 | 3 |
| `ALGORITMO_NAO_SUPORTADO` | 400 | 4 |
| `PAYLOAD_MUITO_GRANDE` | 413 | 5 |
| `ERRO_INTERNO` | 500 | 1 |

---

## 4. Contrato CLI do `assinador.jar`

Modo padrão quando o primeiro argumento **não** é `server`.

```
java -jar assinador.jar <sign|validate> [--input <arquivo>]
```

- Entrada: `--input <arquivo>` ou, na ausência (ou com `-`), o **stdin**.
- Saída de sucesso: JSON no **stdout**, exit `0`.
- Saída de erro: JSON `{"error","message"}` no **stderr**, exit conforme tabela acima.

Exemplo:

```bash
echo '<payload>' | java -jar assinador.jar sign --input -
```

> Atenção: o CLI Go (`assinatura`) e o `assinador.jar` usam **enums de código
> distintos**. O CLI Go também reconhece `ASSINATURA_INVALIDA` (exit 5),
> `DISPOSITIVO_INDISPONIVEL`/`PIN_INVALIDO` (exit 6) e mapeia `ERRO_INTERNO` para 7,
> reservados para a evolução do PKCS#11. O `assinador.jar` em modo simulado emite
> apenas os cinco códigos da seção 3.4.

---

## 5. Design da assinatura (Java)

```
SignatureService (interface)
 ├─ FakeSignatureService     ← implementação principal entregue (simulação)
 └─ PKCS11SignatureService   ← esqueleto (token/smartcard físico indisponível)

SignatureController          ← Javalin, expõe /sign /validate /health /shutdown
AcaoAssinar / AcaoValidar    ← orquestram validação + serviço + montagem FHIR
ValidadorFHIR                ← validação de parâmetros (sem criptografia real)
MapeadorErro / RespostaErro  ← código de domínio → HTTP status / exit code / JSON
```

Princípios:
- A interface `SignatureService` é central; implementações são **injetadas** no controller
  (sem `new` direto). Isso permite trocar `Fake` por `PKCS11` sem alterar o controller.
- `FakeSignatureService.sign(message, privateKey)` valida presença e base64 e devolve a
  constante simulada. `validate(message, signature, publicKey)` compara com essa constante.

### 5.1. PKCS#11 (esqueleto)

`PKCS11SignatureService` documenta a integração prevista com dispositivos criptográficos
via **SunPKCS11** (provider embutido no JDK). Diferença conceitual chave:

> No modo PKCS#11 a **chave privada não é acessível como parâmetro** — ela permanece no
> dispositivo. Os parâmetros `privateKey`/`publicKey` da interface passam a ser apenas
> **aliases** dentro do `KeyStore` PKCS#11; a autenticação é feita por **PIN**.

Fluxo previsto de inicialização (documentado no código):

```java
Provider p = Security.getProvider("SunPKCS11").configure(configPath);
Security.addProvider(p);
KeyStore ks = KeyStore.getInstance("PKCS11", p);
ks.load(null, pin);
```

Sem token físico + driver nativo, os métodos lançam `AssinadorException`. O escopo de
simulação do projeto é coberto integralmente por `FakeSignatureService`.

---

## 6. Startup inteligente dos CLIs

Tanto `assinatura` (`internal/assinador.Garantir`) quanto `simulador`
(`internal/processo.Iniciar`) seguem o mesmo princípio de **reuso**:

```
1. Ler state.json
2. Há processo registrado?  ── não ──► iniciar novo (passo 4)
        │ sim
3. PID vivo?  ── não (stale) ──► limpar registro e iniciar novo
        │ sim
   GET /health responde?  ── sim ──► REUSAR (não sobe nada)
                          └─ não ──► encerrar/limpar e iniciar novo
4. Escolher porta livre (preferida → +20 janela)
5. Subir a JVM, aguardar a porta responder, gravar state.json
```

A localização do `assinador.jar` (`LocalizarJar`) segue a ordem:
`HUBSAUDE_ASSINADOR_JAR` → `~/.hubsaude/assinador/` → diretório atual → `assinador/target/`.

---

## 7. Provisionamento do JDK (pacote `internal/jdk`)

Duplicado em `assinatura` e `simulador` (decisão consciente — ver `CLAUDE.md`).

**Detecção** (`Detectar`), nesta ordem:
1. `JAVA_HOME`
2. `java` no `PATH`
3. `~/.hubsaude/jdk/*/bin/java`

`parseMajor` entende `21.0.3`, `1.8.0_421` (legado) e `22-ea`. Exige major ≥ 21.

**Provisionamento** (`Garantir`), quando ausente:
- Consulta a **Adoptium API** (`api.adoptium.net/v3`; sobrescrevível por `HUBSAUDE_ADOPTIUM_URL`).
- Resolve SO × arquitetura: `linux`/`mac`/`windows` × `amd64`/`arm64`.
- Baixa e extrai `.tar.gz` (Unix) ou `.zip` (Windows).
- **Valida o SHA-256** do download contra os metadados.
- Persiste `hubsaude-meta.json` e devolve o caminho do `bin/java`.

---

## 8. Download do `simulador.jar` (pacote `internal/download`)

- Cache em `~/.hubsaude/simulador/{simulador.jar, simulador-meta.json}`.
- Fonte padrão: **GitHub Releases**
  (`https://api.github.com/repos/hubsaude/simulador/releases`; sobrescrevível por
  `HUBSAUDE_SIMULADOR_REPO`).
- `--versao-simulador`: `latest` (padrão) ou uma tag específica.
- `--source <url>`: aceita `http(s)://` e `file://`. Nesse caso a versão é `"custom"` e o
  GitHub Releases é ignorado.
- A cada chamada o **SHA-256** do cache é revalidado contra os metadados (detecta corrupção).
- `ForcarRedownload` ignora o cache.

---

## 9. Estado local (`~/.hubsaude/`)

`config.json` (configurações do usuário, com defaults):

```json
{
  "assinador": { "portaPadrao": 8088, "timeoutShutdownSegundos": 0 },
  "simulador": { "portaPadrao": 9090, "sourceUrl": "" },
  "jdk":       { "versaoMinima": 21, "distribuicao": "" }
}
```

`state.json` (processos em execução; cada entrada é opcional):

```json
{
  "assinador": { "pid": 0, "porta": 0, "iniciadoEm": "<ISO>", "javaPath": "", "versao": "" },
  "simulador": { "pid": 0, "porta": 0, "iniciadoEm": "<ISO>", "javaPath": "", "versao": "" }
}
```

A checagem de PID vivo usa `windows.OpenProcess` (Windows) ou `Signal(0)` (Unix).
Entradas com PID morto são consideradas _stale_ e limpas.

---

## 10. Distribuição e integridade

- **Build**: cross-compilation Go (Windows/Linux/macOS × amd64/arm64) via goreleaser.
- **Empacotamento**: `.exe` (Windows), `.AppImage` (Linux), `.dmg` (macOS); `assinador.jar`
  incluído como artefato.
- **Assinatura**: Cosign (Sigstore) keyless via OIDC no `release.yml` — gera `.sig` + `.pem`
  por artefato e registra no _transparency log_.
- **Integridade**: `checksums-sha256.txt` agregado.

A verificação pelo usuário está descrita no [Guia de Instalação](instalacao.md#4-verificação-de-autenticidade-cosign).
