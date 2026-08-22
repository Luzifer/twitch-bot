---
title: Install on Uberspace U8
weight: 2
---

> [!TIP]
> [Uberspace](https://uberspace.de/) is a German hoster providing you with a platform suitable to host your own Twitch-Bot on. You will get access to a [shell](https://u8manual.uberspace.de/basics_shell/) and your bot will run on real services without you having to care for managing the whole server yourself. You only need to install, configure and update your bot.

> [!WARNING]
> This document only applies to the new Uberspace U8 (currently in beta), not to the stable U7. U8 is usable for hosting the bot, but as the beta label indicates, it is not finished and not yet fit for general use.

> [!IMPORTANT]
> If you're completely new to working with a shell ("the terminal"), feel free to roam through the [Uberspace Manual](https://u8manual.uberspace.de/) and familiarize yourself with what you can do with your Uberspace account. Also please don't simply copy commands from anywhere on the internet into your Uberspace shell without understanding them first: Use Google or ask your favorite AI to explain the commands!
> 
> When using the `nano` editor, press Ctrl+X to exit. It will ask whether you want to save. Press `Y` to confirm, then confirm the filename with `Enter`.

This whole document assumes you want to install only one bot in your account and want to use `twitch-bot.<your-username>.uber.space` to access the bot. To make it a little easier to read the document will use the `twitch-bot.luzifer.uber.space` domain as an example. Therefore it assumes you registered these three redirects in the [Preparations](./preparations.md) step:

- `https://twitch-bot.luzifer.uber.space/`
- `https://twitch-bot.luzifer.uber.space/auth/update-bot-token`
- `https://twitch-bot.luzifer.uber.space/auth/update-channel-scopes`

## Preparing some Directories

As we need to store the bot itself, its configuration and the overlay directory we need some preparation. Therefore we'll first create some directories:

```console
$ install -dm0700 ~/.config/twitch-bot ~/.local/share/twitch-bot/overlays
```

## Downloading the latest Twitch-Bot binary

To run the bot, we first need to download it into your Uberspace:

```console
$ curl -sSfL https://github.com/Luzifer/twitch-bot/releases/latest/download/twitch-bot_linux_amd64.tgz |
    tar -xvz -C ~/.local/bin twitch-bot
```

Afterwards you should be able to ask it for its version:

```console
$ twitch-bot --version
twitch-bot v3.42.2
```

## Setting up the bot as a Service

In order to have the bot start automatically we need to add a little configuration and a service.

First we need to set some environment variables for the bot:

```console
$ install -Dm0600 /dev/null ~/.config/twitch-bot/environment
$ nano ~/.config/twitch-bot/environment
```

Let's put this into the file we've just opened (replace the brackets, don't leave them inside the file!):

```env
BASE_URL=https://twitch-bot.[username].uber.space/

CONFIG=/home/[username]/.config/twitch-bot/config.yaml
LOG_LEVEL=info
OVERLAYS_DIR=/home/[username]/.local/share/twitch-bot/overlays

STORAGE_CONN_STRING=/home/[username]/.local/share/twitch-bot/storage.db
STORAGE_CONN_TYPE=sqlite
STORAGE_ENCRYPTION_PASS=[put a random secure password here]

TWITCH_CLIENT=[put the client-id from the preparations step here]
TWITCH_CLIENT_SECRET=[put the client-secret from the preparations step here]
```

> [!NOTE]
> Of course you can replace the SQLite database with MariaDB or PostgreSQL but for that please consult the Uberspace manual and the project's [README](https://github.com/Luzifer/twitch-bot/blob/master/README.md#database-connection-strings). For this document we stick with SQLite.

Now that we have an environment for the bot and have told it where to find everything it needs, we can create a [Service](https://u8manual.uberspace.de/services_systemd/) for it:

```console
$ nano ~/.config/systemd/user/twitch-bot.service
```

Inside that file just copy this:

```service
[Unit]
Description=Twitch-Bot Service

[Service]
EnvironmentFile=%h/.config/twitch-bot/environment
ExecStart=%h/.local/bin/twitch-bot
Restart=Always
RestartSec=5

[Install]
WantedBy=default.target
```

After you created that file, you need to reload the configuration and then enable and start the bot:

```console
$ systemctl --user daemon-reload
$ systemctl --user enable --now twitch-bot
$ systemctl --user status twitch-bot
```

This will start the bot in the background and it will create the configuration file in `~/.config/twitch-bot/config.yaml`. For security reasons that file contains two settings we need to change:

```console
$ nano ~/.config/twitch-bot/config.yaml
```

Look for the line `http_listen: "127.0.0.1:0"` and change it into this:

```yaml
http_listen: "0.0.0.0:3000"
```

Also add your Twitch username to the `bot_editors` list (you don't need to do that again when doing the [Configuration](./configuration.md) later):

```yaml
bot_editors: [ 'your-twitch-username' ]
```

Now we need to restart the bot to have it use the new configuration:

```console
$ systemctl --user restart twitch-bot
```

## Allowing access from the internet

Now that the bot is running we just need to tell Uberspace that we want to access it from outside the shell:

```console
$ uberspace web domain add twitch-bot.$USER.uber.space
$ uberspace web backend add twitch-bot.$USER.uber.space/ port 3000
```

Those commands might take a moment as Uberspace needs to do some stuff in the background so everything works.

When you've executed the second command, your bot will be available under `https://twitch-bot.[username].uber.space`! Now you can log in and follow the rest of the [Configuration](./configuration.md).

In case the bot is not responding, have a look at the logs (see below) and check whether you followed all commands.

When you're set up and want to have a look at [Overlays](../overlays/_index.md) you need the directory you created at the beginning: `~/.local/share/twitch-bot/overlays` - that's where your overlays live.

## Updating the bot

Of course you want and need to update the Bot: there might be cool new features and also there might be security updates. Luckily this is also relatively easy:

First we need to stop the bot as it must not be running during the update:

```console
$ systemctl --user stop twitch-bot
```

Then we execute the same command we used for downloading the latest version of the bot which will overwrite the current version:

```console
$ curl -sSfL https://github.com/Luzifer/twitch-bot/releases/latest/download/twitch-bot_linux_amd64.tgz |
    tar -xvz -C ~/.local/bin twitch-bot
```

Afterwards we simply start the bot:

```console
$ systemctl --user start twitch-bot
```

## Looking at the logs

If something doesn't work as expected you might want to take a look at the logs:

```console
$ journalctl --user -fu twitch-bot
```

Press Ctrl+C to stop the log output.
