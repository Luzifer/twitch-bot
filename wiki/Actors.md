# Available Actions


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
