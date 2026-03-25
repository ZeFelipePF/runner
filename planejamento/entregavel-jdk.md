# Entregavel: Provisionamento Automatico do JDK (US-04)

## Objetivo

Detectar se o JDK na versao exigida esta presente na maquina e, quando ausente, baixar e configurar automaticamente para uso pelo assinador.jar e simulador.jar.

## Historias Relacionadas

- **US-04**: Provisionar JDK Automaticamente

## Decisoes

| Decisao | Escolha |
|---------|---------|
| Distribuicao | Eclipse Temurin (Adoptium) |
| API de download | `https://api.adoptium.net/v3/` |
| Versao minima | JDK 21 |
| Diretorio de instalacao | `~/.hubsaude/jdk/` (Linux/macOS) / `%APPDATA%\.hubsaude\jdk\` (Windows) |
| Reutilizacao entre CLIs | Codigo duplicado em `assinatura` e `simulador` |

---

## Fluxo de Deteccao

```
1. Verificar variavel JAVA_HOME
   └── Se definida: verificar versao
       └── Se >= 21: usar
       └── Se < 21: continuar busca
2. Verificar `java` no PATH
   └── Se encontrado: verificar versao
       └── Se >= 21: usar
       └── Se < 21: continuar
3. Verificar JDK local provisionado em ~/.hubsaude/jdk/
   └── Se encontrado e >= 21: usar
4. Nenhum encontrado: baixar automaticamente via Adoptium
```

## Deteccao de Versao

```bash
java -version
# Saida: openjdk version "21.0.3" 2024-04-16
# Parsear versao major (21) e aceitar se >= 21
```

Observacao: `java -version` escreve em stderr, nao stdout.

## Download via Adoptium

### Endpoint

```
GET https://api.adoptium.net/v3/assets/latest/21/hotspot
    ?architecture=x64
    &image_type=jdk
    &os={linux|mac|windows}
    &vendor=eclipse
```

### Fluxo de download

```
1. Detectar SO (linux/mac/windows) e arquitetura (x64)
2. Consultar API Adoptium para obter URL do asset
3. Baixar arquivo:
   - Linux/macOS: .tar.gz
   - Windows: .zip
4. Extrair em ~/.hubsaude/jdk/temurin-21.x.x/
5. Salvar metadados (versao, data, caminho do java executavel)
6. Retornar caminho absoluto do executavel java
```

## Diretorio de Instalacao

```
~/.hubsaude/jdk/
└── temurin-21.0.3/
    ├── bin/
    │   └── java        (java.exe no Windows)
    ├── lib/
    └── ...
```

O caminho do executavel e gravado em `~/.hubsaude/config.json` apos o primeiro download.

## Tratamento de Erros

| Cenario | Comportamento |
|---------|---------------|
| Sem internet | Erro com instrucoes de instalacao manual do JDK 21 |
| API Adoptium fora do ar | Erro com link direto para adoptium.net |
| Espaco insuficiente em disco | Verificar antes de baixar; informar espaco necessario (~200MB) |
| Extracao falha | Limpar arquivos parciais e exibir erro |
| SO/Arquitetura nao suportada | Erro informativo (suporte: linux/mac/windows + x64) |

## Tarefas de Implementacao

- [ ] Implementar deteccao de JDK (JAVA_HOME, PATH, local)
- [ ] Implementar parsing de versao do Java (stderr)
- [ ] Implementar consulta a API Adoptium
- [ ] Implementar download de .tar.gz e .zip
- [ ] Implementar extracao multiplataforma
- [ ] Gravar caminho do java em config.json apos download
- [ ] Testes unitarios (mock de download via interface)
- [ ] Testar nos 3 SOs via CI matrix
