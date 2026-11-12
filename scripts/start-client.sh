#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

[[ -f .env ]] || { echo "[ERROR] .env missing; run scripts/init.sh first." >&2; exit 1; }
set -a
source .env
set +a

: "${REMOTE_WS_URL:?Set REMOTE_WS_URL in .env after Cloud Run exists}"
command -v go >/dev/null 2>&1 || { echo "[ERROR] Go is required." >&2; exit 1; }
XRAY_BIN="${XRAY_BIN:-xray}"
command -v "$XRAY_BIN" >/dev/null 2>&1 || { echo "[ERROR] Xray binary not found: $XRAY_BIN" >&2; exit 1; }

mkdir -p bin
go build -trimpath -o bin/relay ./cmd/relay
go run ./cmd/relay client-config

./bin/relay client &
bridge_pid=$!
cleanup() {
  kill "$bridge_pid" 2>/dev/null || true
  wait "$bridge_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

"$XRAY_BIN" run -config client/xray-client.json
