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
      --command-timeout duration   Timeout for command execution (default 30s)
  -c, --config string              Location of configuration file (default "./config.yaml")
      --log-level string           Log level (debug, info, warn, error, fatal) (default "info")
      --plugin-dir string          Where to find and load plugins (default "/usr/lib/twitch-bot")
      --rate-limit duration        How often to send a message (default: 20/30s=1500ms, if your bot is mod everywhere: 100/30s=300ms, different for known/verified bots) (default 1.5s)
      --storage-file string        Where to store the data (default "./storage.json.gz")
      --twitch-client string       Client ID to act as
      --twitch-token string        OAuth token valid for client
  -v, --validate-config            Loads the config, logs any errors and quits with status 0 on success
      --version                    Prints current version and exits

# twitch-bot help
Supported sub-commands are:
  actor-docs                     Generate markdown documentation for available actors
  api-token <name> <scope...>    Generate an api-token to be entered into the config
  help                           Prints this help message
```
