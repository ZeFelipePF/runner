# Entregavel: CLI simulador (US-03)

## Objetivo

CLI multiplataforma que gerencia o ciclo de vida do Simulador do HubSaude (simulador.jar), permitindo iniciar, parar e monitorar sem conhecer comandos Java.

## Historias Relacionadas

- **US-03**: Gerenciar Ciclo de Vida do Simulador do HubSaude
- **US-04**: Provisionar JDK automaticamente (dependencia)

## Importante

O `simulador.jar` **nao e desenvolvido** neste projeto. Ele deve ser:
1. Obtido dinamicamente via GitHub Releases do repositorio da disciplina
2. Cacheado localmente em `~/.hubsaude/simulador/` para evitar downloads repetidos
3. Versao verificada antes de cada `iniciar`

## Comandos Planejados

```
simulador iniciar [flags]     # Baixar (se necessario) e iniciar o simulador.jar
simulador parar [flags]       # Parar o simulador.jar em execucao
simulador status              # Exibir status do simulador (rodando ou nao)
simulador --help              # Ajuda
simulador --versao            # Versao do CLI
```

## Flags

| Flag | Comando | Descricao | Padrao |
|------|---------|-----------|--------|
| `--porta` | iniciar | Porta para o simulador (0 = auto-detectar) | `0` (auto) |
| `--source` | iniciar | URL alternativa para obter o simulador.jar | URL interna do CLI |
| `--versao-simulador` | iniciar | Versao especifica para baixar | `latest` |
| `--verbose` | todos | Saida detalhada | `false` |

### Flag `--source`

A URL de download do simulador.jar e **hardwired** internamente no CLI (apontando para o GitHub Releases da disciplina). Porem, `--source` permite sobrescrever essa URL sem necessidade de atualizar o binario:

```bash
# Uso padrao (URL interna do CLI)
simulador iniciar

# URL alternativa (ex: ambiente de homologacao, servidor local)
simulador iniciar --source http://servidor-local/simulador.jar

# Arquivo local
simulador iniciar --source file:///tmp/simulador-dev.jar
```

Quando `--source` e fornecido, a verificacao de versao via GitHub Releases e ignorada.

## Fluxo: `simulador iniciar`

Ver `startup.md` para o fluxo completo de inicializacao. Apos o startup:

```
1. (Startup: garantir ~/.hubsaude/, JDK, simulador.jar atualizado)
2. Determinar porta
   a. Se --porta especificada: tentar usar essa porta
      └── Se ocupada: erro informando o conflito
   b. Se --porta nao especificada (padrao): auto-detectar
      └── Tentar porta padrao (config.json ou 9090)
      └── Se ocupada: tentar proximas portas (9091, 9092, ...)
      └── Informar ao usuario qual porta foi escolhida
3. Iniciar simulador.jar como processo em background
   exec: <java_path> -jar ~/.hubsaude/simulador/simulador.jar --port <porta>
4. Aguardar health check (retry com timeout, ex: 10x com 500ms)
5. Gravar PID + porta em ~/.hubsaude/state.json
6. Exibir confirmacao:
   [✓] Simulador iniciado em http://localhost:<porta>
```

## Fluxo: `simulador parar`

```
1. Ler state.json: obter PID e porta do simulador
   └── Se nao registrado: informar que nao esta em execucao
2. Verificar se processo com esse PID existe
   └── Se nao existe: limpar state.json, informar que ja foi encerrado
3. Enviar sinal de parada
   a. Tentar HTTP POST /shutdown (graceful)
   b. Se nao responder em timeout: enviar SIGTERM
   c. Se ainda nao parar: informar e sugerir kill manual
4. Remover entrada do simulador em state.json
5. Exibir confirmacao
```

## Fluxo: `simulador status`

```
1. Ler state.json
2. Se entrada do simulador existe:
   a. Verificar se PID ainda esta ativo
   b. Se ativo:
      - Exibir: PID, porta, tempo em execucao, versao
   c. Se nao ativo (processo morreu):
      - Limpar state.json
      - Exibir: "Simulador nao esta em execucao (processo encerrado)"
3. Se sem entrada:
   - Exibir versao local se disponivel, ou "nao baixado"
   - Exibir: "Simulador nao esta em execucao"
```

## Download do simulador.jar

### URL interna (hardwired)

O CLI mantem internamente a URL padrao do simulador.jar. Exemplo:

```go
const SimuladorDefaultURL = "https://github.com/<org>/runner/releases/latest/download/simulador.jar"
```

### Estrategia de download e cache

1. Ler `~/.hubsaude/simulador/simulador-meta.json`
2. Se nao existe → baixar da URL (interna ou `--source`)
3. Se existe:
   - Se `--source` foi fornecido → baixar da URL indicada
   - Senao → consultar GitHub Releases API para versao mais recente
     - Se versao local == mais recente → usar cache
     - Se versao local != mais recente → baixar e atualizar cache
4. Salvar `simulador-meta.json` com versao, sha256, data, sourceUrl usada

### Verificacao de integridade

Apos download, verificar SHA256 do arquivo contra o valor publicado na release.

## Tratamento de Erros

| Cenario | Comportamento |
|---------|---------------|
| Sem conexao para download | Usar versao local se existir; erro se nao houver cache |
| GitHub Releases indisponivel | Usar versao local em cache |
| --source retorna 404 | Erro com URL fornecida |
| Porta explicita ocupada | Erro informando processo que ocupa a porta |
| Porta auto → todas ocupadas | Erro: "nenhuma porta disponivel no range X-Y" |
| Simulador nao responde no health check | Timeout + exibir saida do processo (stderr) |
| JDK nao disponivel | Provisionar automaticamente (startup) |
| simulador.jar corrompido (sha256 falha) | Apagar e baixar novamente |

## Tarefas de Implementacao

- [ ] Configurar projeto (mesmo workspace do assinatura)
- [ ] Implementar URL hardwired + flag --source
- [ ] Implementar comando `iniciar` com auto-deteccao de porta
- [ ] Implementar comando `parar` (graceful HTTP + SIGTERM fallback)
- [ ] Implementar comando `status`
- [ ] Implementar download via GitHub Releases API
- [ ] Implementar verificacao de versao e cache
- [ ] Implementar verificacao de integridade SHA256
- [ ] Integrar com estado-local (state.json)
- [ ] Integrar com provisionamento JDK
- [ ] Testes unitarios
- [ ] Testes de integracao
