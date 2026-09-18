#!/usr/bin/env bash
set -euo pipefail

repo_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
docs_file="${repo_dir}/docs/content/overlays/eventclient.md"
temporary_file=$(mktemp --suffix=.js)

trap 'rm -f "${temporary_file}"' EXIT

cat >"${docs_file}" <<'EOF'
---
title: EventClient
weight: 10000
---

## Typed overlay development

Import `eventclient.js` directly. TypeScript-aware editors automatically use the adjacent `eventclient.d.ts` and `eventTypes.d.ts` declarations for event-specific handler autocomplete and type checking.

EOF

cd "${repo_dir}"

pnpm exec esbuild \
  ./internal/apimodules/overlays/src/eventclient.ts \
  --format=esm \
  --outfile="${temporary_file}" \
  --target=es2020

pnpx jsdoc-to-markdown \
  --files "${temporary_file}" |
  sed 's/[[:space:]]*$//' \
    >>"${docs_file}"

sed -i '${/^$/d;}' "${docs_file}"
