# Documentação Técnica — Sistema Runner

Público: desenvolvedores e integradores. Descreve a arquitetura, os contratos de
comunicação entre os componentes, o design da assinatura (`SignatureService` / PKCS#11)
e os mecanismos de provisionamento (JRE e `simulador.jar`).

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

### 5.1. PKCS#11 (implementação SunPKCS11)

`PKCS11SignatureService` implementa a interação com dispositivos criptográficos
via **SunPKCS11** (provider embutido no JDK). Diferença conceitual chave:

> No modo PKCS#11 a **chave privada não é acessível como parâmetro** — ela permanece no
> dispositivo. Os parâmetros `privateKey`/`publicKey` da interface passam a ser apenas
> **aliases** dentro do `KeyStore` PKCS#11; a autenticação é feita por **PIN**.

Fluxo de assinatura (`sign`):

```java
Provider p = Security.getProvider("SunPKCS11").configure(pkcs11Cfg);
Security.addProvider(p);
KeyStore ks = KeyStore.getInstance("PKCS11", p);
ks.load(null, pin);
PrivateKey k = (PrivateKey) ks.getKey(alias, null);
Signature sig = Signature.getInstance("SHA256withRSA");
sig.initSign(k);
sig.update(payload);
byte[] assinatura = sig.sign(); // operacao ocorre DENTRO do dispositivo
```

#### Tratamento de erros (alinhado com `cmd/exit.go` no CLI Go)

| Cenário | Código | HTTP | Exit code (Go) |
|---------|--------|------|----------------|
| Driver PKCS#11 não carrega (`ProviderException`, `InvalidParameterException`, config inexistente) | `DISPOSITIVO_INDISPONIVEL` | 503 | 6 |
| PIN incorreto (`FailedLoginException`) | `PIN_INVALIDO` | 401 | 6 |
| Alias inexistente no KeyStore | `PARAM_INVALIDO` | 400 | 3 |
| Base64 inválido em `validate` | `PARAM_INVALIDO` | 400 | 3 |

### 5.2. Setup do SoftHSM2 para desenvolvimento

Para validar a integração PKCS#11 sem token físico, o projeto suporta o **SoftHSM2** —
implementação software de PKCS#11 distribuída pelo OpenDNSSEC.

#### Linux (Debian/Ubuntu)

```bash
sudo apt-get install -y softhsm2 opensc
# inicializa um token no slot 0 (uso unico)
softhsm2-util --init-token --slot 0 --label "test" --pin 1234 --so-pin 1234
# importa um par RSA-2048 com alias "testkey"
pkcs11-tool --module /usr/lib/softhsm/libsofthsm2.so \
    --login --pin 1234 \
    --keypairgen --key-type RSA:2048 --label testkey
# exporta caminho da biblioteca
export SOFTHSM2_LIB=/usr/lib/softhsm/libsofthsm2.so
```

#### macOS

```bash
brew install softhsm opensc
softhsm2-util --init-token --slot 0 --label "test" --pin 1234 --so-pin 1234
pkcs11-tool --module $(brew --prefix softhsm)/lib/softhsm/libsofthsm2.so \
    --login --pin 1234 \
    --keypairgen --key-type RSA:2048 --label testkey
export SOFTHSM2_LIB=$(brew --prefix softhsm)/lib/softhsm/libsofthsm2.so
```

#### Windows

1. Baixar SoftHSM2 do release oficial: <https://github.com/disig/SoftHSM2-for-Windows>
2. Adicionar `C:\SoftHSM2\bin` ao `PATH`.
3. Em um PowerShell de administrador:
   ```powershell
   softhsm2-util --init-token --slot 0 --label "test" --pin 1234 --so-pin 1234
   pkcs11-tool --module "C:\SoftHSM2\lib\softhsm2-x64.dll" `
       --login --pin 1234 `
       --keypairgen --key-type RSA:2048 --label testkey
   $env:SOFTHSM2_LIB = "C:\SoftHSM2\lib\softhsm2-x64.dll"
   ```

#### Exemplo de `pkcs11.cfg` para SunPKCS11

```
name = SoftHSM2
library = /usr/lib/softhsm/libsofthsm2.so
slot = 0
```

#### Rodar os testes de integração

Por padrão a tag `pkcs11` é excluída pelo Surefire. Para executar:

```bash
export SOFTHSM2_LIB=...   # caminho do .so/.dylib/.dll
export SOFTHSM2_PIN=1234
export SOFTHSM2_ALIAS=testkey
cd assinador && ./mvnw test -Dgroups=pkcs11
```

Os 3 cenários cobertos em `PKCS11IntegrationTest`:

1. Assinar e validar com token (caminho feliz).
2. PIN errado → `PIN_INVALIDO`.
3. Alias inexistente → `PARAM_INVALIDO`.

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
  "simulador": { "portaPadrao": 8443, "sourceUrl": "" },
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
