# Sequencia de Startup dos CLIs

## Objetivo

Definir o fluxo de inicializacao que cada CLI executa antes de processar o comando do usuario. Uma sequencia bem projetada e fundamental para a percecao de qualidade pelo usuario: falhas silenciosas ou mensagens confusas nessa fase prejudicam a experiencia.

## Principios

- **Fail fast:** detectar problemas cedo e informar claramente
- **Idempotente:** executar o startup multiplas vezes nao causa efeitos colaterais
- **Progresso visivel:** o usuario ve o que esta acontecendo em operacoes longas (download, extracao)
- **Graceful degradation:** se um passo opcional falha, continuar quando possivel

---

## Startup do CLI `assinatura`

```
assinatura <comando> [flags]
          │
          ▼
┌─────────────────────────────────────────┐
│ 1. Garantir diretorio ~/.hubsaude/      │
│    Existe? ok. Nao existe? Criar.       │
└─────────────────────┬───────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│ 2. Carregar state.json e config.json    │
│    Nao existe? Inicializar com defaults │
└─────────────────────┬───────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│ 3. Verificar JDK disponivel             │
│                                         │
│  a. Checar JAVA_HOME                    │
│  b. Checar `java` no PATH               │
│  c. Checar ~/.hubsaude/jdk/ (local)     │
│                                         │
│  Se encontrado e versao ok → usar       │
│  Se nao encontrado:                     │
│    → Informar: "JDK nao encontrado,     │
│      baixando JDK 21..."                │
│    → Download + extracao                │
│    → Atualizar config.json com caminho  │
└─────────────────────┬───────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│ 4. Verificar modo de invocacao          │
│                                         │
│  Se --modo local:                       │
│    → Nao precisa do servidor, prosseguir│
│                                         │
│  Se --modo http (padrao):               │
│    → Verificar se ha assinador em       │
│      execucao (state.json + PID)        │
│                                         │
│    Se processo valido encontrado:       │
│      → Reusar instancia existente       │
│                                         │
│    Se nao encontrado:                   │
│      → Encontrar porta disponivel       │
│        (porta padrao ou proxima livre)  │
│      → Iniciar assinador.jar            │
│      → Aguardar health check (/health)  │
│      → Gravar PID + porta em state.json │
└─────────────────────┬───────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│ 5. Executar comando do usuario          │
└─────────────────────────────────────────┘
```

---

## Startup do CLI `simulador`

```
simulador <comando> [flags]
          │
          ▼
┌─────────────────────────────────────────┐
│ 1. Garantir diretorio ~/.hubsaude/      │
└─────────────────────┬───────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│ 2. Carregar state.json e config.json    │
└─────────────────────┬───────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│ 3. Verificar JDK (mesmo fluxo acima)    │
└─────────────────────┬───────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│ 4. (Apenas para `iniciar`)              │
│    Verificar simulador.jar local        │
│                                         │
│  Ler simulador-meta.json:               │
│    Nao existe → baixar versao latest    │
│    Existe → comparar versao com GitHub  │
│      Se desatualizado → baixar nova     │
│      Se atualizado → usar local         │
│                                         │
│  Se --source fornecido:                 │
│    → Baixar da URL especificada         │
│    → Ignorar verificacao de versao      │
└─────────────────────┬───────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│ 5. Executar comando do usuario          │
└─────────────────────────────────────────┘
```

---

## Mensagens de Progresso

O usuario deve ver o que esta acontecendo. Exemplos de saida esperada:

```
$ assinatura criar --arquivo doc.json

[✓] JDK 21 encontrado em /home/user/.hubsaude/jdk/temurin-21.0.3
[✓] Assinador em execucao na porta 8088 (PID 12345)
Assinatura criada com sucesso.
```

```
$ assinatura criar --arquivo doc.json

[i] JDK nao encontrado. Baixando JDK 21 (Temurin)...
[✓] JDK baixado e configurado em ~/.hubsaude/jdk/
[i] Assinador nao esta em execucao. Iniciando na porta 8088...
[✓] Assinador pronto.
Assinatura criada com sucesso.
```

```
$ simulador iniciar

[i] Verificando versao do simulador...
[i] Nova versao disponivel (0.0.2). Baixando...
[✓] simulador.jar atualizado.
[i] Porta 9090 disponivel.
[✓] Simulador iniciado em http://localhost:9090
```

---

## Comportamento em Caso de Falha no Startup

| Falha | Comportamento |
|-------|---------------|
| Nao consegue criar `~/.hubsaude/` | Erro fatal: permissao negada + instrucoes |
| JDK download falha (sem internet) | Erro com instrucoes de instalacao manual |
| Porta padrao ocupada por outro app | Auto-detectar proxima porta livre (ver `estado-local.md`) |
| Assinador nao responde apos iniciar | Timeout + erro com log de saida do processo |
| simulador.jar corrompido | Remover e baixar novamente |

---

## Tarefas de Implementacao

- [ ] Implementar funcao de startup em cada CLI
- [ ] Implementar mensagens de progresso (com flag --quiet para suprimir)
- [ ] Implementar deteccao de porta disponivel automatica
- [ ] Implementar health check com retry (ex: 10 tentativas, 500ms intervalo)
- [ ] Implementar limpeza de state.json ao detectar PID invalido
- [ ] Testes unitarios do fluxo de startup (com mocks de filesystem e processo)
