# Entregavel: Documentacao

## Objetivo

Produzir documentacao clara para usuarios e desenvolvedores, incluindo manual de uso, documentacao tecnica, guia de instalacao e exemplos.

## Documentos Necessarios

### 1. Manual do Usuario (`docs/manual-usuario.md`)

Publico: usuario final que quer usar o CLI.

Conteudo:
- Instalacao (download do binario, verificacao com Cosign)
- Primeiros passos
- Referencia de comandos:
  - `assinatura criar` — todas as flags e exemplos
  - `assinatura validar` — todas as flags e exemplos
  - `assinatura servidor` — iniciar, parar, status
  - `simulador iniciar` — flags e exemplos
  - `simulador parar`
  - `simulador status`
- Modos de operacao (local vs HTTP)
- Solucao de problemas comuns

### 2. Documentacao Tecnica (`docs/tecnico.md`)

Publico: desenvolvedor ou integrador.

Conteudo:
- Arquitetura do sistema (diagrama de componentes)
- Fluxo de dados (criacao e validacao de assinatura)
- Contrato da API HTTP do assinador.jar
  - Endpoints, metodos, headers
  - Request/response bodies (JSON schemas)
  - Codigos de erro
- Contrato do modo CLI do assinador.jar
  - Argumentos, saida, codigos de saida
- Provisionamento do JDK (como funciona)
- Download do simulador.jar (como funciona)

### 3. Guia de Instalacao (`docs/instalacao.md`)

Publico: usuario que precisa instalar o sistema.

Conteudo:
- Pre-requisitos (nenhum, JDK provisionado automaticamente)
- Download via GitHub Releases
- Verificacao de integridade com Cosign
- Instalacao por plataforma (Windows, Linux, macOS)
- Configuracao inicial (se necessario)
- Verificacao pos-instalacao (`assinatura --versao`)

### 4. Exemplos de Uso (`docs/exemplos.md`)

Conteudo:
- Criar assinatura (modo local)
- Criar assinatura (modo HTTP)
- Validar assinatura
- Iniciar e gerenciar simulador
- Provisionar JDK manualmente (se necessario)
- Verificar artefato com Cosign

### 5. Diagramas C4 (referenciados na especificacao)

- Nivel 1: Diagrama de Contexto (ja existe referencia)
- Nivel 2: Diagrama de Conteineres (ja existe referencia)
- Atualizar/criar se necessario em `diagramas/`

## Help Integrado no CLI

Cada comando deve ter `--help` com:
- Descricao breve
- Uso (sintaxe)
- Flags disponiveis com descricao
- Exemplo de uso

```
$ assinatura criar --help

Cria uma assinatura digital simulada para o documento fornecido.

Uso:
  assinatura criar [flags]

Flags:
  --arquivo       Caminho do arquivo a assinar (obrigatorio)
  --certificado   Identificador do certificado (obrigatorio)
  --algoritmo     Algoritmo de assinatura (padrao: SHA256withRSA)
  --modo          Modo de invocacao: local ou http (padrao: http)
  --porta         Porta do servidor (padrao: 8088)

Exemplo:
  assinatura criar --arquivo documento.json --certificado meu-cert
```

## Tarefas de Implementacao

- [ ] Escrever manual do usuario
- [ ] Escrever documentacao tecnica (API, fluxos, arquitetura)
- [ ] Escrever guia de instalacao
- [ ] Escrever exemplos de uso
- [ ] Atualizar/criar diagramas C4
- [ ] Implementar --help em todos os comandos
- [ ] Revisar README.md do projeto
