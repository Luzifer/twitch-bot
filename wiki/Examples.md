# Rule examples

## Chat-addable generic text-respond-commands

```yaml
  # Respond with variable content if set
  - actions:
    - type: respond
      attributes:
        message: '{{ variable (list "genericcmd" .channel (group 1) | join ":") }}'
    disable_on_template: '{{ eq (variable (list "genericcmd" .channel (group 1) | joih ":")) "" }}'
    match_channels: ['#mychannel']
    match_message: '^!([^\s]+)(?: |$)'

  # Set variable content to content of chat command
  - actions:
    - type: setvariable
      attributes:
        variable: '{{ list "genericcmd" .channel (group 1) | join ":" }}'
        set: '{{ group 2 }}'
    - type: respond
      attributes:
        message: '[Admin] Set command !{{ group 1 }} to "{{ group 2 }}"'
    enable_on: [broadcaster, moderator]
    match_channels: ['#mychannel']
    match_message: '^!setcmd ([^\s]+) (.*)'

  # Remove variable and therefore delete command
  - actions:
    - type: setvariable
      attributes:
        variable: '{{ list "genericcmd" .channel (group 1) | join ":" }}'
        clear: true
    - type: respond
      attributes:
        message: '[Admin] Deleted command !{{ group 1 }}'
    enable_on: [broadcaster, moderator]
    match_channels: ['#mychannel']
    match_message: '^!clearcmd ([^\s]+)'
```

## Game death counter with dynamic name

```yaml
  - actions:
    - type: counter
      attributes:
        counter: '{{ channelCounter (recentGame .channel) }}'
    - type: respond
      attributes:
        message: >-
          I already died {{ counterValue (channelCounter (recentGame .channel)) }}
          times in {{ recentGame .channel }}'
    cooldown: 60s
    enable_on: [broadcaster, moderator]
    match_channels: ['#mychannel']
    match_message: '^!death'
```

## Link-protection while allowing Twitch clips

```yaml
  - actions:
    - type: timeout
      attributes:
        duration: 1s
    - type: respond
      attributes:
        message: '@{{ .username }}, please ask for permission before posting links.'
    disable_on: [broadcaster, moderator, subscriber, vip]
    disable_on_match_messages:
      - '^(?:https?://)?clips\.twitch\.tv/[a-zA-Z0-9-]+$'
    disable_on_permit: true
    match_channels: ['#mychannel']
    match_message: '(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]'
```

## Post follow date for an user

```yaml
  - actions:
    - type: respond
      attributes:
        message: 'You followed on {{ ( followDate .username ( fixUsername .channel ) ).Format "2006-01-02" }}'
    match_channels: ['#mychannel']
    match_message: '^!followage'
```

## Respond to a message after random delay

```yaml
  - actions:
    # Respond after 30-40s
    - type: delay
      attributes:
        delay: 30s
        jitter: 10s
    - type: respond
      attributes:
        message: 'Hey {{ .username }}'
    match_channels: ['#mychannel']
    match_message: '^Hi'
```

## Send a notification on successful permit

```yaml
  - actions:
    - type: respond
      attributes:
        message: >-
          @{{ fixUsername (arg 1) }}, you will not get timed out
          for the next {{ .permitTimeout }} seconds.
    match_channels: ['#mychannel']
    match_event: 'permit'
```

## Shoutout command with game query

```yaml
  - actions:
    - type: respond
      attributes:
        message: >-
          Check out @{{ fixUsername (group 1) }} and leave a follow,
          they were last playing {{ recentGame (fixUsername (group 1)) "something mysterious" }}
          at https://twitch.tv/{{ fixUsername (group 1) }}
    enable_on: [broadcaster, moderator]
    match_channels: ['#mychannel']
    match_message: '^!so ([@\w]+)'
```

