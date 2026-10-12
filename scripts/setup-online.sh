#!/usr/bin/env bash
# Run on the TARGET computer while Internet is available.
set -euo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"

for command_name in go node npm; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf 'Missing %s. Install the prerequisites from README.md.\n' "$command_name" >&2
    exit 1
  fi
done

node -e 'const [major, minor] = process.versions.node.split(".").map(Number); if (!((major === 22 && minor >= 12) || major === 24)) { console.error("Use Node.js 22.12+ in the 22.x branch, or Node.js 24.x. Current: " + process.version); process.exit(1); }'
go_version="$(go env GOVERSION)"
if [[ "$go_version" =~ ^go([0-9]+)\.([0-9]+) ]]; then
  if (( BASH_REMATCH[1] < 1 || (BASH_REMATCH[1] == 1 && BASH_REMATCH[2] < 22) )); then
    printf 'Go 1.22 or newer is required. Current: %s\n' "$go_version" >&2
    exit 1
  fi
else
  printf 'Cannot determine Go version: %s\n' "$go_version" >&2
  exit 1
fi
go version
node --version
npm --version

printf '\nDownloading Go dependencies and preparing vendor...\n'
cd "$project_root/backend"
GOTOOLCHAIN=local go mod download
GOTOOLCHAIN=local go mod vendor

printf '\nInstalling the exact npm lockfile versions and filling the local cache...\n'
cd "$project_root/frontend"
npm ci --cache "$project_root/.offline/npm" --no-audit --no-fund --fetch-retries=5

printf '\nChecking builds using the installed dependencies...\n'
bash "$project_root/scripts/offline-check.sh"
printf '\nReady. Keep backend/vendor, frontend/node_modules and .offline on THIS computer.\n'
printf 'Now configure MySQL and backend/.env according to README.md, then test with Internet disconnected.\n'
