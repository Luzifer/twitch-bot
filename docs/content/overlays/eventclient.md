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

<a name="Options"></a>

## Options : <code>Object</code>
Options to pass to the EventClient constructor

**Kind**: global typedef  
**Properties**

| Name | Type | Default | Description |
| --- | --- | --- | --- |
| [channel] | <code>string</code> |  | Filter for specific channel events (format: `#channel`) |
| [handlers] | <code>Object</code> | <code>{}</code> | Map event types to callback functions `(event, fields, time, live) => {...}` |
| [maxReplayAge] | <code>number</code> | <code>-1</code> | Number of hours to replay the events for (-1 = infinite) |
| [replay] | <code>boolean</code> | <code>false</code> | Request a replay at connect (requires channel to be set to a channel name) |
| [token] | <code>string</code> |  | API access token to use to connect to the WebSocket (if not set, must be provided through URL hash) |

