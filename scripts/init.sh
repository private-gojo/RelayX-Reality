#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

cat > .env <<'EOF'
LOCAL_ADDR=127.0.0.1:11000
MAX_SESSIONS=256
SERVICE_NAME=relay-real
REGION=asia-east1
EOF
chmod 600 .env

echo "[OK] Local runtime settings written. Credentials are embedded in the client and container."
