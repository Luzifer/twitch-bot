# Available Actions


## Ban User

Ban user from chat

```yaml
- type: ban
  attributes:
    # Reason why the user was banned
    # Optional: true
    # Type:     string
    reason: ""
```

## Delay

Delay next action

```yaml
- type: delay
  attributes:
    # Static delay to wait
    # Optional: true
    # Type:     duration
    delay: 0s
    # Dynamic jitter to add to the static delay (the added extra delay will be between 0 and this value)
    # Optional: true
    # Type:     duration
    jitter: 0s
```

## Delete Message

Delete message which caused the rule to be executed

```yaml
- type: delete
  # Does not have configuration attributes
```

## Execute Script / Command

Execute external script / command

```yaml
- type: script
  attributes:
    # Command to execute
    # Optional: false
    # Type:     array of strings (Supports Templating in each string)
    command: []
    # Do not activate cooldown for route when command exits non-zero
    # Optional: true
    # Type:     bool
    skip_cooldown_on_error: false
```

## Modify Counter

Update counter values

```yaml
- type: counter
  attributes:
    # Name of the counter to update
    # Optional: false
    # Type:     string (Supports Templating)
    counter: ""
    # Value to add to the counter
    # Optional: true
    # Type:     int64
    counter_step: 1
    # Value to set the counter to
    # Optional: true
    # Type:     string (Supports Templating)
    counter_set: ""
```

## Modify Stream

Update stream information

```yaml
- type: modchannel
  attributes:
    # Channel to update
    # Optional: false
    # Type:     string (Supports Templating)
    channel: ""
    # Category / Game to set
    # Optional: true
    # Type:     string (Supports Templating)
    game: ""
    # Stream title to set
    # Optional: true
    # Type:     string (Supports Templating)
    title: ""
```

## Modify Variable

Modify variable contents

```yaml
- type: setvariable
  attributes:
    # Name of the variable to update
    # Optional: false
    # Type:     string (Supports Templating)
    variable: ""
    # Clear variable content and unset the variable
    # Optional: true
    # Type:     bool
    clear: false
    # Value to set the variable to
    # Optional: true
    # Type:     string (Supports Templating)
    set: ""
```

## Punish User

Apply increasing punishments to user

```yaml
- type: punish
  attributes:
    # When to lower the punishment level after the last punishment
    # Optional: true
    # Type:     duration
    cooldown: 168h
    # Actions for each punishment level (ban, delete, duration-value i.e. 1m)
    # Optional: false
    # Type:     array of strings
    levels: []
    # Reason why the user was banned / timeouted
    # Optional: true
    # Type:     string
    reason: ""
    # User to apply the action to
    # Optional: false
    # Type:     string (Supports Templating)
    user: ""
    # Unique identifier for this punishment to differentiate between punishments in the same channel
    # Optional: true
    # Type:     string
    uuid: ""
```

## Quote Database

Manage a database of quotes in your channel

```yaml
- type: quotedb
  attributes:
    # Action to execute (one of: add, del, get)
    # Optional: false
    # Type:     string
    action: ""
    # Index of the quote to work with, must yield a number (required on 'del', optional on 'get')
    # Optional: true
    # Type:     string (Supports Templating)
    index: "0"
    # Quote to add: Format like you like your quote, nothing is added (required on: add)
    # Optional: true
    # Type:     string (Supports Templating)
    quote: ""
    # Format to use when posting a quote (required on: get)
    # Optional: true
    # Type:     string (Supports Templating)
    format: "Quote #{{ .index }}: {{ .quote }}"
```

## Reset User Punishment

Reset punishment level for user

```yaml
- type: reset-punish
  attributes:
    # User to reset the level for
    # Optional: false
    # Type:     string (Supports Templating)
    user: ""
    # Unique identifier for this punishment to differentiate between punishments in the same channel
    # Optional: true
    # Type:     string
    uuid: ""
```

## Respond to Message

Respond to message with a new message

```yaml
- type: respond
  attributes:
    # Message text to send
    # Optional: false
    # Type:     string (Supports Templating)
    message: ""
    # Fallback message text to send if message cannot be generated
    # Optional: true
    # Type:     string (Supports Templating)
    fallback: ""
    # Send message as a native Twitch-reply to the original message
    # Optional: true
    # Type:     bool
    as_reply: false
    # Send message to a different channel than the original message
    # Optional: true
    # Type:     string
    to_channel: ""
```

## Send RAW Message

Send raw IRC message

```yaml
- type: raw
  attributes:
    # Raw message to send (must be a valid IRC protocol message)
    # Optional: false
    # Type:     string (Supports Templating)
    message: ""
```

## Send Whisper

Send a whisper (requires a verified bot!)

```yaml
- type: whisper
  attributes:
    # Message to whisper to the user
    # Optional: false
    # Type:     string (Supports Templating)
    message: ""
    # User to send the message to
    # Optional: false
    # Type:     string (Supports Templating)
    to: ""
```

## Timeout User

Timeout user from chat

```yaml
- type: timeout
  attributes:
    # Duration of the timeout
    # Optional: false
    # Type:     duration
    duration: 0s
```
