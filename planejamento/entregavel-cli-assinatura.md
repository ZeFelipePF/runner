# Entregavel: CLI assinatura (US-01)

## Objetivo

CLI multiplataforma que permite ao usuario invocar operacoes de assinatura digital atraves do assinador.jar, sem precisar conhecer detalhes de Java.

## Historias Relacionadas

- **US-01**: Invocar assinador.jar via CLI
- **US-04**: Provisionar JDK automaticamente (dependencia)
- **US-05**: Disponibilizar binarios multiplataforma

## Comandos Planejados

```
assinatura criar [flags]       # Criar assinatura digital
assinatura validar [flags]     # Validar assinatura digital
assinatura servidor iniciar    # Iniciar assinador.jar no modo servidor
assinatura servidor parar      # Parar assinador.jar no modo servidor
assinatura servidor status     # Verificar se o servidor esta rodando
assinatura --help              # Ajuda geral
assinatura --versao            # Versao do CLI
```

## Flags Globais

| Flag | Descricao | Padrao |
|------|-----------|--------|
| `--modo` | Modo de invocacao: `local` ou `http` | `http` |
| `--porta` | Porta do modo servidor (0 = auto-detectar) | `0` (auto) |
| `--verbose` | Saida detalhada | `false` |
| `--quiet` | Suprimir mensagens de progresso do startup | `false` |

## Parametros de Entrada — Investigacao Necessaria

Os parametros concretos dos comandos `criar` e `validar` **devem ser investigados** nas especificacoes FHIR antes de serem implementados.

**Referencias obrigatorias:**
- https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-criar-assinatura.html
- https://fhir.saude.go.gov.br/r4/seguranca/caso-de-uso-validar-assinatura.html

**Questoes a responder:**
1. O que e o "documento" a assinar? Arquivo JSON? XML? Conteudo em base64?
2. Como o certificado e fornecido? Arquivo `.p12`? Arquivo `.pem`? Alias de keystore? Identificador PKCS#11?
3. Quais campos sao obrigatorios vs opcionais?
4. Como fornecer payloads grandes? Via arquivo (`--message-file`) ou stdin?

**- [ ] TODO (Sprint 1): Investigar e definir os parametros antes de implementar**

## Flags do Comando `criar` (rascunho — sujeito a revisao apos investigacao FHIR)

| Flag | Descricao | Obrigatoria |
|------|-----------|-------------|
| `--message-file` | Caminho do arquivo com o conteudo a assinar | Sim |
| `--private-key` | Caminho do arquivo de chave/certificado (`.p12`, `.pem`) | Sim |
| `--algoritmo` | Algoritmo de assinatura | Nao (padrao: SHA256withRSA) |

## Flags do Comando `validar` (rascunho — sujeito a revisao)

| Flag | Descricao | Obrigatoria |
|------|-----------|-------------|
| `--message-file` | Caminho do arquivo com o conteudo original | Sim |
| `--signature` | Assinatura a validar (base64 ou caminho de arquivo) | Sim |
| `--public-key` | Chave publica ou certificado para validacao | Sim |

## Comportamentos Chave

### Gerenciamento de Porta

O CLI **nao falha** se a porta padrao estiver ocupada. Comportamento:

1. Se `--porta` especificada: tentar usar exatamente essa porta
   - Se ocupada: erro informando qual processo ocupa
2. Se `--porta` nao especificada (padrao):
   - Tentar porta do `config.json` (ou 8088 como default)
   - Se ocupada: tentar proximas (8089, 8090, ...) ate encontrar livre
   - Informar ao usuario qual porta foi escolhida

```
[i] Porta 8088 em uso. Usando porta 8089.
[✓] Assinador iniciado em http://localhost:8089
```

### Modo HTTP (padrao)

Ver `startup.md` para o fluxo completo. Em resumo:

1. Verificar se assinador.jar ja esta rodando (state.json + PID)
2. Se nao estiver: detectar porta disponivel e iniciar automaticamente
3. Enviar requisicao `POST /sign` ou `POST /validate`
4. Exibir resultado formatado

### Modo Local (CLI)

1. (Startup: garantir JDK disponivel)
2. Executar `java -jar assinador.jar sign [flags]`
3. Capturar stdout/stderr
4. Parsear JSON e exibir resultado formatado

### Gerenciamento Explicito do Servidor

```
assinatura servidor iniciar [--porta X]   # Iniciar e registrar em state.json
assinatura servidor parar [--porta X]     # Parar via HTTP shutdown ou SIGTERM
assinatura servidor status                # Ler state.json e verificar processo
```

Suportar shutdown programado: `assinatura servidor iniciar --timeout 30` (desliga apos 30 minutos sem uso).

## Tratamento de Erros

| Cenario | Comportamento |
|---------|---------------|
| JDK nao encontrado e falha no download | Mensagem clara com instrucoes manuais |
| assinador.jar nao encontrado | Orientar usuario sobre como obter |
| Servidor nao responde (health check timeout) | Exibir saida do processo + sugestao |
| Parametros invalidos (validacao no CLI) | Mostrar erro + uso correto |
| Parametros invalidos (erro do assinador) | Propagar mensagem estruturada do assinador |
| Porta explicita ocupada | Informar processo que ocupa a porta |
| Todas as portas no range ocupadas | Erro com sugestao de porta manual |

## Tarefas de Implementacao

- [ ] Investigar parametros FHIR e definir flags concretas
- [ ] Configurar projeto na linguagem escolhida
- [ ] Implementar parsing de comandos e flags
- [ ] Implementar `--help` para todos os comandos
- [ ] Implementar auto-deteccao de porta disponivel
- [ ] Implementar invocacao direta (modo local)
- [ ] Implementar invocacao HTTP (modo http) com `POST /sign` e `POST /validate`
- [ ] Implementar deteccao de servidor existente via state.json
- [ ] Implementar iniciar/parar/status do servidor
- [ ] Implementar shutdown programado
- [ ] Implementar formatacao de saida (sucesso e erro)
- [ ] Integrar com estado-local (state.json, config.json)
- [ ] Testes unitarios do parsing e logica de porta
- [ ] Testes de integracao com assinador.jar
