# Pipeline de CI/CD

## Exigencia

CI/CD e obrigatorio pela especificacao (secao 9.4):
> "A assinatura dos artefatos DEVE ser realizada automaticamente pelo pipeline de CI/CD"

A distribuicao de binarios via GitHub Releases (secao 7) tambem depende do pipeline de release.

## Decisoes

| Decisao | Escolha |
|---------|---------|
| Plataforma | GitHub Actions |
| Linguagem CLI | Go 1.22+ |
| JDK | Temurin 21 (`actions/setup-java`) |
| Build Java | Maven (`./mvnw`) |
| Assinatura de artefatos | Cosign via `sigstore/cosign-installer@v3` |
| Release | `softprops/action-gh-release@v2` acionado por tag `v*` |

---

## Workflows

### 1. CI — Build e Testes (`ci.yml`)

**Trigger:** push e pull request em qualquer branch.

**Matrix:** `ubuntu-latest`, `windows-latest`, `macos-latest`

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test-cli:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: cd assinatura && go test ./...
      - run: cd simulador && go test ./...

  test-assinador:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '21'
      - run: cd assinador && ./mvnw verify
```

### 2. Release (`release.yml`)

**Trigger:** criacao de tag `v*` (ex: `v1.0.0`).

Gera binarios para as 3 plataformas, empacota, assina com Cosign e publica no GitHub Releases.

```yaml
name: Release

on:
  push:
    tags: ['v*']

permissions:
  id-token: write   # OIDC para Cosign
  contents: write   # Criar release e upload

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '21'

      - name: Build assinador.jar
        run: cd assinador && ./mvnw package -DskipTests

      - name: Build CLIs multiplataforma
        run: |
          VERSION=${GITHUB_REF_NAME#v}
          mkdir -p dist

          # assinatura
          GOOS=linux   GOARCH=amd64 go build -o dist/assinatura-${VERSION}-linux-amd64   ./assinatura
          GOOS=windows GOARCH=amd64 go build -o dist/assinatura-${VERSION}-windows-amd64.exe ./assinatura
          GOOS=darwin  GOARCH=amd64 go build -o dist/assinatura-${VERSION}-macos-amd64   ./assinatura

          # simulador
          GOOS=linux   GOARCH=amd64 go build -o dist/simulador-${VERSION}-linux-amd64   ./simulador
          GOOS=windows GOARCH=amd64 go build -o dist/simulador-${VERSION}-windows-amd64.exe ./simulador
          GOOS=darwin  GOARCH=amd64 go build -o dist/simulador-${VERSION}-macos-amd64   ./simulador

          # assinador.jar
          cp assinador/target/assinador.jar dist/assinador-${VERSION}.jar

      - name: Empacotamento (AppImage / dmg)
        run: |
          # TODO Sprint 5: empacotar binarios Linux em .AppImage e macOS em .dmg
          # Fallback por enquanto: renomear para o padrao da especificacao sem extensao de pacote
          echo "Empacotamento pendente — Sprint 5"

      - name: Checksums SHA256
        run: cd dist && sha256sum * > checksums-sha256.txt

      - uses: sigstore/cosign-installer@v3

      - name: Assinar artefatos com Cosign
        run: |
          for file in dist/*; do
            [[ "$file" == *.sig ]] && continue
            [[ "$file" == *.pem ]] && continue
            [[ "$file" == *checksums* ]] && continue
            cosign sign-blob --yes \
              --output-signature "${file}.sig" \
              --output-certificate "${file}.pem" \
              "$file"
          done

      - uses: softprops/action-gh-release@v2
        with:
          files: dist/*
```

### 3. Lint (Opcional — `lint.yml`)

**Trigger:** push e pull request.

```yaml
# CLI
- uses: golangci/golangci-lint-action@v4

# Java
- run: cd assinador && ./mvnw checkstyle:check
```

---

## Branches e Fluxo

```
main ──────────────────────────────────
  │
  ├── feature/us01-cli-skeleton
  ├── feature/us02-assinador-validacao
  ├── feature/us03-simulador
  └── ...
```

- **main**: sempre estavel, CI verde obrigatorio
- **feature/\***: desenvolvimento por funcionalidade
- **Tags `v*`**: disparam o workflow de release

---

## Tarefas de Implementacao

- [ ] Criar `.github/workflows/ci.yml`
- [ ] Criar `.github/workflows/release.yml`
- [ ] Validar matrix CI nas 3 plataformas com primeiro PR
- [ ] Validar assinatura Cosign com tag de teste (`v0.0.1-test`)
- [ ] Implementar empacotamento .AppImage e .dmg (Sprint 5)
- [ ] (Opcional) Configurar `golangci-lint` e `checkstyle`
