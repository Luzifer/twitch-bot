---
title: Ko-fi Event-Integration
---

> [!TIP]
> If you are an active user of Ko-fi you probably want to have the Ko-fi events sent to your bot to trigger chat-messages, alerts in overlays or just to have the event registered in the bot for overlays or other purposes. To do so you can use the Ko-fi integration of the bot to receive events for donations and subscriptions (shop-orders currently are not supported).

## Setting up

You will need

- a Ko-fi account
- an instance of the bot with access to the configuration
- the verification token available in the "API" menu entry in your settings page on Ko-fi

Given the bot web-interface is available at `https://example.com/` your webhook URL would be `https://example.com/kofi/webhook/<your channel>` so for example `https://example.com/kofi/webhook/luziferus`. You will later enter this into the "Webhook URL" field in the "API" settings-panel.

At first copy your "Verification Token" from the "API" settings-panel (and don't tell anyone!). You need to create a new block within your bots configuration file (at the moment there is no way to configure this through the web-interface):

```yaml
# Module configuration by channel or defining bot-wide defaults. See
# module specific documentation for options to configure in this
# section. All modules come with internal defaults so there is no
# need to configure this but you can overwrite the internal defaults.
module_config:
  kofi:
    luziferus:  # put your channel name here, is the same as in the URL
      verification_token: 'your verification token'
```

You can configure one token per channel and **should not** use the `default` entry as that would be used for all channels. This for example applies if your bot instance manages multiple channels with different Ko-fi accounts attached to them.

Now that we know about the verification token, you can put the URL into the "API" settings-panel and click the "Update" button.

As soon as you now click the "Send Single Donation Test", "Send First Monthly Test" or "Send Membership Tier Test" buttons you should see a `kofi_donation` event in the [debug overlay]({{< ref "../overlays/_index.md" >}}).

From here you can create rules using the [`kofi_donation` event]({{< ref "../configuration/events.md" >}}#kofi_donation) doing stuff.
