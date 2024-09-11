---
title: EventClient
weight: 10000
---

## Classes

<dl>
<dt><a href="#EventClient">EventClient</a></dt>
<dd><p>EventClient abstracts the connection to the bot websocket for events</p>
</dd>
</dl>

## Typedefs

<dl>
<dt><a href="#Options">Options</a> : <code>Object</code></dt>
<dd><p>Options to pass to the EventClient constructor</p>
</dd>
<dt><a href="#SocketMessage">SocketMessage</a> : <code>Object</code></dt>
<dd><p>SocketMessage received for every event and passed to the new <code>(eventObj) =&gt; { ... }</code> handlers</p>
</dd>
</dl>

<a name="EventClient"></a>

## EventClient
EventClient abstracts the connection to the bot websocket for events

**Kind**: global class  

* [EventClient](#EventClient)
    * [new EventClient(opts)](#new_EventClient_new)
    * [.apiBase()](#EventClient+apiBase) ⇒ <code>string</code>
    * [.paramOptionFallback(key, [fallback])](#EventClient+paramOptionFallback) ⇒ <code>\*</code>
    * [.renderTemplate(template)](#EventClient+renderTemplate) ⇒ <code>Promise</code>
    * [.replayEvent(eventId)](#EventClient+replayEvent) ⇒ <code>Promise</code>

<a name="new_EventClient_new"></a>

### new EventClient(opts)
Creates, initializes and connects the EventClient


| Param | Type | Description |
| --- | --- | --- |
| opts | [<code>Options</code>](#Options) | Options for the EventClient |

<a name="EventClient+apiBase"></a>

### eventClient.apiBase() ⇒ <code>string</code>
Returns the API base URL without trailing slash

**Kind**: instance method of [<code>EventClient</code>](#EventClient)  
**Returns**: <code>string</code> - API base URL  
<a name="EventClient+paramOptionFallback"></a>

### eventClient.paramOptionFallback(key, [fallback]) ⇒ <code>\*</code>
Resolves the given key through url hash parameters with fallback to constructor options

**Kind**: instance method of [<code>EventClient</code>](#EventClient)  
**Returns**: <code>\*</code> - Value of the key or `null`  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| key | <code>string</code> |  | The key to resolve |
| [fallback] | <code>\*</code> | <code></code> | Fallback to return if neither params nor options contained that key |

<a name="EventClient+renderTemplate"></a>

### eventClient.renderTemplate(template) ⇒ <code>Promise</code>
Renders a given template using the bots msgformat API (supports all templating you can use in bot messages). To use this function the token passed through the constructor or the URL hash must have the `msgformat` permission in addition to the `overlays` permission.

**Kind**: instance method of [<code>EventClient</code>](#EventClient)  
**Returns**: <code>Promise</code> - Promise resolving to the rendered output of the template  

| Param | Type | Description |
| --- | --- | --- |
| template | <code>string</code> | The template to render |

<a name="EventClient+replayEvent"></a>

### eventClient.replayEvent(eventId) ⇒ <code>Promise</code>
Triggers a replay of the given event to all overlays currently listening for events. This event will have the `is_live` flag set to `false`.

**Kind**: instance method of [<code>EventClient</code>](#EventClient)  
**Returns**: <code>Promise</code> - Promise of the fetch request  

| Param | Type | Description |
| --- | --- | --- |
| eventId | <code>Number</code> | The ID of the event received through the SocketMessage object |

<a name="Options"></a>

## Options : <code>Object</code>
Options to pass to the EventClient constructor

**Kind**: global typedef  
**Properties**

| Name | Type | Default | Description |
| --- | --- | --- | --- |
| [channel] | <code>String</code> |  | Filter for specific channel events (format: `#channel`) |
| [handlers] | <code>Object</code> | <code>{}</code> | Map event types to callback functions `(eventObj) => { ... }` (new) or `(event, fields, time, live) => {...}` (old) |
| [maxReplayAge] | <code>Number</code> | <code>-1</code> | Number of hours to replay the events for (-1 = infinite) |
| [replay] | <code>Boolean</code> | <code>false</code> | Request a replay at connect (requires channel to be set to a channel name) |
| [token] | <code>String</code> |  | API access token to use to connect to the WebSocket (if not set, must be provided through URL hash) |

<a name="SocketMessage"></a>

## SocketMessage : <code>Object</code>
SocketMessage received for every event and passed to the new `(eventObj) => { ... }` handlers

**Kind**: global typedef  
**Properties**

| Name | Type | Description |
| --- | --- | --- |
| [event_id] | <code>String</code> | UID of the event used to re-trigger an event |
| [is_live] | <code>Boolean</code> | Whether the event was sent through a replay (false) or occurred live (true) |
| [reason] | <code>String</code> | Reason of this message (one of `bulk-replay`, `live-event`, `single-replay`) |
| [time] | <code>String</code> | RFC3339 timestamp of the event |
| [type] | <code>String</code> | Event type (i.e. `raid`, `sub`, ...) |
| [fields] | <code>Object</code> | string->any mapping of fields available for the event |

