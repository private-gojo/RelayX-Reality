#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

read -r -p "Cloud Run run.app host or URL: " RUN_APP
read -r -p "DEST [googleapis.com:443]: " DEST
read -r -p "SNI [googleapis.com]: " SNI
DEST="${DEST:-googleapis.com:443}"
SNI="${SNI:-googleapis.com}"

RUN_APP="${RUN_APP#https://}"
RUN_APP="${RUN_APP#wss://}"
RUN_APP="${RUN_APP#http://}"
RUN_APP="${RUN_APP%/}"
[[ -n "$RUN_APP" && "$RUN_APP" != */* ]] || { echo "[ERROR] Enter only the Cloud Run hostname." >&2; exit 1; }
[[ "$DEST" == *:443 ]] || { echo "[ERROR] DEST must use port 443." >&2; exit 1; }
[[ -n "$SNI" && "$SNI" != *[[:space:]/]* ]] || { echo "[ERROR] Invalid SNI hostname." >&2; exit 1; }

command -v go >/dev/null 2>&1 || { echo "[ERROR] Go is required." >&2; exit 1; }
XRAY_BIN="${XRAY_BIN:-xray}"
command -v "$XRAY_BIN" >/dev/null 2>&1 || { echo "[ERROR] Xray binary not found: $XRAY_BIN" >&2; exit 1; }

export REMOTE_WS_URL="wss://$RUN_APP/reality"
export REALITY_TARGET="$DEST"
export REALITY_SERVER_NAME="$SNI"
export LOCAL_ADDR="${LOCAL_ADDR:-127.0.0.1:11000}"

mkdir -p bin
go build -trimpath -o bin/relay ./cmd/relay
./bin/relay client-config

./bin/relay client &
bridge_pid=$!
cleanup() {
  kill "$bridge_pid" 2>/dev/null || true
  wait "$bridge_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

"$XRAY_BIN" run -config client/xray-client.json
