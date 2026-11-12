#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

[[ -f .env ]] || { echo "[ERROR] .env missing; run scripts/init.sh first." >&2; exit 1; }
set -a
source .env
set +a

image="${1:?Usage: bash scripts/deploy.sh DOCKERHUB_USERNAME/x-relay-real:tag}"

command -v gcloud >/dev/null 2>&1 || { echo "[ERROR] Google Cloud CLI is required." >&2; exit 1; }

: "${VLESS_UUID:?VLESS_UUID missing in .env}"
: "${REALITY_PRIVATE_KEY:?REALITY_PRIVATE_KEY missing in .env}"
: "${REALITY_SHORT_ID:?REALITY_SHORT_ID missing in .env}"
: "${BRIDGE_TOKEN:?BRIDGE_TOKEN missing in .env}"

SERVICE_NAME="${SERVICE_NAME:-relayx-reality}"
REGION="${REGION:-asia-east1}"
MAX_SESSIONS="${MAX_SESSIONS:-256}"

runtime_env="VLESS_UUID=$VLESS_UUID,REALITY_PRIVATE_KEY=$REALITY_PRIVATE_KEY,REALITY_SHORT_ID=$REALITY_SHORT_ID,REALITY_SERVER_NAME=${REALITY_SERVER_NAME:-googleapis.com},REALITY_TARGET=${REALITY_TARGET:-googleapis.com:443},REALITY_PATH=${REALITY_PATH:-/reality},BRIDGE_TOKEN=$BRIDGE_TOKEN,MAX_SESSIONS=$MAX_SESSIONS"

echo "[DEPLOY] service=$SERVICE_NAME region=$REGION image=$image"
gcloud run deploy "$SERVICE_NAME" \
  --image "$image" \
  --region "$REGION" \
  --platform managed \
  --allow-unauthenticated \
  --port 8080 \
  --timeout 3600 \
  --concurrency "$MAX_SESSIONS" \
  --cpu 2 \
  --memory 1Gi \
  --set-env-vars "$runtime_env" \
  --quiet

gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --format='value(status.url)'
