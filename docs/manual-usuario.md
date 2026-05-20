# Manual do Usuário — Sistema Runner

Este manual destina-se aos usuários dos CLIs `assinatura` e `simulador`. Não é
necessário conhecer Java nem instalar o JDK manualmente — o provisionamento é
realizado automaticamente pelo sistema.

> Para instruções de instalação, consulte o [Guia de Instalação](instalacao.md).
> Para detalhes de funcionamento interno, consulte a [Documentação Técnica](tecnico.md).
> Para exemplos prontos, consulte os [Exemplos de Uso](exemplos.md).

---

## 1. Conceitos fundamentais

O Sistema Runner é composto por dois programas de linha de comando:

| CLI | Função |
|-----|--------|
| `assinatura` | Cria e valida assinaturas digitais (simuladas), invocando o `assinador.jar`. |
| `simulador` | Gerencia o ciclo de vida do `simulador.jar` do HubSaúde (iniciar, parar e monitorar). |

Ambos persistem o estado no diretório `~/.hubsaude/`, no diretório pessoal do usuário
(no Windows: `C:\Users\<usuário>\.hubsaude\`). Em uso normal, não é necessária
intervenção manual nessa pasta.

O `assinatura` comunica-se com o `assinador.jar` em **dois modos**:

- **modo `http`** (padrão): inicia o `assinador.jar` como um servidor persistente.
  A primeira chamada tem maior latência, devido à inicialização da JVM; as chamadas
  subsequentes são atendidas com menor latência. Recomendado para múltiplas operações.
- **modo `local`**: executa `java -jar assinador.jar ...` a cada chamada, encerrando o
  processo ao final. Cada execução incorre no custo de inicialização da JVM
  (_cold start_). Recomendado para execuções pontuais ou scripts de automação.

---

## 2. Primeiros passos

```bash
# 1. Confirme que o binário está acessível
assinatura --versao
simulador --versao

# 2. Consulte a ajuda de qualquer comando
assinatura --help
assinatura criar --help
simulador iniciar --help
```

Na primeira execução efetiva, o sistema pode efetuar os seguintes downloads automáticos:
- o **JDK 21**, caso não haja Java instalado na máquina — destinado a `~/.hubsaude/jdk/`;
- o **`simulador.jar`** (no comando `simulador iniciar`) — destinado a `~/.hubsaude/simulador/`.

Esses downloads ocorrem uma única vez e são mantidos em cache.

---

## 3. Flags globais

Aplicáveis a qualquer subcomando dos dois CLIs:

| Flag | Efeito |
|------|--------|
| `--verbose` | Saída detalhada (nível debug) no stderr. |
| `--quiet` | Suprime as mensagens de progresso (stderr); mantém o resultado JSON no stdout. |
| `-h`, `--help` | Exibe a ajuda do comando. |

**Convenção de saída** (relevante para scripts):
- O **`stdout`** recebe sempre o **resultado final em JSON** (formatado).
- O **`stderr`** recebe **progresso e erros** operacionais.
- O **código de saída** (exit code) indica sucesso (`0`) ou o tipo de erro (ver seção 6).

---

## 4. CLI `assinatura`

### 4.1. `assinatura criar`

Cria uma assinatura digital simulada a partir de um payload FHIR em JSON.

```
Uso:
  assinatura criar [flags]

Flags:
  --payload <arquivo>   Arquivo JSON com o payload FHIR ('-' = lê do stdin)
  --modo <local|http>   Modo de invocação (padrão: http)
  --porta <n>           Porta do servidor HTTP (0 = usa config / auto-detecta)
  --jar <caminho>       Caminho do assinador.jar (padrão: localizar automaticamente)
```

Exemplos:

```bash
# Modo HTTP (padrão) — inicia o servidor na primeira chamada e o reaproveita
assinatura criar --payload pedido-assinatura.json

# Leitura do payload via stdin
cat pedido-assinatura.json | assinatura criar --payload -

# Modo local — invocação direta, sem manter o servidor em execução
assinatura criar --payload pedido-assinatura.json --modo local
```

O conteúdo do `--payload` segue o contrato FHIR de criação. Campos obrigatórios:
`bundle`, `provenance`, `materialCriptografico`, `certificados`, `timestampReferencia`,
`estrategiaTimestamp` (`iat` ou `tsa`) e `politicaAssinatura`. O campo `algoritmo`
é opcional, mas, se presente, deve ser `RS256` ou `ES256`.
Um payload completo está disponível em [Exemplos de Uso](exemplos.md).

Saída de sucesso (recurso `Signature` do FHIR R4):

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

> O valor de `data` é uma assinatura **simulada** fixa — o projeto não executa
> criptografia real (ver escopo na especificação).

### 4.2. `assinatura validar`

Valida uma assinatura. As flags são idênticas às de `criar`.

```bash
assinatura validar --payload pedido-validacao.json
```

Campos obrigatórios do payload de validação: `jws` (string base64), `trustStore`
(lista de hashes SHA-256 em hexadecimal, 64 caracteres cada), `minCertIssueDate` (inteiro),
`referenceTimestamp` (inteiro) e `signaturePolicyId`.

Saída de sucesso (recurso `OperationOutcome` do FHIR R4):

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
    "diagnostics": "Politica: <sua-politica>"
  } ]
}
```

> Como a assinatura é simulada, a validação retorna `severity: information`
> quando o campo `jws` corresponde ao valor fixo retornado por `criar`
> (`QVNTSU5BVFVSQS1TSU1VTEFEQS12MQ==`); caso contrário, retorna `severity: error`.

### 4.3. `assinatura servidor`

Gerencia manualmente o `assinador.jar` no modo servidor. Aplicável quando se deseja
manter o servidor em execução antes de uma sequência de operações ou encerrá-lo ao final.

```bash
# Iniciar (auto-detecta a porta a partir da 8088 se --porta não for informada)
assinatura servidor iniciar
assinatura servidor iniciar --porta 8088

# Consultar status (PID, porta e resposta do endpoint /health)
assinatura servidor status

# Parar (tenta shutdown HTTP; em caso de falha, encerra o processo)
assinatura servidor parar
```

`servidor iniciar` reaproveita a instância existente: havendo uma instância saudável
registrada em `state.json`, ela é **reutilizada** (a saída traz `"reusado": true` e
`"status": "REUSED"`).

Flags de `servidor iniciar`:

| Flag | Efeito |
|------|--------|
| `--porta <n>` | Porta desejada (0 = config / auto-detecta a partir da 8088). |
| `--jar <caminho>` | Caminho explícito do `assinador.jar`. |

---

## 5. CLI `simulador`

### 5.1. `simulador iniciar`

Efetua o download (quando necessário) e inicia o `simulador.jar` do HubSaúde,
provisionando o JDK quando ausente.

```
Uso:
  simulador iniciar [flags]

Flags:
  --porta <n>                 Porta do simulador (0 = auto-detecta a partir da 9090)
  --source <url>              URL alternativa para o simulador.jar (http://, https://, file://)
  --versao-simulador <tag>    Versão a baixar via GitHub Releases (padrão: latest)
```

Exemplos:

```bash
# Início padrão (baixa a versão latest do GitHub Releases na primeira execução)
simulador iniciar

# Versão específica
simulador iniciar --versao-simulador v1.2.0

# A partir de um arquivo local (não utiliza o GitHub Releases)
simulador iniciar --source file:///home/usuario/simulador.jar

# A partir de um servidor interno
simulador iniciar --source https://intranet/artefatos/simulador.jar

# Definir uma porta específica
simulador iniciar --porta 9100
```

À semelhança de `assinatura servidor`, este comando **reutiliza** uma instância já em
execução quando há uma instância saudável registrada.

### 5.2. `simulador status`

```bash
simulador status
```

Retorna se há simulador registrado, se o PID está ativo, a porta, a versão, o `uptime` e
a resposta do health check:

```json
{
  "registrado": true,
  "running": true,
  "pid": 41234,
  "porta": 9090,
  "iniciadoEm": "2025-12-01T09:00:00Z",
  "pidVivo": true,
  "versao": "v1.2.0",
  "uptimeSegundos": 137
}
```

### 5.3. `simulador parar`

Encerra o simulador de forma controlada, com escalonamento de estratégia quando
necessário: `POST /shutdown` (HTTP) → `SIGTERM` → `SIGKILL`.

```bash
simulador parar
```

```json
{ "status": "STOPPED", "pid": 41234, "porta": 9090, "metodo": "http_shutdown" }
```

O campo `metodo` indica a forma de encerramento: `http_shutdown`, `sigterm`, `sigkill`,
`stale_pid` (o PID já não existia) ou `nao_estava_rodando`.

---

## 6. Códigos de saída (exit codes)

O `assinatura` propaga, no código de saída, o tipo de erro retornado pelo `assinador.jar`,
permitindo o tratamento condicional em scripts.

| Exit | Código | Significado |
|------|--------|-------------|
| 0 | — | Sucesso |
| 1 | `CLI_ERROR` / genérico | Erro do próprio CLI (jar não localizado, falha de conexão, etc.) |
| 2 | `PARAM_AUSENTE` | Campo FHIR obrigatório ausente |
| 3 | `PARAM_INVALIDO` | Campo FHIR mal formado (ex.: base64 inválido) |
| 4 | `ALGORITMO_NAO_SUPORTADO` | Algoritmo distinto de `RS256`/`ES256` |
| 5 | `ASSINATURA_INVALIDA` | Reservado — validação de assinatura falhou |
| 6 | `DISPOSITIVO_INDISPONIVEL` / `PIN_INVALIDO` | Reservado — PKCS#11 |
| 7 | `ERRO_INTERNO` | Falha interna do `assinador.jar` |

Em caso de erro, o `assinatura` emite no **stderr** um JSON estruturado:

```json
{ "error": "PARAM_AUSENTE", "message": "campo 'bundle' obrigatorio" }
```

> Os códigos 5 e 6 estão **reservados** no CLI para a evolução do PKCS#11 e da
> validação criptográfica real; o `assinador.jar` atual (modo simulado) não os emite.

---

## 7. Solução de problemas

| Sintoma | Causa provável | Ação recomendada |
|---------|----------------|-------------------|
| `assinador.jar não localizado` | O CLI não localizou o jar | Defina `HUBSAUDE_ASSINADOR_JAR=/caminho/assinador.jar` ou utilize `--jar`. Consulte a ordem de busca na [documentação técnica](tecnico.md). |
| Primeira execução com latência elevada | Download do JDK/`simulador.jar` | Comportamento esperado; ocorre apenas uma vez. Utilize `--verbose` para acompanhar. |
| `porta em uso` / o comando selecionou outra porta | A porta padrão estava ocupada | O CLI seleciona automaticamente a próxima porta livre (janela de +20) e a registra em `state.json`. Verifique com `... status`. |
| `status` indica `running: false` apesar de haver processo | PID obsoleto em `state.json` | Execute `... parar` para limpar o estado e, em seguida, `iniciar` novamente. |
| Ausência de conexão ao baixar JDK/simulador | Ambiente offline | Para o simulador, utilize `--source file://...`. Para o JDK, instale o Java 21 e exponha-o via `JAVA_HOME` ou `PATH`. |

### Variáveis de ambiente

| Variável | Uso |
|----------|-----|
| `HUBSAUDE_ASSINADOR_JAR` | Caminho explícito para o `assinador.jar`. |
| `HUBSAUDE_ADOPTIUM_URL` | Sobrescreve o endpoint da API Adoptium (download do JDK). |
| `HUBSAUDE_SIMULADOR_REPO` | Sobrescreve o repositório GitHub Releases do `simulador.jar`. |

---

## 8. Localização dos dados

```
~/.hubsaude/
├── jdk/                 JDK 21 provisionado automaticamente (Temurin)
├── simulador/           simulador.jar + simulador-meta.json (versão/hash)
├── state.json           PID e porta dos processos em execução
└── config.json          Configurações do usuário (portas padrão, etc.)
```

Para reinicializar o ambiente, encerre os processos (`assinatura servidor parar`,
`simulador parar`) e remova o diretório `~/.hubsaude/`. Ele será recriado na execução
seguinte.
