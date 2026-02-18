#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/frontend/public"
CMD_DIR="$ROOT_DIR"

mkdir -p "$DIST_DIR"

echo "[+] Building Umbra Client (WASM)"

GOOS=js GOARCH=wasm \
go build \
	-trimpath \
	-ldflags="-s -w" \
	-o "$DIST_DIR/umbra.wasm" \
	"$CMD_DIR"

# Copy wasm_exec.js if missing or outdated
WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
cp "$WASM_EXEC" "$DIST_DIR/wasm_exec.js"

echo "[✓] WASM build complete"
