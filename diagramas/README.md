# Diagramas C4 — Sistema Runner

Diagramas C4 (Context Level 1 e Container Level 2) em PlantUML, conforme
exigido pela §7.5 da [especificação](../../runnerProfessor/especificacao.md).

## Arquivos

| Arquivo | Nível C4 | Conteúdo |
|---------|---------|----------|
| [`contexto.puml`](contexto.puml) | 1 — Context | Usuário, Sistema Runner, dispositivos/sistemas externos (PKCS#11, Simulador HubSaúde, Adoptium, GitHub Releases) |
| [`conteineres.puml`](conteineres.puml) | 2 — Container | CLIs `assinatura` e `simulador`, `assinador.jar`, estado local, e suas interações |

## Gerar SVG/PNG

### Local (recomendado)

Instalar PlantUML:

- macOS: `brew install plantuml`
- Linux: `apt-get install plantuml` ou `pacman -S plantuml`
- Windows: baixar `plantuml.jar` e expor no `PATH`

Depois rodar:

```bash
bash geraimagens.sh        # Unix
.\geraimagens.bat          # Windows
```

### Via Docker (sem instalar PlantUML)

```bash
docker run --rm -v "$PWD":/data plantuml/plantuml -tsvg -o /data/imagens /data/*.puml
```

Os scripts `geraimagens.sh` / `geraimagens.bat` detectam automaticamente
PlantUML local ou Docker.

### Online (preview rápido)

Cole o conteúdo do `.puml` em https://www.plantuml.com/plantuml/uml/

## Saídas geradas

Os SVGs e PNGs vão para `imagens/`. O README e a documentação técnica
referenciam esses arquivos:

- `imagens/contexto.svg`
- `imagens/conteineres.svg`
