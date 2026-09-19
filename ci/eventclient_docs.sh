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

### Bundled TypeScript overlays

For overlays built locally with TypeScript, Vue, or another bundler, install the EventClient archive matching your bot release by adding it to your `package.json`. Replace `<version>` with the bot version in both places:

```json
{
  "dependencies": {
    "@luzifer/twitch-bot-eventclient": "https://github.com/Luzifer/twitch-bot/releases/download/v<version>/twitch-bot-eventclient-<version>.tgz"
  }
}
```

The package contains the EventClient and its event type declarations:

```typescript
import EventClient from '@luzifer/twitch-bot-eventclient'
import type { EventSocketMessage } from '@luzifer/twitch-bot-eventclient/event-types'
```

Your bundler includes the EventClient in the generated overlay bundle, so the overlay does not need to import the `eventclient.js` served by the bot.

EventClient releases within the same major version are intended to remain compatible. Keeping the EventClient version in sync with the bot version is recommended so the bundled client matches the running bot.

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
