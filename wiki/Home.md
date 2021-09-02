## Configuration

```yaml
---

# Channels to join (only those can be acted on)
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

auto_messages:
  - channel: 'mychannel'          # String, channel to send message to
    message: 'Automated message'  # String, message to send
    use_action: true              # Bool, optional, send message as action (`/me <message>`)

    # Even though all of these are optional, at least one MUST be specified for the entry to be valid
    cron: '*/10 * * * *'  # String, optional, cron syntax when to send the message
    message_interval: 3   # Integer, optional, how many non-bot-messages must be sent in between
    time_interval: 900s   # Duration, optional, how long to wait before repeating the message

    only_on_live: true    # Boolean, optional, only send the message when channel is live

    # Disable message using templating, must yield string `true` to disable the automated message
    disable_on_template: '{{ ne .myvariable true }}'

rules: # See below for examples

  - actions: # Array of actions to take when this rule matches

    # Issue a ban on the user who wrote the chat-line
    - ban: "reason of ban"

    # Command to execute for the chat message, must return an JSON encoded array of actions
    - command: [/bin/bash, -c, "echo '[{\"respond\": \"Text\"}]'"]

    # Modify an internal counter value (does NOT send a chat line)
    - counter: "counterid" # String to identify the counter, applies templating
      counter_set: 25      # String, set counter to value (counter_step is ignored if set),
                           # applies templating but MUST result in a parseable integer
      counter_step: 1      # Integer, can be negative or positive, default: +1

    # Introduce a delay between two actions
    - delay: 1m         # Duration, how long to wait (fixed)
      delay_jitter: 1m  # Duration, add random delay to fixed delay between 0 and this value

    # Issue a delete on the message caught
    - delete_message: true # Bool, set to true to delete

    # Send raw IRC message to Twitch servers
    - raw_message: 'PRIVMSG #{{ .channel }} :Test' # String, applies templating

    # Send responding message to the channel the original message was received in
    - respond: 'Hello chatter'    # String, applies templating
      respond_as_reply: true      # Boolean, optional, use Twitch-Reply feature in respond
      respond_fallback: 'Oh noes' # String, text to send if the template function causes
                                  # an error, applies templating (default: unset)

    # Issue a timeout on the user who wrote the chat-line
    - timeout: 1s # Duration value: 1s / 1m / 1h

    # Set a variable to value defined for later usage
    - variable: myvar       # String, name of the variable to set (applies templating)
      clear: false          # Boolean, clear the variable
      set: '{{ .channel }}' # String, value to set the variable to (applies templating)

    # Send a whisper (ATTENTION: You need to have a known / verified bot for this!)
    # Without being known / verified your whisper will just silently get dropped by Twitch
    # Go here to get that verification: https://dev.twitch.tv/limit-increase
    - whisper_to: '{{ .username }}' # String, username to send to, applies templating
      whisper_message: 'Ohai!'      # String, message to send, applies templating

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
- `channelCounter <counter name>` - Wraps the counter name into a channel specific counter name including the channel name
- `concat <delimiter> <...parts>` - Join the given string parts with delimiter
- `counterValue <counter name>` - Returns the current value of the counter which identifier was supplied
- `displayName <username> [fallback]` - Returns the display name the specified user set for themselves
- `fixUsername <username>` - Ensures the username no longer contains the `@` or `#` prefix
- `followDate <from> <to>` - Looks up when `from` followed `to`
- `group <idx>` - Gets matching group specified by index from `match_message` regular expression
- `recentGame <username> [fallback]` - Returns the last played game name of the specified user (see shoutout example) or the `fallback` if the game could not be fetched. If no fallback was supplied the message will fail and not be sent.
- `tag <tagname>` - Takes the message sent to the channel, returns the value of the tag specified
- `toLower <string>` - Converts the given string to lower-case
- `toUpper <string>` - Converts the given string to upper-case
- `variable <name> [default]` - Returns the variable value or default in case it is empty

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

## Rule examples

### Chat-addable generic text-respond-commands

```yaml
  # Respond with variable content if set
  - actions:
    - respond: '{{ variable (concat ":" "genericcmd" .channel (group 1)) }}'
    disable_on_template: '{{ eq (variable (concat ":" "genericcmd" .channel (group 1))) "" }}'
    match_channels: ['#mychannel']
    match_message: '^!([^\s]+)(?: |$)'

  # Set variable content to content of chat command
  - actions:
    - variable: '{{ concat ":" "genericcmd" .channel (group 1) }}'
      set: '{{ group 2 }}'
    - respond: '[Admin] Set command !{{ group 1 }} to "{{ group 2 }}"'
    enable_on: [broadcaster, moderator]
    match_channels: ['#mychannel']
    match_message: '^!setcmd ([^\s]+) (.*)'

  # Remove variable and therefore delete command
  - actions:
    - variable: '{{ concat ":" "genericcmd" .channel (group 1) }}'
      clear: true
    - respond: '[Admin] Deleted command !{{ group 1 }}'
    enable_on: [broadcaster, moderator]
    match_channels: ['#mychannel']
    match_message: '^!clearcmd ([^\s]+)'
```

### Game death counter with dynamic name

```yaml
  - actions:
    - counter: '{{ channelCounter (recentGame .channel) }}'
    - respond: >-
        I already died {{ counterValue (channelCounter (recentGame .channel)) }}
        times in {{ recentGame .channel }}'
    cooldown: 60s
    enable_on: [broadcaster, moderator]
    match_channels: ['#mychannel']
    match_message: '^!death'
```

### Link-protection while allowing Twitch clips

```yaml
  - actions:
    - timeout: 1s
    - respond: '@{{ .username }}, please ask for permission before posting links.'
    disable_on: [broadcaster, moderator, subscriber, vip]
    disable_on_match_messages:
      - '^(?:https?://)?clips\.twitch\.tv/[a-zA-Z0-9-]+$'
    disable_on_permit: true
    match_channels: ['#mychannel']
    match_message: '(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]'
```

### Post follow date for an user

```yaml
  - actions:
    - respond: 'You followed on {{ ( followDate .username ( fixUsername .channel ) ).Format "2006-01-02" }}'
    match_channels: ['#mychannel']
    match_message: '^!followage'
```

### Respond to a message after random delay

```yaml
  - actions:
    # Respond after 30-40s
    - delay: 30s
      delay_jitter: 10s
    - respond: 'Hey {{ .username }}'
    match_channels: ['#mychannel']
    match_message: '^Hi'
```

### Send a notification on successful permit

```yaml
  - actions:
    - respond: >-
        @{{ fixUsername (arg 1) }}, you will not get timed out
        for the next {{ .permitTimeout }} seconds.
    match_channels: ['#mychannel']
    match_event: 'permit'
```

### Shoutout command with game query

```yaml
  - actions:
    - respond: >-
        Check out @{{ fixUsername (group 1) }} and leave a follow,
        they were last playing {{ recentGame (fixUsername (group 1)) "something mysterious" }}
        at https://twitch.tv/{{ fixUsername (group 1) }}
    enable_on: [broadcaster, moderator]
    match_channels: ['#mychannel']
    match_message: '^!so ([@\w]+)'
```
