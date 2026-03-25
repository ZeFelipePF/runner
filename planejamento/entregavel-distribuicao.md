# Entregavel: Distribuicao Multiplataforma e Assinatura de Artefatos (US-05, S9)

## Objetivo

Gerar binarios pre-compilados para Windows, Linux e macOS, distribui-los via GitHub Releases com checksums SHA256 e assinatura criptografica via Cosign.

## Historias Relacionadas

- **US-05**: Disponibilizar binarios multiplataforma
- **Secao 9**: Integridade e assinatura de artefatos

## Artefatos por Release

Para cada versao (ex: 1.0.0):

```
assinatura-1.0.0-windows-amd64.exe
assinatura-1.0.0-windows-amd64.exe.sig
assinatura-1.0.0-windows-amd64.exe.pem
assinatura-1.0.0-linux-amd64.AppImage
assinatura-1.0.0-linux-amd64.AppImage.sig
assinatura-1.0.0-linux-amd64.AppImage.pem
assinatura-1.0.0-macos-amd64.dmg
assinatura-1.0.0-macos-amd64.dmg.sig
assinatura-1.0.0-macos-amd64.dmg.pem
simulador-1.0.0-windows-amd64.exe
simulador-1.0.0-windows-amd64.exe.sig
simulador-1.0.0-windows-amd64.exe.pem
simulador-1.0.0-linux-amd64.AppImage
simulador-1.0.0-linux-amd64.AppImage.sig
simulador-1.0.0-linux-amd64.AppImage.pem
simulador-1.0.0-macos-amd64.dmg
simulador-1.0.0-macos-amd64.dmg.sig
simulador-1.0.0-macos-amd64.dmg.pem
checksums-sha256.txt
```

## Cross-Compilation

Go suporta cross-compilation nativa via variaveis de ambiente, sem dependencias externas.

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o assinatura-windows-amd64.exe ./assinatura

# Linux
GOOS=linux GOARCH=amd64 go build -o assinatura-linux-amd64 ./assinatura

# macOS
GOOS=darwin GOARCH=amd64 go build -o assinatura-macos-amd64 ./assinatura
```

O mesmo se aplica ao CLI `simulador` — trocar `./assinatura` por `./simulador`.

`goreleaser` automatiza build, checksum, changelog e upload para GitHub Releases.

## Empacotamento

### Alternativas para empacotamento multiplataforma

| Formato     | Plataforma | Ferramenta                      | Complexidade |
| ----------- | ---------- | ------------------------------- | ------------ |
| `.exe`      | Windows    | Binario direto (Go/Rust)        | Baixa        |
| `.AppImage` | Linux      | `appimagetool` ou `linuxdeploy` | Media        |
| `.dmg`      | macOS      | `create-dmg` ou `hdiutil`       | Media-Alta   |

### Opcao simplificada (fallback)

Se o empacotamento em .AppImage e .dmg for muito complexo para o prazo:

- Distribuir binarios nativos (sem empacotamento especial)
- Nomear como `assinatura-1.0.0-linux-amd64` (sem extensao)
- Documentar instrucoes de uso direto

### GoReleaser

Ferramenta que automatiza: build, checksum, changelog e upload para GitHub Releases.

```yaml
# .goreleaser.yml (exemplo simplificado)
builds:
  - id: assinatura
    main: ./assinatura
    binary: assinatura
    goos: [linux, windows, darwin]
    goarch: [amd64]
  - id: simulador
    main: ./simulador
    binary: simulador
    goos: [linux, windows, darwin]
    goarch: [amd64]
```

## Cosign (Sigstore)

### Fluxo de assinatura no CI/CD

```
1. Build dos binarios
2. Para cada artefato:
   a. cosign sign-blob --yes --oidc-issuer=https://token.actions.githubusercontent.com <artefato>
   b. Gera: <artefato>.sig e <artefato>.pem
3. Upload de todos os arquivos para GitHub Release
```

### Comando Cosign

```bash
# Assinar
cosign sign-blob \
  --yes \
  --output-signature artefato.sig \
  --output-certificate artefato.pem \
  artefato

# Verificar
cosign verify-blob \
  --certificate artefato.pem \
  --signature artefato.sig \
  artefato
```

### Identidade OIDC no GitHub Actions

O GitHub Actions fornece tokens OIDC nativamente. A identidade sera o workflow do repositorio.

```yaml
# Permissoes necessarias no workflow
permissions:
  id-token: write # Para OIDC
  contents: write # Para criar release
```

## Versionamento

- Seguir SemVer: `MAJOR.MINOR.PATCH`
- Tags git: `v1.0.0`, `v1.1.0`, etc.
- Release criada automaticamente ao criar tag

## Tarefas de Implementacao

- [ ] Configurar cross-compilation na linguagem escolhida
- [ ] Configurar empacotamento (.exe, .AppImage, .dmg) ou fallback
- [ ] Gerar checksums SHA256 automaticamente
- [ ] Instalar e configurar Cosign no pipeline
- [ ] Gerar .sig e .pem para cada artefato
- [ ] Configurar GitHub Release automatica (via tag)
- [ ] Testar fluxo completo de release
- [ ] Documentar processo de verificacao para usuarios
