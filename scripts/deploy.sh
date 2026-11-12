#!/usr/bin/env bash
set -euo pipefail

image="${1:?Usage: bash scripts/deploy.sh DOCKERHUB_USERNAME/x-relay-real:tag}"
command -v gcloud >/dev/null 2>&1 || { echo "[ERROR] Google Cloud CLI is required." >&2; exit 1; }

SERVICE_NAME="${SERVICE_NAME:-relay-real}"
REGION="${REGION:-asia-east1}"
MAX_SESSIONS="${MAX_SESSIONS:-256}"

echo "[DEPLOY] service=$SERVICE_NAME region=$REGION image=$image"
echo "[DEPLOY] REALITY target and SNI are fixed in the application."

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
  --quiet

gcloud run services describe "$SERVICE_NAME" \
  --region "$REGION" \
  --format='value(status.url)'
