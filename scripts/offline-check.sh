#!/usr/bin/env bash
# Does not install packages or connect to MySQL.
set -euo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ ! -f "$project_root/backend/vendor/modules.txt" ]]; then
  printf 'Missing backend/vendor. Run scripts/setup-online.sh while online first.\n' >&2
  exit 1
fi
if [[ ! -d "$project_root/frontend/node_modules" ]]; then
  printf 'Missing frontend/node_modules. Run scripts/setup-online.sh while online first.\n' >&2
  exit 1
fi

mkdir -p "$project_root/bin"
cd "$project_root/backend"
printf 'Building Go without downloading modules or toolchains...\n'
GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go build -mod=vendor -o "$project_root/bin/server" .

cd "$project_root/frontend"
printf '\nBuilding Angular using local dependencies...\n'
npm_config_offline=true NG_CLI_ANALYTICS=false npm run build
printf '\nBoth builds passed. This checks compilation; also verify login and saving a request with MySQL running.\n'
