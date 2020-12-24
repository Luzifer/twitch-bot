[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/twitch-bot)](https://goreportcard.com/report/github.com/Luzifer/twitch-bot)
![](https://badges.fyi/github/license/Luzifer/twitch-bot)
![](https://badges.fyi/github/downloads/Luzifer/twitch-bot)
![](https://badges.fyi/github/latest-release/Luzifer/twitch-bot)
![](https://knut.in/project-status/twitch-bot)

# Luzifer / twitch-bot

Twitch-Bot is intended as an alternative to having a bot managed by Streamlabs or Streamelements and therefore having more control over it, the availability and how it works.

At the moment it is a work-in-progress and does not yet implment all features it shall in the future.

## Configuration

Please see the [Wiki](https://github.com/Luzifer/twitch-bot/wiki) for documentation of the configuration file.

```console
# twitch-bot --help
Usage of twitch-bot:
  -c, --config string          Location of configuration file (default "./config.yaml")
      --log-level string       Log level (debug, info, warn, error, fatal) (default "info")
      --storage-file string    Where to store the data (default "./storage.json.gz")
      --twitch-client string   Client ID to act as
      --twitch-token string    OAuth token valid for client
      --version                Prints current version and exits
```
