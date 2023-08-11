---
title: Overlays
---

{{< lead >}}
Overlays in OBS are added as a **Browser Source** and while often graphical tooling to build them is available it's possible to build custom overlays with well-known web-technology like **HTML, Javascript, CSS**.
{{< /lead >}}

{{< alert style="info" >}}
In the [Service Setup]({{< ref "../getting-started/setup.md" >}}) we configured an `OVERLAYS_DIR` which is used to serve the files for the overlays. Every file you put into that directory is available to the public at `https://your-bot.example.com/overlays/`. Therefore pay attention not to put any secrets into that directory!
{{< /alert >}}

The bot includes some files which are merged with the files you put into that directory. If you put a file with the same name as the included, your file will overwrite the included one! Therefore you can modify included templates by just copying them into your overlay directory and changing the stuff you want changed.

Currently the following files are available in the default distribution:

- `debug.html` - The Debug Overlay (see below for an example how to use)
- `eventclient.js` - The [EventClient]({{< ref "eventclient.md" >}}) Javascript library to aid you in developing overlays and communicating with the bot
- `sounds.html` / `sounds.js` - The [Sound-Alerts Overlay]({{< ref "soundalerts.md" >}})
- `template.html` - A very simple example overlay without external dependencies

You can see the sources for these included files in the [project repository](https://github.com/Luzifer/twitch-bot/tree/master/internal/apimodules/overlays/default).

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
