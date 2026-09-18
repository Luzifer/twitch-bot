---
title: EventClient
weight: 10000
---

## Typed overlay development

Import `eventclient.js` directly. TypeScript-aware editors automatically use the adjacent `eventclient.d.ts` and `eventTypes.d.ts` declarations for event-specific handler autocomplete and type checking.

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
