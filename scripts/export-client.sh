#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

[[ -f .env ]] || { echo "[ERROR] .env missing; run scripts/init.sh first." >&2; exit 1; }
set -a
source .env
set +a

go run ./cmd/relay client-config
