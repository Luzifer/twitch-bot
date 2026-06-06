---
title: Overlays
---

> [!TIP]
> Overlays in OBS are added as a **Browser Source** and while often graphical tooling to build them is available it's possible to build custom overlays with well-known web-technology like **HTML, Javascript, CSS**.

> [!INFO]
> In the Service Setup we configured an `OVERLAYS_DIR` which is used to serve the files for the overlays. Every file you put into that directory is available to the public at `https://your-bot.example.com/overlays/`. Therefore pay attention not to put any secrets into that directory!

The bot includes some files which are merged with the files you put into that directory. If you put a file with the same name as the included, your file will overwrite the included one! Therefore you can modify included templates by just copying them into your overlay directory and changing the stuff you want changed.

Currently the following files are available in the default distribution:

- `debug.html` - The Debug Overlay (see below for an example how to use)
- `eventclient.js` - The [EventClient]({{< ref "eventclient.md" >}}) Javascript library to aid you in developing overlays and communicating with the bot
- `eventfeed.html` / `eventfeed.js` / `eventfeed.custom.js` - The Event-Feed Overlay and a customization file to adapt event rendering
- `sounds.html` / `sounds.js` - The [Sound-Alerts Overlay]({{< ref "soundalerts.md" >}})
- `template.html` - A very simple example overlay without external dependencies

You can see the sources for these included files in the [project repository](https://github.com/Luzifer/twitch-bot/tree/master/internal/apimodules/overlays/default).

## Event-Feed

The Event-Feed overlay can be opened in a normal browser or added to OBS as a Browser Source / Custom Browser Dock. It shows recent supported events and a per-stream summary for bits, donations, follows, raids, and subs.

Use this URL format:

```text
https://your-bot.example.com/overlays/eventfeed.html#token=[your-token]&channel=[your-channel]&replay=true&maxReplayAge=168
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

## Configuring Secrets and Parameters

As you shouldn't put secrets in a public place the [EventClient]({{< ref "eventclient.md" >}}) library provides a mechanism to fetch secrets from the URL hash through the `paramOptionFallback` method. This also is used for all configuration you can pass into the class options. Especially for the token you should use this mechaism.

As an example lets have a look at this URL path:

`/overlays/debug.html#token=55cdb1e4-c776-4467-8560-a47a4abc55de&replay=true&channel=%23luziferus&maxReplayAge=720&hide=join,part`

Here you can see the Debug Overlay configured with:

- `token` - A token configured in the bot having access to the `overlays` permission
- `replay` - Enabled to retrieve past events
- `channel` - Set to `#luziferus` to retrieve events from my channel
- `maxReplayAge` - Set to 720h (30d) not to retrieve events from the infinite past
- `hide` - Set to omit `join` and `part` events (custom parameter for the Debug Overlay)

As those parameters are configured through the URL hash (`#...`) they are never sent to the server, therefore are not logged in any access-logs and exist only in the local URL. So with a custom overlay you would put `https://your-bot.example.com/overlays/myoverlay.html#token=55cdb1e4-c776-4467-8560-a47a4abc55de` into your OBS browser source and your overlay would be able to communicate with the bot.

The debug-overlay can be used to view all events received within the bot you can react on in overlays and bot rules.

## Remote editing Overlays with local Editor

In order to enable you to edit the overlays remotely when hosting the bot on a server the bot exposes a WebDAV interface you can locally mount and work on using your favorite editor. To mount the WebDAV I recommend [rclone](https://rclone.org/). You will need the URL your bot is available at and a token with `overlays` permission:

```
# rclone obscure 55cdb1e4-c776-4467-8560-a47a4abc55de
MqO0FLdbg3txom2IpUMsVVIqnHwYDefms4EKRqoV1MGhCFkBmWnhvVRdqTyCSFtmvP-AYg

# cat /tmp/rclone.conf
[bot]
type = webdav
url = https://your-bot.example.com/overlays/dav/
user = dav
pass = MqO0FLdbg3txom2IpUMsVVIqnHwYDefms4EKRqoV1MGhCFkBmWnhvVRdqTyCSFtmvP-AYg

# rclone --config /tmp/rclone.conf mount bot:/ /tmp/bot-overlays

# code /tmp/bot-overlays
```

What I've done here is to obscure the token (`rclone` wants the token to be in an obscured format), create a config containing the WebDAV remote, mount the WebDAV remote to a local directory and open it with VSCode to edit the overlays. When saving the files locally `rclone` will upload them to the bot and refreshing the overlay in your browser / OBS will give you the new version.
