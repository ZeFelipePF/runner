# Entregavel: Estrategia de Testes

## Objetivo

Garantir qualidade e corretude do sistema atraves de testes unitarios, de integracao e de aceitacao, com cobertura adequada.

## Niveis de Teste

### 1. Testes Unitarios

Testam funcoes e metodos isoladamente.

**CLI (assinatura/simulador):**
- Parsing de argumentos e flags
- Validacao de entrada do usuario
- Formatacao de saida
- Logica de deteccao de JDK
- Parsing de versao

**assinador.jar:**
- Validacao de cada parametro FHIR
- Geracao de resposta simulada
- Tratamento de erros
- Serialization/deserialization JSON

### 2. Testes de Integracao

Testam a comunicacao entre componentes.

**CLI <-> assinador.jar (modo CLI):**
- CLI invoca assinador.jar e recebe resposta valida
- CLI trata erro do assinador.jar corretamente
- Parametros sao passados corretamente

**CLI <-> assinador.jar (modo HTTP):**
- CLI envia requisicao e recebe resposta
- Timeout e erros de rede tratados
- Servidor inicia e para corretamente

**Provisionamento JDK:**
- Deteccao de JDK existente
- Download e extracao (pode usar mock para CI)

### 3. Testes de Aceitacao

Baseados nos criterios de aceitacao de cada US.

**US-01:**
- [ ] CLI aceita comandos criar e validar
- [ ] CLI invoca assinador.jar com parametros fornecidos
- [ ] CLI funciona em modo local e HTTP
- [ ] CLI exibe resultado de forma legivel
- [ ] CLI inicia servidor na porta padrao
- [ ] CLI detecta instancia existente do servidor
- [ ] CLI interrompe servidor
- [ ] CLI suporta shutdown programado

**US-02:**
- [ ] assinador.jar valida todos os parametros
- [ ] assinador.jar simula criacao de assinatura
- [ ] assinador.jar simula validacao de assinatura
- [ ] assinador.jar suporta PKCS#11
- [ ] assinador.jar retorna mensagens de erro claras

**US-03:**
- [ ] CLI inicia simulador
- [ ] CLI verifica portas disponiveis
- [ ] CLI para simulador
- [ ] CLI exibe status
- [ ] simulador.jar baixado via GitHub Releases
- [ ] Nao baixa se versao ja esta local

**US-04:**
- [ ] Detecta JDK presente
- [ ] Baixa JDK quando ausente
- [ ] JDK disponivel para uso
- [ ] Funciona nas 3 plataformas

**US-05:**
- [ ] Binario para Windows
- [ ] Binario para Linux
- [ ] Binario para macOS
- [ ] Via GitHub Releases
- [ ] Checksums SHA256
- [ ] SemVer

## Frameworks de Teste

### Para o CLI (depende da linguagem)

| Linguagem | Framework | Observacao |
|-----------|-----------|------------|
| Go | `testing` (stdlib) + `testify` | Nativo, sem dependencia extra obrigatoria |
| Rust | `cargo test` (builtin) | Nativo |
| Python | `pytest` | Precisa instalar |

### Para o assinador.jar

| Framework | Uso |
|-----------|-----|
| JUnit 5 | Testes unitarios e de integracao |
| Mockito | Mocks quando necessario |
| AssertJ | Assertions fluentes (opcional) |
| REST Assured | Testes do servidor HTTP (opcional) |

## Cobertura

- Meta minima: **80%** de cobertura de linhas
- Foco em logica de validacao (maior risco)
- Nao exigir cobertura em codigo boilerplate (main, config)

## CI Matrix

Testes devem executar nas 3 plataformas via GitHub Actions:

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest, macos-latest]
```

## Tarefas de Implementacao

- [ ] Configurar framework de teste do CLI
- [ ] Configurar JUnit 5 no assinador.jar
- [ ] Escrever testes unitarios do CLI
- [ ] Escrever testes unitarios do assinador.jar
- [ ] Escrever testes de integracao CLI <-> assinador.jar
- [ ] Escrever testes de aceitacao por US
- [ ] Configurar CI matrix para 3 plataformas
- [ ] Configurar relatorio de cobertura
