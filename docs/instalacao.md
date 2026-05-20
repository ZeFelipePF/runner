# Guia de Instalação — Sistema Runner

Como baixar, verificar e começar a usar os CLIs `assinatura` e `simulador`.

> Já instalou? Vá para o [Manual do Usuário](manual-usuario.md).

---

## 1. Pré-requisitos

**Nenhum obrigatório.** Os binários são autocontidos e o **JDK 21 é provisionado
automaticamente** no primeiro uso (baixado da Adoptium para `~/.hubsaude/jdk/`).

Opcional, recomendado:
- [Cosign](https://docs.sigstore.dev/cosign/installation/) — para verificar a autenticidade dos artefatos.
- Conexão à internet no primeiro uso — para baixar o JDK e o `simulador.jar`
  (depois ficam em cache). Em ambientes offline, veja a seção 6.

---

## 2. Download via GitHub Releases

Baixe o artefato da sua plataforma na página de **[Releases](../../releases)**.
Substitua `x.y.z` pela versão desejada (SemVer, ex.: `1.0.0`).

| Plataforma | `assinatura` | `simulador` |
|------------|--------------|-------------|
| Windows (amd64) | `assinatura-x.y.z-windows-amd64.exe` | `simulador-x.y.z-windows-amd64.exe` |
| Linux (amd64) | `assinatura-x.y.z-linux-amd64.AppImage` | `simulador-x.y.z-linux-amd64.AppImage` |
| macOS (amd64) | `assinatura-x.y.z-macos-amd64.dmg` | `simulador-x.y.z-macos-amd64.dmg` |

Cada artefato acompanha dois arquivos de assinatura: `<artefato>.sig` e `<artefato>.pem`.
A release também publica `checksums-sha256.txt` e o `assinador-x.y.z.jar`.

---

## 3. Verificação de integridade (checksums)

Confirme que o download não corrompeu, comparando o hash SHA-256:

```bash
# Linux
sha256sum assinatura-x.y.z-linux-amd64.AppImage
# macOS
shasum -a 256 assinatura-x.y.z-macos-amd64.dmg
# Windows (PowerShell)
Get-FileHash .\assinatura-x.y.z-windows-amd64.exe -Algorithm SHA256
```

Compare a saída com a linha correspondente em `checksums-sha256.txt`.

---

## 4. Verificação de autenticidade (Cosign)

Todos os artefatos são assinados com **Cosign** (Sigstore), de forma _keyless_ via OIDC,
pelo pipeline de CI. Isso prova que o binário veio do pipeline oficial do projeto e não
foi adulterado.

```bash
cosign verify-blob \
  --certificate assinatura-x.y.z-linux-amd64.AppImage.pem \
  --signature   assinatura-x.y.z-linux-amd64.AppImage.sig \
  --certificate-identity-regexp ".*" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  assinatura-x.y.z-linux-amd64.AppImage
```

Saída esperada: `Verified OK`.

> As flags `--certificate-identity-regexp` e `--certificate-oidc-issuer` são
> **obrigatórias** nas versões atuais do Cosign. Para uma verificação mais estrita,
> troque `".*"` pela identidade exata do workflow que assinou (visível no `.pem`).

---

## 5. Instalação por plataforma

### Linux (`.AppImage`)

```bash
chmod +x assinatura-x.y.z-linux-amd64.AppImage
# Opcional: instalar no PATH com um nome curto
sudo mv assinatura-x.y.z-linux-amd64.AppImage /usr/local/bin/assinatura
assinatura --versao
```

Repita para o `simulador`. (Se faltar `libfuse2`, instale-o ou rode com `--appimage-extract-and-run`.)

### macOS (`.dmg`)

1. Abra o `.dmg` e arraste o binário para `/Applications` ou `/usr/local/bin`.
2. Na primeira execução, o Gatekeeper pode bloquear um app não notarizado:
   `Ajustes do Sistema → Privacidade e Segurança → Abrir Assim Mesmo`,
   ou via terminal:
   ```bash
   xattr -d com.apple.quarantine /usr/local/bin/assinatura
   assinatura --versao
   ```

### Windows (`.exe`)

1. Mova o `.exe` para uma pasta no `PATH` (ex.: `C:\Users\<você>\bin`) ou para a pasta do projeto.
2. O SmartScreen pode avisar sobre app desconhecido: `Mais informações → Executar assim mesmo`.
3. Verifique:
   ```powershell
   .\assinatura-x.y.z-windows-amd64.exe --versao
   ```

---

## 6. Configuração inicial

Nenhuma configuração é necessária para começar. O Runner cria `~/.hubsaude/`
automaticamente (no Windows: `C:\Users\<você>\.hubsaude\`).

### Ambientes offline / sem acesso à Adoptium

- **JDK**: instale o Java 21 manualmente e exponha-o via `JAVA_HOME` ou `PATH`.
  O Runner detecta nessa ordem e dispensa o download.
- **`simulador.jar`**: use `simulador iniciar --source file:///caminho/simulador.jar`.
- **`assinador.jar`**: aponte com `HUBSAUDE_ASSINADOR_JAR=/caminho/assinador.jar`
  (o jar é publicado em cada release).

---

## 7. Verificação pós-instalação

```bash
assinatura --versao        # ex.: assinatura 1.0.0 (<git-commit>)
simulador --versao         # ex.: simulador 1.0.0 (<git-commit>)

assinatura --help          # lista subcomandos
simulador status           # deve responder { "registrado": false, "running": false }
```

Se os quatro comandos respondem, a instalação está concluída.
Siga para os [Exemplos de Uso](exemplos.md) ou o [Manual do Usuário](manual-usuario.md).

---

## 8. Atualização e desinstalação

- **Atualizar**: baixe a nova versão na Releases, verifique (seções 3–4) e substitua o binário.
- **Desinstalar**: remova os binários e, se quiser limpar o estado, a pasta `~/.hubsaude/`
  (apague-a apenas após `assinatura servidor parar` e `simulador parar`).
