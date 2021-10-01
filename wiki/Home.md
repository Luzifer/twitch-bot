## Configuration

```yaml
---

# This must be the config version you've used below. Current version
# is version 2 so probably keep it at 2 until the bot tells you to
# upgrade.
config_version: 2

# List of strings: Either Twitch user-ids or nicknames (best to stick
# with IDs as they can't change while nicknames can be changed every
# 60 days). Those users are able to use the config editor web-interface.
bot_editors: []

# List of channels to join. Channels not listed here will not be
# joined and therefore cannot be actioned on.
channels:
  - mychannel

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
    # Available events: category_update, join, host, part, permit, raid, resub,
    #                   stream_offline, stream_online, sub, subgift, title_update, whisper
    match_event: 'permit'

    # Execute action when the chat message matches this regular expression
    match_message: '' # String, regular expression

    # Disable the actions on this rule if one of these regular expression matches the chat message
    disable_on_match_messages: []

...
```

## Templating

There are certain variables available in the strings with templating enabled:

- `channel` - Channel the message was sent to, only available for regular messages not events
- `msg` - The message object, used in functions, should not be sent to chat
- `permitTimeout` - Value of `permit_timeout` in seconds
- `username` - The username of the message author

Additionally there are some functions available in the templates:

- `arg <idx>` - Takes the message sent to the channel, splits by space and returns the Nth element
- `botHasBadge <badge>` - Checks whether bot has the given badge in the current channel
- `channelCounter <counter name>` - Wraps the counter name into a channel specific counter name including the channel name
- `concat <delimiter> <...parts>` - Join the given string parts with delimiter
- `counterValue <counter name>` - Returns the current value of the counter which identifier was supplied
- `displayName <username> [fallback]` - Returns the display name the specified user set for themselves
- `fixUsername <username>` - Ensures the username no longer contains the `@` or `#` prefix
- `followDate <from> <to>` - Looks up when `from` followed `to`
- `group <idx> [fallback]` - Gets matching group specified by index from `match_message` regular expression, when `fallback` is defined, it is used when group has an empty match
- `recentGame <username> [fallback]` - Returns the last played game name of the specified user (see shoutout example) or the `fallback` if the game could not be fetched. If no fallback was supplied the message will fail and not be sent.
- `tag <tagname>` - Takes the message sent to the channel, returns the value of the tag specified
- `toLower <string>` - Converts the given string to lower-case
- `toUpper <string>` - Converts the given string to upper-case
- `variable <name> [default]` - Returns the variable value or default in case it is empty

For some events special variables are made available:
- `title` - Available in `title_update` - The new stream title
- `category` - Available in `category_update` - The new stream category

## Command executions

Your command will get a JSON object passed through `stdin` you can parse to gain details about the message. It is expected to yield an array of actions on `stdout` and exit with status `0`. If it does not the action will be marked failed. In case you need to output debug output you can use `stderr` which is directly piped to the bots `stderr`.

This is an example input you might get on `stdin`:

```json
{
  "badges": {
    "glhf-pledge": 1,
    "moderator": 1
  },
  "channel": "#tezrian",
  "message": "!test",
  "tags": {
    "badge-info": "",
    "badges": "moderator/1,glhf-pledge/1",
    "client-nonce": "6801c82a341f728dbbaad87ef30eae49",
    "color": "#A72920",
    "display-name": "Luziferus",
    "emotes": "",
    "flags": "",
    "id": "dca06466-3741-4b22-8339-4cb5b07a02cc",
    "mod": "1",
    "room-id": "485884564",
    "subscriber": "0",
    "tmi-sent-ts": "1610313040489",
    "turbo": "0",
    "user-id": "69699328",
    "user-type": "mod"
  },
  "username": "luziferus"
}
```

The example was dumped using this action:

```yaml
  - actions:
    - command: [/usr/bin/bash, -c, "jq . >&2"]
    match_channels: ['#tezrian']
    match_message: '^!test'
```
