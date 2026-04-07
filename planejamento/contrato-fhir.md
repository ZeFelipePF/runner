# Contrato FHIR — Parametros de /sign e /validate

Investigacao dos parametros FHIR para as operacoes de assinatura digital, conforme:
- https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-criar-assinatura.html
- https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-validar-assinatura.html

**Data da investigacao:** 2026-03-31

---

## Resumo Executivo

A especificacao FHIR e **muito mais complexa** do que o contrato provisorio estimado. O escopo real inclui:
- Bundles e Provenance FHIR como entrada (nao apenas mensagens base64 simples)
- Material criptografico com 5 modalidades (PEM, PKCS#12, SMARTCARD, TOKEN, REMOTE)
- Cadeia de certificados ICP-Brasil completa
- Timestamp com estrategia `iat` ou `tsa`
- Politica de assinatura versionada
- Configuracoes operacionais extensas (OCSP, CRL, TSA, trust store)
- Canonicalizacao JSON (RFC 8785)
- Saida em formato JWS JSON Serialization (RFC 7515)

**Para o escopo deste projeto (FakeSignatureService)**, vamos simplificar significativamente, mantendo a interface compativel com os campos reais.

---

## Operacao: Criar Assinatura (`POST /sign`)

### Parametros de Entrada (Simplificados para o Assinador)

| Campo | Tipo | Obrigatorio | Descricao |
|-------|------|-------------|-----------|
| `bundle` | string (JSON) | Sim | Bundle FHIR serializado em JSON |
| `provenance` | string (JSON) | Sim | Provenance FHIR serializado em JSON |
| `materialCriptografico` | object | Sim | Material criptografico (ver subtipos) |
| `certificados` | array[string] | Sim | Cadeia X.509v3 em base64 (folha ate raiz) |
| `timestampReferencia` | integer | Sim | Unix timestamp UTC |
| `estrategiaTimestamp` | string | Sim | `iat` ou `tsa` |
| `politicaAssinatura` | string (URI) | Sim | URI versionada da politica |
| `configuracoesOperacionais` | object | Nao | Parametros de verificacao |

### Subtipos de Material Criptografico

**PEM:**
```json
{ "tipo": "PEM", "chavePrivada": "<PKCS#8 PEM>", "senha": "<opcional>" }
```

**PKCS#12:**
```json
{ "tipo": "PKCS12", "conteudo": "<base64>", "senha": "<obrigatorio>", "alias": "<obrigatorio>" }
```

**SMARTCARD / TOKEN (PKCS#11):**
```json
{ "tipo": "SMARTCARD", "pin": "<obrigatorio>", "identificador": "<obrigatorio>", "slotId": 0, "tokenLabel": "<opcional>" }
```

**REMOTE:**
```json
{ "tipo": "REMOTE", "enderecoServico": "<URL>", "credencial": {} }
```

### Parametros de Saida

**Sucesso (200):**
```json
{
  "resourceType": "Signature",
  "type": [{ "system": "urn:iso-astm:E1762-95:2013", "code": "1.2.840.10065.1.12.1.1" }],
  "when": "2026-03-31T10:00:00Z",
  "who": {
    "identifier": { "system": "urn:brasil:cpf", "value": "12345678901" }
  },
  "sigFormat": "application/jose",
  "targetFormat": "application/octet-stream",
  "data": "<base64 JWS>"
}
```

**Erro (400):**
```json
{
  "resourceType": "OperationOutcome",
  "issue": [{
    "severity": "error",
    "code": "invalid",
    "details": {
      "coding": [{ "system": "https://fhir.saude.go.gov.br/CodeSystem/situacao-excepcional-assinatura", "code": "CERT.EXPIRED" }],
      "text": "Certificado expirado"
    },
    "diagnostics": "Certificado expirou em 2025-12-15T10:30:00Z"
  }]
}
```

### Algoritmos Suportados

| Algoritmo | Tipo Chave | Tamanho Minimo |
|-----------|-----------|----------------|
| RS256 | RSA | 2048 bits (recomendado 3072+) |
| ES256 | ECC | P-256 (secp256r1) |

---

## Operacao: Validar Assinatura (`POST /validate`)

### Parametros de Entrada

| Campo | Tipo | Obrigatorio | Descricao |
|-------|------|-------------|-----------|
| `jws` | string (base64) | Sim | JWS JSON Serialization completa em base64 |
| `trustStore` | array[string] | Sim | Hashes SHA-256 (hex, 64 chars) das AC-Raiz aceitas |
| `minCertIssueDate` | integer | Sim | Data minima de certificados (Unix UTC) |
| `referenceTimestamp` | integer | Sim | Timestamp para checagens temporais (Unix UTC) |
| `signaturePolicyId` | string (URI) | Sim | URI versionada da politica |
| `ocspTimeout` | integer | Nao | Timeout OCSP em segundos (padrao: 30) |
| `crlTimeout` | integer | Nao | Timeout CRL em segundos (padrao: 30) |
| `tsaTimeout` | integer | Nao | Timeout TSA em segundos (padrao: 30) |
| `revocationCacheTtl` | integer | Nao | TTL cache de revogacao (padrao: 3600s) |
| `nearExpiryThresholdDays` | integer | Nao | Limiar proximidade expiracao (padrao: 30) |
| `signatureAgeThresholdDays` | integer | Nao | Idade maxima da assinatura (padrao: 365) |
| `revocationPolicy` | string enum | Nao | `strict`, `soft-fail`, `warn` (padrao: strict) |
| `bundleOriginal` | object | Nao | Bundle original (verificacao integridade opcional) |
| `provenanceOriginal` | object | Nao | Provenance original (verificacao integridade opcional) |

### Parametros de Saida

**Sucesso (200):**
```json
{
  "resourceType": "OperationOutcome",
  "issue": [{
    "severity": "information",
    "code": "informational",
    "details": {
      "coding": [{ "code": "VALIDATION.SUCCESS" }],
      "text": "Assinatura digital validada com sucesso"
    },
    "diagnostics": "Algoritmo: RS256 | Politica: v1.0 | Estrategia: tsa"
  }]
}
```

**Erro (400):**
```json
{
  "resourceType": "OperationOutcome",
  "issue": [{
    "severity": "error",
    "code": "invalid",
    "details": {
      "coding": [{ "code": "CERT.EXPIRED" }],
      "text": "Certificado expirado na cadeia"
    },
    "location": ["signatures[0].protected.x5c[0]"],
    "diagnostics": "Certificado expirou em 2024-12-15T10:30:00Z"
  }]
}
```

---

## Codigos de Erro Padronizados

### Criar Assinatura
| Codigo | Descricao |
|--------|-----------|
| POLICY.MISSING | Politica nao fornecida |
| POLICY.URI-INVALID | Formato URI invalido |
| POLICY.VERSION-UNSUPPORTED | Versao nao suportada |
| FORMAT.BUNDLE-MALFORMED | Bundle nao conforme FHIR |
| FORMAT.BUNDLE-EMPTY | Bundle sem entradas |
| CERT.CHAIN-INCOMPLETE | Cadeia < 2 certificados |
| CERT.EXPIRED | Certificado expirado |
| CERT.NOT-ICP-BRASIL | Certificado nao ICP-Brasil |
| CERT.REVOKED | Certificado revogado |
| CERT.WEAK-KEY | Tamanho chave insuficiente |
| TIMESTAMP.OUT-OF-TOLERANCE-WINDOW | Desvio > +/-5 minutos |
| SECURITY.BUNDLE-SIZE-LIMIT-EXCEEDED | Bundle excede limite |
| CRYPTO.PIN-INVALID | PIN invalido |

### Validar Assinatura
| Codigo | Descricao |
|--------|-----------|
| VALIDATION.SUCCESS | Validacao bem-sucedida |
| VALIDATION.SIGNATURE-VERIFICATION-FAILED | Verificacao criptografica falhou |
| VALIDATION.UNSUPPORTED-ALGORITHM | Algoritmo nao suportado |
| VALIDATION.TIMESTAMP-STRATEGY-INVALID | Estrategia timestamp invalida |
| FORMAT.BASE64-INVALID | Dados base64 invalidos |
| FORMAT.JWS-MALFORMED | JWS nao conforme RFC 7515 |
| CERT.INVALID-FORMAT | Certificados x5c malformados |
| CERT.CHAIN-INCOMPLETE | Menos de 2 certificados |
| CERT.EXPIRED | Certificado expirado |
| CERT.REVOKED | Certificado revogado |
| POLICY.VERSION-UNSUPPORTED | Versao politica nao suportada |

---

## Contrato Simplificado para FakeSignatureService (Sprint 2)

Para a implementacao simulada, o `FakeSignatureService` aceitara os parametros completos mas:
1. **Validara** campos obrigatorios (bundle, certificados, etc.)
2. **Ignorara** verificacoes criptograficas reais (ICP-Brasil, OCSP, CRL)
3. **Retornara** respostas pre-construidas no formato FHIR correto
4. A validacao de parametros reais (tipos, ranges, formatos) sera implementada no pacote `validacao/`

### Interface Simplificada do Assinador

**POST /sign** — Aceita request com campos obrigatorios, retorna Signature FHIR simulada
**POST /validate** — Aceita JWS + trustStore + campos, retorna OperationOutcome

O contrato completo acima serve como referencia para a implementacao da validacao de parametros na Sprint 2.
