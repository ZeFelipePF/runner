#!/usr/bin/env bash
# Renomeia um binario Windows produzido pelo goreleaser para o padrao da especificacao:
#   <bin>-<versao>-windows-<arch>.exe
#
# Uso:
#   ./scripts/rename-windows.sh <binario> <versao> <arch> <bin-path> <out-dir>

set -euo pipefail

binario="${1:?binario obrigatorio}"
versao="${2:?versao obrigatoria}"
arch_alvo="${3:?arch obrigatoria}"
bin_path="${4:?caminho do binario obrigatorio}"
out_dir="${5:?diretorio de saida obrigatorio}"

mkdir -p "$out_dir"
saida="$out_dir/${binario}-${versao}-windows-${arch_alvo}.exe"
cp "$bin_path" "$saida"
echo "$saida"
