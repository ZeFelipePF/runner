# Exemplos de Uso — Sistema Runner

Exemplos prontos para copiar e colar. Os payloads são **completos e válidos** contra o
`assinador.jar` em modo simulado.

> Referência de comandos: [Manual do Usuário](manual-usuario.md) ·
> Contratos: [Documentação Técnica](tecnico.md).

---

## 1. Criar assinatura (modo local)

Invocação direta, sem deixar servidor de pé. Bom para scripts pontuais.

`pedido-assinatura.json`:

```json
{
  "bundle": "eyJyZXNvdXJjZVR5cGUiOiJCdW5kbGUifQ==",
  "provenance": "eyJyZXNvdXJjZVR5cGUiOiJQcm92ZW5hbmNlIn0=",
  "materialCriptografico": {
    "tipo": "PEM",
    "chavePrivada": "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----"
  },
  "certificados": ["Q2VydGlmaWNhZG9YNTA5LWJhc2U2NA=="],
  "timestampReferencia": 1764547200,
  "estrategiaTimestamp": "tsa",
  "politicaAssinatura": "https://politicas.icpbrasil.gov.br/PA_AD_RB_v2_3.der",
  "algoritmo": "RS256"
}
```

```bash
assinatura criar --payload pedido-assinatura.json --modo local
```

Saída (stdout):

```json
{
  "resourceType": "Signature",
  "type": [ { "system": "urn:iso-astm:E1762-95:2013", "code": "1.2.840.10065.1.12.1.1" } ],
  "when": "2025-12-01T00:00:00Z",
  "sigFormat": "application/jose",
  "targetFormat": "application/octet-stream",
  "data": "QVNTSU5BVFVSQS1TSU1VTEFEQS12MQ=="
}
```

> Guarde o valor de `data` — ele é usado para validar a assinatura (exemplo 3).

---

## 2. Criar assinatura (modo HTTP)

O modo padrão. Sobe o `assinador.jar` como servidor na primeira chamada e reaproveita
nas seguintes (mais rápido para várias operações).

```bash
# (opcional) deixar o servidor de pé explicitamente
assinatura servidor iniciar
# → { "status": "STARTED", "porta": 8088, "pid": 40231, "reusado": false }

# criar (modo http é o padrão; --modo http é opcional)
assinatura criar --payload pedido-assinatura.json

# a próxima chamada reaproveita a mesma instância (warm start)
assinatura criar --payload outro-pedido.json
```

Lendo o payload do stdin (útil em pipelines):

```bash
cat pedido-assinatura.json | assinatura criar --payload -
```

---

## 3. Validar assinatura

`pedido-validacao.json` — o campo `jws` recebe o `data` retornado por `criar`:

```json
{
  "jws": "QVNTSU5BVFVSQS1TSU1VTEFEQS12MQ==",
  "trustStore": ["e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"],
  "minCertIssueDate": 1700000000,
  "referenceTimestamp": 1764547200,
  "signaturePolicyId": "https://politicas.icpbrasil.gov.br/PA_AD_RB_v2_3.der"
}
```

```bash
assinatura validar --payload pedido-validacao.json
```

Saída (assinatura corresponde ao valor simulado → sucesso):

```json
{
  "resourceType": "OperationOutcome",
  "issue": [ {
    "severity": "information",
    "code": "informational",
    "details": {
      "coding": [ { "code": "VALIDATION.SUCCESS" } ],
      "text": "Assinatura digital validada com sucesso"
    },
    "diagnostics": "Politica: https://politicas.icpbrasil.gov.br/PA_AD_RB_v2_3.der"
  } ]
}
```

Se o `jws` **não** for o valor simulado, a saída traz `severity: error` e
`code: VALIDATION.SIGNATURE-VERIFICATION-FAILED`.

---

## 4. Tratamento de erros (para scripts)

O exit code reflete o tipo de erro. Exemplo: payload sem o campo `bundle`.

```bash
echo '{"provenance":"x"}' | assinatura criar --payload -
echo "exit code: $?"
```

Saída no **stderr**:

```json
{ "error": "PARAM_AUSENTE", "message": "campo 'bundle' obrigatorio" }
```

```
exit code: 2
```

Uso em script:

```bash
if assinatura criar --payload pedido.json > assinatura.json 2> erro.json; then
  echo "OK: $(cat assinatura.json)"
else
  echo "Falhou (exit $?): $(cat erro.json)"
fi
```

Tabela de exit codes no [Manual do Usuário](manual-usuario.md#6-códigos-de-saída-exit-codes).

---

## 5. Iniciar e gerenciar o simulador

```bash
# iniciar (baixa o latest do GitHub Releases na primeira vez; provisiona o JRE se preciso)
simulador iniciar
# → { "status": "STARTED", "pid": 41234, "porta": 8443, "versao": "v1.2.0", ... }

# status
simulador status
# → { "registrado": true, "running": true, "pid": 41234, "porta": 8443, "uptimeSegundos": 42, ... }

# parar
simulador parar
# → { "status": "STOPPED", "pid": 41234, "porta": 8443, "metodo": "http_shutdown" }
```

Variações úteis:

```bash
# versão específica
simulador iniciar --versao-simulador v1.2.0

# a partir de um jar local (ambiente offline)
simulador iniciar --source file:///home/voce/simulador.jar

# a partir de um servidor interno
simulador iniciar --source https://intranet/artefatos/simulador.jar

# forçar uma porta
simulador iniciar --porta 9100
```

---

## 6. Provisionar o JDK manualmente (opcional)

O JRE 21 é baixado automaticamente quando ausente. Para usar um Java já instalado e
**evitar o download**, basta expô-lo antes de rodar os CLIs:

```bash
# usa o JDK apontado por JAVA_HOME
export JAVA_HOME=/usr/lib/jvm/temurin-21
assinatura servidor iniciar    # detecta JAVA_HOME e dispensa o download
```

Ordem de detecção: `JAVA_HOME` → `java` no `PATH` → `~/.hubsaude/jdk/*/bin/java`.

Para forçar o download de uma fonte alternativa (ex.: mirror interno da Adoptium):

```bash
export HUBSAUDE_ADOPTIUM_URL=https://mirror.interno/adoptium/v3
simulador iniciar
```

---

## 7. Verificar um artefato com Cosign

Após baixar um artefato e seus arquivos `.sig`/`.pem` da Releases:

```bash
cosign verify-blob \
  --certificate assinatura-1.0.0-linux-amd64.AppImage.pem \
  --signature   assinatura-1.0.0-linux-amd64.AppImage.sig \
  --certificate-identity-regexp ".*" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  assinatura-1.0.0-linux-amd64.AppImage
# → Verified OK
```

Detalhes no [Guia de Instalação](instalacao.md#4-verificação-de-autenticidade-cosign).

---

## 8. Fluxo completo de ponta a ponta

```bash
# 1. criar assinatura e salvar o resultado
assinatura criar --payload pedido-assinatura.json > assinatura.json

# 2. extrair o campo data e montar o payload de validação
DATA=$(grep -o '"data": *"[^"]*"' assinatura.json | sed 's/.*"data": *"\(.*\)"/\1/')
cat > validacao.json <<EOF
{
  "jws": "$DATA",
  "trustStore": ["e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"],
  "minCertIssueDate": 1700000000,
  "referenceTimestamp": 1764547200,
  "signaturePolicyId": "https://politicas.icpbrasil.gov.br/PA_AD_RB_v2_3.der"
}
EOF

# 3. validar
assinatura validar --payload validacao.json

# 4. encerrar o servidor ao final
assinatura servidor parar
```
