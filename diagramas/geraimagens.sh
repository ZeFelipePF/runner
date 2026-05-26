#!/usr/bin/env bash
# Renderiza os diagramas C4 (.puml) em SVG/PNG.
# Requer: plantuml (https://plantuml.com/download) ou Docker.

set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
OUT="$DIR/imagens"
mkdir -p "$OUT"

render_local() {
    plantuml -tsvg -o "$OUT" "$DIR"/*.puml
    plantuml -tpng -o "$OUT" "$DIR"/*.puml
}

render_docker() {
    docker run --rm -v "$DIR":/data plantuml/plantuml -tsvg -o /data/imagens /data/*.puml
    docker run --rm -v "$DIR":/data plantuml/plantuml -tpng -o /data/imagens /data/*.puml
}

if command -v plantuml >/dev/null 2>&1; then
    echo "Usando plantuml local..."
    render_local
elif command -v docker >/dev/null 2>&1; then
    echo "Usando plantuml via Docker..."
    render_docker
else
    echo "ERRO: instale plantuml (brew install plantuml) ou docker." >&2
    exit 1
fi

echo "OK. SVGs/PNGs em $OUT"
