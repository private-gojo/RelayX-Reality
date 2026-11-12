#!/usr/bin/env bash
set -euo pipefail

image="${1:?Usage: bash scripts/deploy.sh DOCKERHUB_USERNAME/x-relay-real:tag}"
command -v gcloud >/dev/null 2>&1 || { echo "[ERROR] Google Cloud CLI is required." >&2; exit 1; }

SERVICE_NAME="${SERVICE_NAME:-relay-real}"
REGION="${REGION:-asia-east1}"
MAX_SESSIONS="${MAX_SESSIONS:-256}"

read -r -p "REALITY DEST [googleapis.com:443]: " REALITY_TARGET
read -r -p "REALITY SNI [googleapis.com]: " REALITY_SERVER_NAME
REALITY_TARGET="${REALITY_TARGET:-googleapis.com:443}"
REALITY_SERVER_NAME="${REALITY_SERVER_NAME:-googleapis.com}"
[[ "$REALITY_TARGET" == *:443 ]] || { echo "[ERROR] DEST must use port 443." >&2; exit 1; }
[[ -n "$REALITY_SERVER_NAME" && "$REALITY_SERVER_NAME" != *[[:space:]/]* ]] || { echo "[ERROR] Invalid SNI hostname." >&2; exit 1; }

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
  --set-env-vars "REALITY_TARGET=$REALITY_TARGET,REALITY_SERVER_NAME=$REALITY_SERVER_NAME" \
  --quiet

gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --format='value(status.url)'
