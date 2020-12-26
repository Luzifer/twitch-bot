## Configuration

```yaml
---

# Channels to join (only those can be acted on)
channels:
  - mychannel

# Allow moderators to hand out permits (if set to false only broadcaster can do this)
permit_allow_moderator: true
# How long to permit on !permit command
permit_timeout: 60s

rules: # See below for examples

  - actions: # Array of actions to take when this rule matches

    # Issue a ban on the user who wrote the chat-line
    - ban: "reason of ban"

    # Modify an internal counter value (does NOT send a chat line)
    - counter: "counterid" # String to identify the counter, applies templating
      counter_step: 1      # Integer, can be negative or positive, default: +1

    # Send responding message to the channel the original message was received in
    - respond: 'Hello chatter' # String, applies templating

    # Issue a timeout on the user who wrote the chat-line
    - timeout: 1s # Duration value: 1s / 1m / 1h

    # Add a cooldown to the command (not to trigger counters twice, ...)
    cooldown: 1s # Duration value: 1s / 1m / 1h

    # Disable actions on this rule if the user has an active permit
    disable_on_permit: false

    # Disable actions on this rule if the user has one of these badges
    disable_on: [broadcaster, moderator]

    # Enable actions on this rule only if the user has one of these badges
    enable_on: [broadcaster, moderator]

    # Require the chat message to be sent in this channel
    match_channels: ['#mychannel']

    # Execute actions when this event occurs
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
- `counterValue <counter name>` - Returns the current value of the counter which identifier was supplied
- `fixUsername <username>` - Ensures the username no longer contains the `@` prefix
- `group <idx>` - Gets matching group specified by index from `match_message` regular expression
- `recentGame <username> [fallback]` - Returns the last played game name of the specified user (see shoutout example) or the `fallback` if the game could not be fetched. If no fallback was supplied the message will fail and not be sent.
- `tag <tagname>` - Takes the message sent to the channel, returns the value of the tag specified

## Rule examples

### Game death counter with dynamic name

```yaml
  - actions:
    - counter: '{{ recentGame "mychannel" }}'
    - respond: 'I already died {{ counterValue (recentGame "mychannel") }} times in {{ recentGame "mychannel" }}'
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
