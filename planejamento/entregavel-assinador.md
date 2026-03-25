# Entregavel: assinador.jar (US-02)

## Objetivo

Aplicacao Java que valida rigorosamente parametros de entrada e simula operacoes de assinatura digital (criacao e validacao), retornando respostas pre-construidas.

## Historias Relacionadas

- **US-02**: Simular assinatura digital com validacao de parametros
- **US-01**: Modos de invocacao (CLI e HTTP)

---

## Design da Camada de Servico

### Interface SignatureService

O professor define a interface central da aplicacao:

```java
public interface SignatureService {
    String sign(String message, String privateKey);
    String validate(String message, String signature, String publicKey);
}
```

**Atencao:** Esta assinatura de metodo e um "rumo", nao um contrato fixo. Ha uma limitacao importante no modo PKCS#11: quando se usa dispositivo criptografico (token/smartcard), a chave privada **nao e acessivel como parametro** — ela nunca sai do dispositivo. O design deve acomodar isso.

### FakeSignatureService (simulacao)

A implementacao de simulacao deve ser uma classe separada que implementa `SignatureService`:

```java
public class FakeSignatureService implements SignatureService {
    @Override
    public String sign(String message, String privateKey) {
        // Retorna assinatura simulada pre-construida
    }

    @Override
    public String validate(String message, String signature, String publicKey) {
        // Retorna resultado pre-determinado
    }
}
```

Esta e a implementacao usada em toda a aplicacao para simulacao.

### PKCS11SignatureService (integracao com dispositivo)

Para o modo com dispositivo criptografico, ha uma segunda implementacao:

```java
public class PKCS11SignatureService implements SignatureService {
    // Usa SunPKCS11 para delegar operacoes ao dispositivo fisico
    // Nao recebe privateKey como parametro — a chave fica no token
}
```

---

## Integracao com Dispositivo Criptografico (SunPKCS11)

### Contexto

O dispositivo fisico (USB token ou smartcard com certificado ICP-Brasil tipo A3) nao esta disponivel para testes, mas o **driver nativo do fabricante** (biblioteca `.so` no Linux, `.dll` no Windows, `.dylib` no macOS) esta.

### Analogia com JDBC

| JDBC | PKCS#11 |
|------|---------|
| API JDBC (java.sql.*) | SunPKCS11 (sun.security.pkcs11.*) |
| Driver MySQL, Oracle, etc | Driver do fabricante do token (.so/.dll) |
| DataSource | Provider PKCS#11 configurado |

O SunPKCS11 acompanha o JDK e serve como ponte padronizada entre Java e qualquer biblioteca nativa de dispositivo criptografico.

### Como funciona

```
Java App
   │
   └── SunPKCS11 (JDK built-in)
            │
            └── Biblioteca nativa do fabricante (.so/.dll)
                     │
                     └── Dispositivo fisico (token/smartcard)
```

### Configuracao do SunPKCS11

O SunPKCS11 e configurado via arquivo de propriedades:

```
# pkcs11.cfg
name = MeuToken
library = /usr/lib/libpkcs11.so
```

E carregado em Java:

```java
Provider provider = Security.getProvider("SunPKCS11");
provider = provider.configure("/caminho/pkcs11.cfg");
Security.addProvider(provider);
```

### Implicacoes para o design

Quando se usa PKCS#11:
- A **chave privada nao e passada como parametro** — e obtida do dispositivo via `KeyStore`
- O PIN do token e necessario para autenticacao
- A assinatura ocorre dentro do dispositivo

Isso significa que a implementacao `PKCS11SignatureService` tem uma assinatura de metodo diferente na pratica. A interface `SignatureService` com `privateKey` como String e adequada para o `FakeSignatureService`, mas para PKCS#11 e necessario um mecanismo alternativo (ex: callback de PIN, configuracao externa).

**Para o escopo deste projeto:** implementar `FakeSignatureService` completamente e projetar `PKCS11SignatureService` com esqueleto funcional, documentando as limitacoes de ambiente (dispositivo nao disponivel).

---

## Dois Modos de Operacao

### Modo CLI (invocacao direta)

```bash
java -jar assinador.jar sign \
  --message <base64> \
  --private-key <identificador-ou-caminho>

java -jar assinador.jar validate \
  --message <base64> \
  --signature <base64> \
  --public-key <identificador-ou-caminho>
```

- Entrada: argumentos de linha de comando
- Saida: JSON em stdout
- Erros: JSON em stderr com codigo de saida != 0

### Modo Servidor HTTP

```
POST /sign          # Criar assinatura
POST /validate      # Validar assinatura
GET  /health        # Health check
POST /shutdown      # Desligar servidor
```

Controlado por `SignatureController`.

- Porta: configuravel via `--port` (padrao a definir)
- Content-Type: application/json

#### SignatureController

```java
@Controller
public class SignatureController {
    // POST /sign
    // POST /validate
}
```

#### Contrato /sign

Request:
```json
{
  "message": "<conteudo em base64>",
  "privateKey": "<identificador ou conteudo>"
}
```

Response 200:
```json
{
  "signature": "<valor base64 simulado>"
}
```

#### Contrato /validate

Request:
```json
{
  "message": "<conteudo em base64>",
  "signature": "<base64>",
  "publicKey": "<identificador ou conteudo>"
}
```

Response 200:
```json
{
  "valid": true
}
```

Response de erro (400):
```json
{
  "error": "PARAM_AUSENTE",
  "message": "O campo 'message' e obrigatorio"
}
```

---

## Parametros de Entrada e Saida

### Investigacao necessaria (a fazer)

Os parametros concretos devem ser investigados nas especificacoes FHIR antes da implementacao:
- https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-criar-assinatura.html
- https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-validar-assinatura.html

### Questoes a definir antes de implementar

1. **O que e `message`?** Documento completo em base64? Hash do documento? XML FHIR?
2. **O que e `privateKey`?** Caminho de arquivo `.p12`? Alias de keystore? Identificador PKCS#11?
3. **O que e `publicKey`?** Certificado X.509 em PEM? DER em base64?
4. **Qual o algoritmo padrao?** SHA256withRSA? Definido pelo certificado?
5. **Como fornecer os parametros?** Flags individuais? Arquivo JSON? Combinacao?

### Formas de fornecimento (decisao pendente)

| Opcao | Exemplo | Vantagem |
|-------|---------|----------|
| Flags individuais | `--message doc.b64 --key cert.pem` | Simples para scripts |
| Arquivo JSON | `--input params.json` | Facil para payloads grandes |
| Combinacao | flags para metadados + arquivo para conteudo | Flexivel |

**Recomendacao:** Combinacao — conteudo do documento via arquivo (`--message-file`), metadados via flags.

- [ ] **TODO:** Investigar especificacoes FHIR e definir parametros concretos antes da Sprint 2

---

## Validacao de Parametros

### Regras de validacao

| Campo | Regra |
|-------|-------|
| `message` | Obrigatorio; base64 valido; tamanho maximo a definir |
| `privateKey` / `publicKey` | Obrigatorio; formato valido (PEM, DER, alias PKCS#11) |
| `signature` (validate) | Obrigatorio; base64 valido |
| Algoritmo | Se fornecido, deve estar na lista suportada |

### Algoritmos suportados (a confirmar com FHIR)

- SHA256withRSA
- SHA384withRSA
- SHA512withRSA
- SHA256withECDSA

---

## Simulacao (FakeSignatureService)

### sign — resposta pre-construida

Quando parametros sao validos, retornar valor base64 fixo simulado, sempre o mesmo.

### validate — resultado pre-determinado

Logica simples:
- Se a assinatura recebida e igual ao valor fixo simulado pelo `sign` → `valid: true`
- Caso contrario → `valid: false`

---

## Tratamento de Erros

| Codigo | HTTP Status | Descricao |
|--------|-------------|-----------|
| `PARAM_AUSENTE` | 400 | Parametro obrigatorio nao fornecido |
| `PARAM_INVALIDO` | 400 | Parametro com formato invalido |
| `ALGORITMO_NAO_SUPORTADO` | 400 | Algoritmo fora da lista |
| `PAYLOAD_MUITO_GRANDE` | 413 | Conteudo excede tamanho maximo |
| `ERRO_INTERNO` | 500 | Erro inesperado |

---

## Tarefas de Implementacao

- [ ] Investigar parametros FHIR e definir contrato concreto
- [ ] Criar projeto Java (Maven/Gradle) com estrutura de pacotes
- [ ] Definir interface `SignatureService`
- [ ] Implementar `FakeSignatureService`
- [ ] Projetar esqueleto de `PKCS11SignatureService` (sem dispositivo fisico)
- [ ] Implementar `SignatureController` com `/sign`, `/validate`, `/health`, `/shutdown`
- [ ] Implementar modo CLI (parsing de args, saida JSON)
- [ ] Implementar validacao de parametros
- [ ] Implementar tratamento de erros estruturado
- [ ] Testes unitarios do `FakeSignatureService`
- [ ] Testes unitarios da validacao (todos os cenarios)
- [ ] Testes do `SignatureController` (modo servidor HTTP)
