#!/usr/bin/env bash
# Empacota um binario macOS como .dmg no padrao da especificacao:
#   <bin>-<versao>-macos-<arch>.dmg
#
# Uso:
#   ./scripts/build-dmg.sh <binario> <versao> <arch> <bin-path> <out-dir>
#
# Requer hdiutil (incluido em macOS).

set -euo pipefail

binario="${1:?binario obrigatorio (ex: assinatura)}"
versao="${2:?versao obrigatoria (ex: 1.0.0)}"
arch_alvo="${3:?arch obrigatoria (amd64 | arm64)}"
bin_path="${4:?caminho do binario obrigatorio}"
out_dir="${5:?diretorio de saida obrigatorio}"

if ! command -v hdiutil >/dev/null 2>&1; then
    echo "hdiutil nao encontrado (este script requer macOS)" >&2
    exit 1
fi

mkdir -p "$out_dir"
work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT

stage="$work_dir/stage"
mkdir -p "$stage"
cp "$bin_path" "$stage/${binario}"
chmod +x "$stage/${binario}"

cat > "$stage/README.txt" <<EOF
${binario} ${versao} (macOS ${arch_alvo})

Para usar, arraste o binario para /usr/local/bin ou execute diretamente.
Verifique a assinatura Cosign antes de executar.
EOF

saida="$out_dir/${binario}-${versao}-macos-${arch_alvo}.dmg"

hdiutil create \
    -volname "${binario} ${versao}" \
    -srcfolder "$stage" \
    -ov \
    -format UDZO \
    -fs HFS+ \
    "$saida" >/dev/null

echo "$saida"
