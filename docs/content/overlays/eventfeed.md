---
title: Event-Feed
weight: 50
---

> [!TIP]
> The Event-Feed overlay can be opened in a normal browser or added to OBS as a Custom Browser Dock. It shows recent supported events and a per-stream summary for bits, donations, follows, raids, and subs.

Use this URL format:

```text
/overlays/eventfeed.html#token=[your-token]&channel=[your-channel]&replay=true&maxReplayAge=168
```

Parameters are passed through the URL hash:

- `token` - Token configured in the bot with access to the `overlays` permission
- `channel` - Channel filter including the leading `#`; in URLs the `#` is encoded as `%23`
- `replay` - Set to `true` to load past events when opening the feed
- `maxReplayAge` - Number of hours to load when replay is enabled; default is `168`, `-1` loads all stored events

The feed has a filter menu to hide event categories. Filter settings and the `Mark read` timestamp are stored per channel in your browser local storage. `Mark read` dims older events so newer entries stand out. Some events also show a replay button, which re-sends that event to currently connected overlays.

### Customizing the Event-Feed

To customize the Event-Feed, copy `eventfeed.custom.js` into your overlay directory and edit it there. Your file will override the bundled default file.

The custom file must export these functions:

- `customFilters()` - Returns additional filter definitions as `{ filterKey: { name: "Name", visible: true } }`
- `customHandler(eventObj)` - Receives custom events and returns an event object for display or `null` to ignore it

Returned event objects must contain `eventId`, `filterKey`, `time`, and `title`. Useful optional fields are `icon`, `text`, `subtext`, `extraData`, and `hasReplay`.

Example:

```js
function customFilters() {
  return {
    timer: { name: 'Timers', visible: true },
  }
}

function customHandler({ event_id, fields, time }) {
  if (fields.type !== 'timer') {
    return null
  }

  return {
    eventId: event_id,
    filterKey: 'timer',
    icon: 'fas fa-stopwatch',
    text: fields.name ? `Timer-Name: ${fields.name}` : null,
    time: new Date(new Date(time).getTime() + fields.time * 1000),
    title: `Timer for ${fields.displayTime || `${fields.time}s`} started`,
  }
}

export { customFilters, customHandler }
```
