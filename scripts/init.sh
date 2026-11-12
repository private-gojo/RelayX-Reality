#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -e .env ]]; then
  echo "[INFO] .env already exists; credentials were not regenerated."
  exit 0
fi

command -v openssl >/dev/null 2>&1 || { echo "[ERROR] openssl is required." >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "[ERROR] python3 is required." >&2; exit 1; }

key_file="$(mktemp)"
trap 'rm -f "$key_file"' EXIT

openssl genpkey -algorithm X25519 -out "$key_file" >/dev/null 2>&1
private_key="$(openssl pkey -in "$key_file" -outform DER | tail -c 32 | python3 -c 'import base64,sys; print(base64.urlsafe_b64encode(sys.stdin.buffer.read()).decode().rstrip("="))')"
uuid="$(python3 -c 'import uuid; print(uuid.uuid4())')"
short_id="$(openssl rand -hex 8)"
bridge_token="$(openssl rand -hex 32)"

cat > .env <<EOF
VLESS_UUID=$uuid
REALITY_PRIVATE_KEY=$private_key
REALITY_SHORT_ID=$short_id
REALITY_SERVER_NAME=googleapis.com
REALITY_TARGET=googleapis.com:443
REALITY_PATH=/reality
BRIDGE_TOKEN=$bridge_token
REMOTE_WS_URL=
LOCAL_ADDR=127.0.0.1:11000
MAX_SESSIONS=256
SERVICE_NAME=relayx-reality
REGION=asia-east1
EOF
chmod 600 .env

echo "[OK] Static credentials generated in .env (mode 600)."
echo "[OK] Re-running this script will not overwrite existing credentials."
