---
title: Spotify Integration
---

> [!TIP]
> You are using Spotify and are tired of not working third-party overlays and chat commands? The bot has you covered with its Spotify integration. The integration can retrieve the current playing track and show that in templating as for example through the [EventClient]({{< ref "../overlays/eventclient.md" >}}) and its `renderTemplate` function or the `respond` actor in a rule.

## Setting up

You will need

- a Spotify account
- an instance of the bot with access to the configuration

For this documentation we assume your bots web-interface is available at `https://example.com/` and everywhere you see that below, you need to replace it with your own instance URL.

Start with going to the [Spotify for Developers Dashboard](https://developer.spotify.com/dashboard) and create a new app:

- "App name" is something you can choose yourself
- "App description" is also required, choose yourself
- "Redirect URI" must be `https://example.com/spotify/<channel>` so for example `https://example.com/spotify/luziferus`
- Select "Web API" for the "API/SDKs you are planning to use"
- Check the ToS box (of course after reading those!)
- Click "Save"
- From the "Settings" button of your app get the "Client ID" and note it down
- Optional: If you need to authorize multiple channels (i.e. for multiple users of the bot instance) you can edit the "Redirect URIs" on the "Settings" page and add more.

> [!INFO]
> If you are managing a bot instance for multiple persons having their own Spotify accounts you need to invite them to the Spotify app as long as it is in development-mode. You can do that in the Spotify Developer Dashboard under "User Management" (up to 25 users). As an alternative every person can create an own Spotify app and you can enter their `clientId` into the config for their respective channel.

Now head into the configuration file and configure the Spotify module:

```yaml
# Module configuration by channel or defining bot-wide defaults. See
# module specific documentation for options to configure in this
# section.
module_config:
  spotify:
    # Use one client-id for all channels (invite users)
    default:
      clientId: 'put the client ID you noted down here'
    # Use one client-id per channel (have each user create an app)
    anotherttvuser:
      clientId: 'put the client ID they sent you here'
```

Now send the user which currently playing track should be displayed to the `https://example.com/spotify/<channel>` URL. So I for example would visit `https://example.com/spotify/luziferus`. They are redirected to Spotify, need to authorize the app and if everything went well the bot tells them "Spotify is now authorized for this channel, you can close this page".

Now you can for example add a new rule for the channel:
```yaml
- uuid: 0cd18de8-d70b-4651-a51a-3de1a2eb87c5
  description: Spotify
  actions:
    - type: respond
      attributes:
        message: '{{ spotifyCurrentPlaying .channel }}'
  match_message: '!spotify'
```
