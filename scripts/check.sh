#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[CHECK] Go formatting"
unformatted="$(gofmt -l $(find cmd internal -name '*.go' -type f))"
if [[ -n "$unformatted" ]]; then
  printf '%s\n' "$unformatted"
  exit 1
fi

echo "[CHECK] Go modules"
go mod verify
go mod tidy -diff

echo "[CHECK] Static analysis"
go vet ./...

echo "[CHECK] Unit and WebSocket integration tests"
go test -race -cover ./...

echo "[CHECK] Shell syntax"
for script in scripts/*.sh; do
  bash -n "$script"
done

echo "[OK] All local source checks passed."
