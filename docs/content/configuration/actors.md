---
title: "Available Actions"
---

{{< lead >}}
All these actions can be executed by your bot as soon as you add them to rules. Read their documentation to learn how to master them.
{{< /lead >}}


## Add Fields to Event

Add custom fields to the event to be used as template variables later on

```yaml
- type: eventmod
  attributes:
    # Fields to set in the event (must produce valid JSON: `map[string]any`)
    # Optional: false
    # Type:     string (Supports Templating)
    fields: ""
```

## Add VIP

Add VIP for the given channel

```yaml
- type: vip
  attributes:
    # Channel to add the VIP to
    # Optional: false
    # Type:     string (Supports Templating)
    channel: ""
    # User to add as VIP
    # Optional: false
    # Type:     string (Supports Templating)
    user: ""
```

## Ban User

Ban user from chat

```yaml
- type: ban
  attributes:
    # Reason why the user was banned
    # Optional: false
    # Type:     string (Supports Templating)
    reason: ""
```

## Commercial

Start Commercial

```yaml
- type: commercial
  attributes:
    # Duration of the commercial (must not be longer than 180s and must yield an integer)
    # Optional: false
    # Type:     string (Supports Templating)
    duration: ""
```

## Create Clip

Triggers the creation of a Clip from the given channel owned by the creator (subsequent actions can use variables `create_clip_slug` and `create_clip_edit_url`)

```yaml
- type: clip
  attributes:
    # Channel to create the clip from, defaults to the channel of the event / message
    # Optional: true
    # Type:     string (Supports Templating)
    channel: ""
    # User which should trigger and therefore own the clip (must have given clips:edit permission to the bot in extended permissions!), defaults to the value of `channel`
    # Optional: true
    # Type:     string (Supports Templating)
    creator: ""
    # Whether to add an artificial delay before creating the clip
    # Optional: true
    # Type:     bool
    add_delay: false
```

## Custom Event

Create a custom event

```yaml
- type: customevent
  attributes:
    # JSON representation of fields in the event (`map[string]any`)
    # Optional: false
    # Type:     string (Supports Templating)
    fields: "{}"
    # Time until the event is triggered (must be valid duration like 1h, 1h1m, 10s, ...)
    # Optional: true
    # Type:     string (Supports Templating)
    schedule_in: ""
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

## Discord Message-Webhook

Sends a message to a Discord Web-hook

```yaml
- type: discordhook
  attributes:
    # URL to send the POST request to
    # Optional: false
    # Type:     string
    hook_url: ""
    # Overwrites the username set in the webhook configuration
    # Optional: true
    # Type:     string (Supports Templating)
    username: ""
    # Overwrites the avatar set in the webhook configuration
    # Optional: true
    # Type:     string (Supports Templating)
    avatar_url: ""
    # Message content to send to the web-hook (this must be set if embed is disabled)
    # Optional: true
    # Type:     string (Supports Templating)
    content: ""
    # Whether to include the embed in the post
    # Optional: true
    # Type:     bool
    add_embed: false
    # Title of the embed
    # Optional: true
    # Type:     string (Supports Templating)
    embed_title: ""
    # Description of the embed
    # Optional: true
    # Type:     string (Supports Templating)
    embed_description: ""
    # URL the title should link to
    # Optional: true
    # Type:     string (Supports Templating)
    embed_url: ""
    # URL of the big image displayed in the embed
    # Optional: true
    # Type:     string (Supports Templating)
    embed_image: ""
    # URL of the small image displayed in the embed
    # Optional: true
    # Type:     string (Supports Templating)
    embed_thumbnail: ""
    # Name of the post author (if empty all other author-fields are ignored)
    # Optional: true
    # Type:     string (Supports Templating)
    embed_author_name: ""
    # URL the author name should link to
    # Optional: true
    # Type:     string (Supports Templating)
    embed_author_url: ""
    # URL of the author avatar
    # Optional: true
    # Type:     string (Supports Templating)
    embed_author_icon_url: ""
    # Fields to display in the embed (must yield valid JSON: `[{"name": "", "value": "", "inline": false}]`)
    # Optional: true
    # Type:     string (Supports Templating)
    embed_fields: ""
```

## Enforce Link-Protection

Uses link- and clip-scanner to detect links / clips and applies link protection as defined

```yaml
- type: linkprotect
  attributes:
    # Allowed links (if any is specified all non matching links will cause enforcement action, link must contain any of these strings)
    # Optional: true
    # Type:     array of strings
    allowed_links: []
    # Disallowed links (if any is specified all non matching links will not cause enforcement action, link must contain any of these strings)
    # Optional: true
    # Type:     array of strings
    disallowed_links: []
    # Allowed clip channels (if any is specified clips of all other channels will cause enforcement action, clip-links will be ignored in link-protection when this is used)
    # Optional: true
    # Type:     array of strings
    allowed_clip_channels: []
    # Disallowed clip channels (if any is specified clips of all other channels will not cause enforcement action, clip-links will be ignored in link-protection when this is used)
    # Optional: true
    # Type:     array of strings
    disallowed_clip_channels: []
    # Enforcement action to take when disallowed link / clip is detected (ban, delete, duration-value i.e. 1m)
    # Optional: false
    # Type:     string
    action: ""
    # Reason why the enforcement action was taken
    # Optional: false
    # Type:     string
    reason: ""
    # Stop rule execution when action is applied (i.e. not to post a message after a ban for spam links)
    # Optional: true
    # Type:     bool
    stop_on_action: false
    # Stop rule execution when no action is applied (i.e. not to post a message when no enforcement action is taken)
    # Optional: true
    # Type:     bool
    stop_on_no_action: false
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

## FileSay

Takes the content of an URL and pastes it to the current channel

```yaml
- type: filesay
  attributes:
    # Source of the content to post
    # Optional: false
    # Type:     string (Supports Templating)
    source: ""
```

## Log output

Print info log-line to bot log

```yaml
- type: log
  attributes:
    # Messsage to log into bot-log
    # Optional: false
    # Type:     string (Supports Templating)
    message: ""
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
    # Type:     string (Supports Templating)
    counter_step: "1"
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
    # Category / Game to set (use `@1234` format to pass an explicit ID)
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

## Nuke Chat

Mass ban, delete, or timeout messages based on regex. Be sure you REALLY know what you do before using this! Used wrongly this will cause a lot of damage!

```yaml
- type: nuke
  attributes:
    # How long to scan into the past, template must yield a duration (max 10m)
    # Optional: true
    # Type:     string (Supports Templating)
    scan: "10m"
    # What action to take when message matches (delete / ban / <timeout duration>)
    # Optional: true
    # Type:     string (Supports Templating)
    action: "delete"
    # Regular expression (RE2) to select matching messages
    # Optional: false
    # Type:     string (Supports Templating)
    match: ""
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

## Remove VIP

Remove VIP for the given channel

```yaml
- type: unvip
  attributes:
    # Channel to remove the VIP from
    # Optional: false
    # Type:     string (Supports Templating)
    channel: ""
    # User to remove as VIP
    # Optional: false
    # Type:     string (Supports Templating)
    user: ""
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

## Scan for Clips

Scans for clip-links in the message and adds the "clips" field to the event data

```yaml
- type: clipdetector
  # Does not have configuration attributes
```

## Scan for Links

Scans for links in the message and adds the "links" field to the event data

```yaml
- type: linkdetector
  attributes:
    # Enable heuristic scans to find links with spaces or other means of obfuscation in them
    # Optional: true
    # Type:     bool
    heuristic: false
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

## Shoutout

Perform a Twitch-native shoutout

```yaml
- type: shoutout
  attributes:
    # User to give the shoutout to
    # Optional: false
    # Type:     string (Supports Templating)
    user: ""
```

## Slack Message-Webhook

Sends a message to a Slack(-compatible) Web-hook

```yaml
- type: slackhook
  attributes:
    # URL to send the POST request to
    # Optional: false
    # Type:     string
    hook_url: ""
    # Text to send to the web-hook
    # Optional: false
    # Type:     string (Supports Templating)
    text: ""
```

## Stop Execution

Stop Rule Execution on Condition

```yaml
- type: stopexec
  attributes:
    # Condition when to stop execution (must evaluate to "true" to stop execution)
    # Optional: false
    # Type:     string (Supports Templating)
    when: ""
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
    # Reason why the user was timed out
    # Optional: false
    # Type:     string (Supports Templating)
    reason: ""
```

## Update Shield Mode

Update shield mode for the given channel

```yaml
- type: shield
  attributes:
    # Whether the shield-mode should be enabled or disabled
    # Optional: false
    # Type:     bool
    enable: false
```
