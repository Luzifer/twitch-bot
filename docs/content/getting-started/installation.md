---
title: "Installation"
weight: 1
---

{{< lead >}}
Installation is the first step you need to take, therefore choose how you want to proceed:
{{< /lead >}}

## Downloading a pre-compiled Binary

You can always find the latest pre-compiled binary on the [Releases-Page](https://github.com/Luzifer/twitch-bot/releases) of the Github repository. The latest release contains binaries for MacOS (`darwin`), Linux (`amd64` / `arm`) and Windows.

This way of installation is the easiest one:

- Download the archive for your system
- Unzip / untar the archive
- Find a binary in the location you unpacked the archive to

The binary you get from this is everything you need: It contains the bot as well as the web-interface to configure the bot.

Move this binary to a location you will find it again later.

## Using a pre-built Docker image

The Docker image is automatically built from the source and provided in two different variants:

| Image | Description |
| ----- | ----------- |
| <span style="white-space:nowrap">`luzifer/twitch-bot:latest`</span> | The latest development version, not recommended for use as your main bot, perfect for testing the latest changes not yet released. |
| <span style="white-space:nowrap">`luzifer/twitch-bot:stable`</span> | Automatically updated on every versioned release, you just can use this tag to always have the latest stable version. Pay attention: This automatically switches over on major / breaking releases! |
| <span style="white-space:nowrap">`luzifer/twitch-bot:<version>`</span> | If you don't want to auto-update you can use the tagged version in the Docker image (i.e. tag `v3.15.0` would be available as `luzifer/twitch-bot:v3.15.0`) |

## Building the Binary yourself

If you want to do customizations or just don't trust my build system you can execute the whole build yourself which requires you to have the [latest Golang release](https://go.dev/dl/) and [LTS NodeJS release](https://nodejs.org/en) available on your machine.

```console
# First checkout the code
$ git clone https://github.com/Luzifer/twitch-bot.git
$ cd twitch-bot

# Optionally switch to a release tag
$ git checkout v3.15.0

# Second build the binary
$ make build_prod
```

Move this binary to a location you will find it again later.
