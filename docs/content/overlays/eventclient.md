---
title: EventClient
weight: 10000
---

> [!TIP]
> The EventClient connects an overlay to the bot, dispatches incoming events to handler functions, and provides helpers for replaying events and rendering bot templates.

## Loading the EventClient

### Directly from the bot

The bot serves `eventclient.js` from the overlays directory. Import it from an overlay using a module script:

```html
<script type="module">
  import EventClient from './eventclient.js'

  const client = new EventClient({
    handlers: {},
  })
</script>
```

### Bundling with TypeScript

For overlays developed locally with TypeScript or Vue and built with a bundler, add the EventClient archive matching your bot release to your `package.json`. Replace `<version>` with the bot version in both places:

```json
{
  "dependencies": {
    "@luzifer/twitch-bot-eventclient": "https://github.com/Luzifer/twitch-bot/releases/download/v<version>/twitch-bot-eventclient-<version>.tgz"
  }
}
```

Import the client and any event types required by the overlay:

```typescript
import EventClient from '@luzifer/twitch-bot-eventclient'
import type {
  FollowSocketMessage,
} from '@luzifer/twitch-bot-eventclient/event-types'

const client = new EventClient({
  handlers: {},
})
```

The bundler includes the EventClient in the generated overlay bundle, so the overlay does not need to load the `eventclient.js` served by the bot.

EventClient releases within the same major version are intended to remain compatible. Keeping the EventClient version in sync with the bot version is recommended so the bundled client matches the running bot.

## Connecting and handling events

Create one client when the overlay starts. The token is best supplied through the URL hash instead of being included in the overlay source:

```text
https://your-bot.example.com/overlays/my-overlay.html#token=YOUR_TOKEN
```

Register handlers by event type. The `_` handler receives every event and runs in addition to a matching event-specific handler:

```javascript
const client = new EventClient({
  channel: '#mychannel',
  handlers: {
    follow: event => {
      console.log(`${event.fields.user} followed at ${event.time}`)
    },
    _: event => {
      console.debug(`Received ${event.type}`, event.fields)
    },
  },
  maxReplayAge: 24,
  replay: true,
})
```

The handler receives an event with these common properties:

| Property | Description |
| --- | --- |
| `event_id` | Unique event ID, which can be passed to `replayEvent` |
| `type` | Event type used to select the handler |
| `fields` | Event-specific data; see [Available Events]({{< ref "../configuration/events.md" >}}) |
| `time` | Event timestamp as a JavaScript `Date` |
| `is_live` | Whether the event was received while the stream was live |
| `reason` | Whether this is a live event, bulk replay, or single-event replay |

### Client options

| Option | Description |
| --- | --- |
| `token` | Required token with the `overlays` permission. Prefer passing it through the URL hash. |
| `handlers` | Object mapping event types to handler functions. Use `_` as the handler for all events. |
| `channel` | Only dispatch events whose `fields.channel` matches this value. |
| `replay` | Fetch stored events for `channel` when the client connects. Defaults to `false`. |
| `maxReplayAge` | Maximum age of replayed events in hours. By default all stored events are fetched. |

Options may also be supplied through the URL hash. Hash parameters take precedence over constructor options.

## Helper methods

### `paramOptionFallback(key, fallback)`

Reads a value from the URL hash, then from the constructor options, and finally returns the supplied fallback. This can also be used for overlay-specific settings:

```javascript
const duration = Number(client.paramOptionFallback('duration', 10))
```

With `#duration=30` in the overlay URL, `paramOptionFallback` returns `"30"` and `Number` converts it to a number. URL parameters are strings, so convert numbers and booleans as needed.

### `renderTemplate(template)`

Renders a template through the bot's `msgformat` API and returns the rendered text:

```javascript
await client.renderTemplate('{{ recentTitle "mychannel" }}')
```

The token requires the `msgformat` permission in addition to `overlays`.

### `replayEvent(eventId)`

Replays one stored event to all connected overlays and returns the HTTP `Response` for the request:

```javascript
const response = await client.replayEvent(event.event_id)
```

### `apiBase()`

Returns the bot API base URL derived from the current overlay URL, without a trailing slash.
