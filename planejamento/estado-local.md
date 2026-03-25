# Estado Local — "Banco de Dados" dos CLIs

## Objetivo

Os CLIs precisam persistir informacoes entre execucoes: onde esta o JDK baixado, qual processo esta rodando, em qual porta, qual o PID. Este documento define a estrutura de armazenamento local que serve como "banco de dados" para ambos os CLIs.

## Diretorio Raiz

```
~/.hubsaude/          (Linux e macOS)
%APPDATA%\.hubsaude\  (Windows)
```

O diretorio e criado automaticamente na primeira execucao de qualquer CLI.

## Estrutura de Diretorios

```
~/.hubsaude/
├── jdk/                        # Runtime Java provisionado
│   └── temurin-21.0.3/
│       ├── bin/java
│       └── ...
├── simulador/                  # simulador.jar e metadados
│   ├── simulador.jar
│   └── simulador-meta.json
├── assinador/                  # assinador.jar (se obtido via download)
│   └── assinador.jar
├── state.json                  # Estado dos processos em execucao
└── config.json                 # Configuracao do usuario (portas, preferencias)
```

## Arquivo: state.json

Armazena informacoes sobre processos Java atualmente em execucao (ou da ultima execucao).

```json
{
  "assinador": {
    "pid": 12345,
    "porta": 8088,
    "iniciadoEm": "2026-03-24T10:00:00Z",
    "javaPath": "/home/usuario/.hubsaude/jdk/temurin-21.0.3/bin/java"
  },
  "simulador": {
    "pid": 12346,
    "porta": 9090,
    "iniciadoEm": "2026-03-24T10:01:00Z",
    "versao": "0.0.1-SNAPSHOT",
    "javaPath": "/home/usuario/.hubsaude/jdk/temurin-21.0.3/bin/java"
  }
}
```

**Campos ausentes ou null** indicam que o processo nao esta registrado como em execucao.

### Ciclo de vida do state.json

| Evento | Acao no state.json |
|--------|-------------------|
| Processo iniciado com sucesso | Gravar PID, porta, timestamp |
| Processo parado | Remover entrada ou zerar PID |
| CLI inicia e processo nao existe mais | Limpar entrada obsoleta |
| Maquina reiniciada | PIDs invalidos serao detectados e limpos |

**Verificacao de validade:** antes de usar um PID do state.json, verificar se o processo ainda existe via `/proc/<PID>` (Linux/macOS) ou `tasklist` (Windows).

## Arquivo: config.json

Configuracoes persistidas do usuario, sobrescrevem os padroes do CLI.

```json
{
  "assinador": {
    "portaPadrao": 8088,
    "timeoutShutdownSegundos": 30
  },
  "simulador": {
    "portaPadrao": 9090,
    "sourceUrl": "https://github.com/disciplina/runner/releases/latest/download/simulador.jar"
  },
  "jdk": {
    "versaoMinima": 21,
    "distribuicao": "temurin"
  }
}
```

## Arquivo: simulador-meta.json

Metadados sobre a versao do simulador.jar baixado.

```json
{
  "versao": "0.0.1-SNAPSHOT",
  "dataDownload": "2026-03-24T10:00:00Z",
  "sha256": "abc123def456...",
  "sourceUrl": "https://github.com/.../simulador.jar"
}
```

## Uso pelos CLIs

### Leitura no startup

Todo CLI le o estado local antes de executar qualquer comando:

```
1. Verificar se ~/.hubsaude/ existe → criar se nao existir
2. Ler state.json → verificar processos ativos
3. Ler config.json → aplicar configuracoes do usuario
```

### Escrita ao iniciar processo

```
1. Iniciar processo Java
2. Aguardar processo responder (health check)
3. Gravar PID + porta + timestamp em state.json
```

### Limpeza automatica

Se state.json contem um PID e o processo nao existe mais:
- Remover a entrada do state.json silenciosamente
- Nao considerar erro — maquina pode ter sido reiniciada

## Tarefas de Implementacao

- [ ] Implementar criacao do diretorio `~/.hubsaude/` no primeiro uso
- [ ] Implementar leitura e escrita de `state.json`
- [ ] Implementar validacao de PID antes de usar (processo ainda existe?)
- [ ] Implementar limpeza de entradas obsoletas no state.json
- [ ] Implementar leitura de `config.json` com fallback para defaults
- [ ] Implementar escrita de metadados do simulador.jar
- [ ] Testes unitarios com filesystem mockado (tmpdir)
