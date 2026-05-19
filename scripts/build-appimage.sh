#!/usr/bin/env bash
# Empacota um binario Linux como .AppImage no padrao da especificacao:
#   <bin>-<versao>-linux-<arch>.AppImage
#
# Uso:
#   ./scripts/build-appimage.sh <binario> <versao> <arch> <bin-path> <out-dir>
#
# Requisitos: appimagetool no PATH (ou em $APPIMAGETOOL).
# Em CI sem FUSE, instale via: appimagetool-x86_64.AppImage --appimage-extract,
# e exporte APPIMAGETOOL=squashfs-root/AppRun.

set -euo pipefail

binario="${1:?binario obrigatorio (ex: assinatura)}"
versao="${2:?versao obrigatoria (ex: 1.0.0)}"
arch_alvo="${3:?arch obrigatoria (amd64 | arm64)}"
bin_path="${4:?caminho do binario obrigatorio}"
out_dir="${5:?diretorio de saida obrigatorio}"

case "$arch_alvo" in
    amd64) appimage_arch="x86_64" ;;
    arm64) appimage_arch="aarch64" ;;
    *) echo "arch nao suportada: $arch_alvo" >&2; exit 1 ;;
esac

appimagetool="${APPIMAGETOOL:-appimagetool}"
if ! command -v "$appimagetool" >/dev/null 2>&1; then
    echo "appimagetool nao encontrado no PATH (defina APPIMAGETOOL ou instale)" >&2
    exit 1
fi

mkdir -p "$out_dir"
work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT

app_dir="$work_dir/${binario}.AppDir"
mkdir -p "$app_dir/usr/bin"
cp "$bin_path" "$app_dir/usr/bin/${binario}"
chmod +x "$app_dir/usr/bin/${binario}"

cat > "$app_dir/AppRun" <<EOF
#!/bin/sh
HERE="\$(dirname "\$(readlink -f "\$0")")"
exec "\$HERE/usr/bin/${binario}" "\$@"
EOF
chmod +x "$app_dir/AppRun"

cat > "$app_dir/${binario}.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=${binario}
Exec=${binario}
Icon=${binario}
Categories=Utility;
Terminal=true
EOF

# Icone placeholder (1x1 PNG transparente em base64); AppImage exige um icone.
python3 -c "import base64,sys; sys.stdout.buffer.write(base64.b64decode('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII='))" \
    > "$app_dir/${binario}.png"

saida="$out_dir/${binario}-${versao}-linux-${arch_alvo}.AppImage"
ARCH="$appimage_arch" "$appimagetool" --no-appstream "$app_dir" "$saida" >/dev/null

echo "$saida"
