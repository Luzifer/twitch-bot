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

    # Command to execute for the chat message, must return an JSON encoded array of actions
    - command: [/bin/bash, -c, "echo '[{\"respond\": \"Text\"}]'"]

    # Modify an internal counter value (does NOT send a chat line)
    - counter: "counterid" # String to identify the counter, applies templating
      counter_set: 25      # String, set counter to value (counter_step is ignored if set),
                           # applies templating but MUST result in a parseable integer
      counter_step: 1      # Integer, can be negative or positive, default: +1

    # Issue a delete on the message caught
    - delete_message: true # Bool, set to true to delete

    # Send responding message to the channel the original message was received in
    - respond: 'Hello chatter'    # String, applies templating
      respond_fallback: 'Oh noes' # String, text to send if the template function causes
                                  # an error, applies templating (default: unset)

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

    # Require the chat message to be sent by one of these users
    match_users: ['mychannel'] # List of users, all names MUST be all lower-case

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
- `channelCounter <counter name>` - Wraps the counter name into a channel specific counter name including the channel name
- `counterValue <counter name>` - Returns the current value of the counter which identifier was supplied
- `fixUsername <username>` - Ensures the username no longer contains the `@` or `#` prefix
- `followDate <from> <to>` - Looks up when `from` followed `to`
- `group <idx>` - Gets matching group specified by index from `match_message` regular expression
- `recentGame <username> [fallback]` - Returns the last played game name of the specified user (see shoutout example) or the `fallback` if the game could not be fetched. If no fallback was supplied the message will fail and not be sent.
- `tag <tagname>` - Takes the message sent to the channel, returns the value of the tag specified
- `toLower <string>` - Converts the given string to lower-case
- `toUpper <string>` - Converts the given string to upper-case

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
