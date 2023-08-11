---
title: "Config-File Syntax"
---

{{< lead >}}
The YAML configuration file is the heart of the bot configuration. You can configure every aspect of the bot using the configuration file. The web-interface afterwards allows to modify the configuration file to assist you with the configuration.
{{< /lead >}}

```yaml
---

# This must be the config version you've used below. Current version
# is version 2 so probably keep it at 2 until the bot tells you to
# upgrade.
config_version: 2

# List of tokens allowed to access the HTTP API with write access.
# You can generate a token using the web-based config-editor or the
# `api-token` sub-command:
#     $ twitch-bot api-token 'mytoken' '*'
# The token will only be printed ONCE and cannot be retrieved afterards.
auth_tokens:
  89196495-68eb-4f50-94f0-5c5d99f26be5:
    hash: '243261[...]36532e'
    modules:
    - '*'
    name: mytoken

# List of strings: Either Twitch user-ids or nicknames (best to stick
# with IDs as they can't change while nicknames can be changed every
# 60 days). Those users are able to use the config editor web-interface.
bot_editors: []

# List of channels to join. Channels not listed here will not be
# joined and therefore cannot be actioned on.
channels:
  - mychannel

# The bot is able to track config changes made through the config-editor
# web-interface using Git (https://git-scm.com/). To use this feature
# create a Git repository in the folder the config is placed: `git init`
#
# Afterwards switch this option to `true` and you're good to go: Each
# change made by the editor causes a Git commit with the logged in user
# as author of the commit.
git_track_config: false

# Enable HTTP server to control plugins / core functionality
# if unset the server is not started, to change the bot must be restarted
http_listen: "127.0.0.1:3000"

# Allow moderators to hand out permits (if set to false only broadcaster can do this)
permit_allow_moderator: true
# How long to permit on !permit command
permit_timeout: 60s

# Variables are made available in templating (for example useful to disable several
# rules at once using the `disable_on_template` directive)
# Supported data types: Boolean, Float, Integer, String
variables:
  myvariable: true
  anothervariable: "string"

# List of auto-messages. See documentation for details or use
# web-interface to configure.
auto_messages:
  - channel: 'mychannel'          # String, channel to send message to
    message: 'Automated message'  # String, message to send
    use_action: true              # Bool, optional, send message as action (`/me <message>`)

    # Even though all of these are optional, at least one MUST be specified for the entry to be valid
    cron: '*/10 * * * *'  # String, optional, cron syntax when to send the message
    message_interval: 3   # Integer, optional, how many non-bot-messages must be sent in between

    only_on_live: true    # Boolean, optional, only send the message when channel is live

    # Disable message using templating, must yield string `true` to disable the automated message
    disable_on_template: '{{ ne .myvariable true }}'

# List of rules. See documentation for details or use web-interface
# to configure.
rules: # See below for examples

  - actions: # Array of actions to take when this rule matches

    # See the Actors page in the Wiki for available actors:
    # https://github.com/Luzifer/twitch-bot/wiki/Actors
    - type: "<actor type>"
      attributes:
        key: value

    # Add a cooldown to the rule in general (not to trigger counters twice, ...)
    # Using this will prevent the rule to be executed in all matching channels
    # as long as the cooldown is active.
    cooldown: 1s # Duration value: 1s / 1m / 1h

    # Add a cooldown to the rule per channel (not to trigger counters twice, ...)
    # Using this will prevent the rule to be executed in the channel it was triggered
    # which means other channels are not affected.
    channel_cooldown: 1s # Duration value: 1s / 1m / 1h

    # Add a cooldown to the rule per user (not to trigger counters twice, ...)
    # Using this will prevent the rule to be executed for the user which triggered it
    # in any of the matching channels, which means other users can trigger the command
    # while that particular user cannot
    user_cooldown: 1s # Duration value: 1s / 1m / 1h

    # Do not apply cooldown for these badges
    skip_cooldown_for: [broadcaster, moderator]

    # Disable the rule by setting to true
    disable: false

    # Disable actions when the matched channel has no active stream
    disable_on_offline: false

    # Disable actions on this rule if the user has an active permit
    disable_on_permit: false

    # Disable actions using templating, must yield string `true` to disable the rule
    disable_on_template: '{{ ne .myvariable true }}'

    # Disable actions on this rule if the user has one of these badges
    disable_on: [broadcaster, moderator]

    # Enable actions on this rule only if the user has one of these badges
    enable_on: [broadcaster, moderator]

    # Require the chat message to be sent in this channel
    match_channels: ['#mychannel']

    # Require the chat message to be sent by one of these users
    match_users: ['mychannel'] # List of users, all names MUST be all lower-case

    # Execute actions when this event occurs
    # See the Events page in the Wiki for available events and field documentation
    # https://github.com/Luzifer/twitch-bot/wiki/Events
    match_event: 'permit'

    # Execute action when the chat message matches this regular expression
    match_message: '' # String, regular expression

    # Disable the actions on this rule if one of these regular expression matches the chat message
    disable_on_match_messages: []

...
```
