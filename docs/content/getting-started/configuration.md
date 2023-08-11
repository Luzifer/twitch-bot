+++
title = "Configuration"
weight = 5
+++

{{< lead >}}
After you finally can access your bot through the [External Access]({{< ref "external-access.md" >}}) you can start to configure it!
{{< /lead >}}

In order to gain access to the bot you need to add yourself as an editor to the configuration. To do so edit the configuration created during the [Service Setup]({{< ref "setup.md" >}}) (in our example `/var/lib/twitch-bot/config.yaml`) and modify the bot editors:

```yaml
# List of strings: Either Twitch user-ids or nicknames (best to stick
# with IDs as they can't change while nicknames can be changed every
# 60 days). Those users are able to use the config editor web-interface.
bot_editors: [ 'luziferus' ]
```

Of course you should not enter my nickname but yours into the list. Don't worry about the hint IDs should be used. Using the nickname is fine for now and the bot will automatically adjust the entry when you start configuring the bot through the web-interface.

The confiuration change is automatically loaded by the bot (you should see that in the service log) and you should now be able to log into the web-interface using the **Login with Twitch** button.

At this point you've fully set up the bot to run and you've gained access to it and with this state the "Getting Started" guide ends. You can now

- add a channel in the web-interface, configure rules with [actions]({{< ref "../configuration/actors.md" >}}) and [templates]({{< ref "../configuration/templating.md" >}})
- have a look on all the other options in the [Config-File Syntax]({{< ref "../configuration/config-file.md" >}})
- discover what the bot can do in the [Rule Examples]({{< ref "../configuration/rule-examples.md" >}})
