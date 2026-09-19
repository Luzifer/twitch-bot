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

<a name="EventClient"></a>

## EventClient
**Kind**: global class

* [EventClient](#EventClient)
    * [new EventClient(opts)](#new_EventClient_new)
    * [.apiBase()](#EventClient+apiBase) ⇒
    * [.paramOptionFallback(key, fallback)](#EventClient+paramOptionFallback) ⇒
    * [.renderTemplate(template)](#EventClient+renderTemplate) ⇒
    * [.replayEvent(eventId)](#EventClient+replayEvent) ⇒

<a name="new_EventClient_new"></a>

### new EventClient(opts)
Creates, initializes and connects the EventClient.


| Param | Description |
| --- | --- |
| opts | EventClient options |

<a name="EventClient+apiBase"></a>

### eventClient.apiBase() ⇒
Returns the API base URL without trailing slash.

**Kind**: instance method of [<code>EventClient</code>](#EventClient)
**Returns**: API base URL
<a name="EventClient+paramOptionFallback"></a>

### eventClient.paramOptionFallback(key, fallback) ⇒
Resolves a URL hash parameter with a fallback to the constructor options.

**Kind**: instance method of [<code>EventClient</code>](#EventClient)
**Returns**: Resolved option value

| Param | Default | Description |
| --- | --- | --- |
| key |  | Option key to resolve |
| fallback | <code></code> | Value returned when the option is absent |

<a name="EventClient+renderTemplate"></a>

### eventClient.renderTemplate(template) ⇒
Renders a template using the bot's msgformat API.

The token requires the `msgformat` permission in addition to `overlays`.

**Kind**: instance method of [<code>EventClient</code>](#EventClient)
**Returns**: Rendered template

| Param | Description |
| --- | --- |
| template | Template to render |

<a name="EventClient+replayEvent"></a>

### eventClient.replayEvent(eventId) ⇒
Triggers a replay of an event for all connected overlays.

**Kind**: instance method of [<code>EventClient</code>](#EventClient)
**Returns**: Fetch response

| Param | Description |
| --- | --- |
| eventId | Event ID received in a socket message |
