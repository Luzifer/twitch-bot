---
title: "Rule Examples"
---

{{< lead >}}
These are only a few examples of rules. If you want to share your own rules and show what can be achieved with the bot, head over to the ["Share your Rules"](https://github.com/Luzifer/twitch-bot/discussions/categories/share-your-rules) discussion board and create a new discussion with the YAML definition and a description what your rule does.
{{< /lead >}}

## Chat-addable generic text-respond-commands

```yaml
- uuid: 688e631f-08a8-5544-b4b2-1737ea71ce00
  description: Trigger Generic Command
  actions:
    - type: respond
      attributes:
        message: '{{ variable (list "genericcmd" .channel (group 1) | join ":") }}'
  cooldown: 1m0s
  match_channels:
    - '#luziferus'
    - '#tezrian'
  match_message: '^!([^\s]+)(?: |$)'
  disable_on_template: '{{ eq (variable (list "genericcmd" .channel (group 1) | join ":")) "" }}'

- uuid: ba4f7bb3-af39-5c57-bb97-216a8af69246
  description: Set Generic Command
  actions:
    - type: setvariable
      attributes:
        set: '{{ group 2 }}'
        variable: '{{ list "genericcmd" .channel (group 1) | join ":" }}'
    - type: respond
      attributes:
        message: '[Admin] Set command !{{ group 1 }} to "{{ group 2 }}"'
  match_channels:
    - '#luziferus'
    - '#tezrian'
  match_message: ^!setcmd ([^\s]+) (.*)
  enable_on:
    - broadcaster
    - moderator

- uuid: 21619e80-2c6a-536e-8b83-e5fe6c580356
  description: Clear Generic Command
  actions:
    - type: setvariable
      attributes:
        clear: true
        variable: '{{ list "genericcmd" .channel (group 1) | join ":" }}'
    - type: respond
      attributes:
        message: '[Admin] Deleted command !{{ group 1 }}'
  match_channels:
    - '#luziferus'
    - '#tezrian'
  match_message: ^!clearcmd ([^\s]+)
  enable_on:
    - broadcaster
    - moderator
```

## Display Stream-Schedule in Chat

```yaml
- actions:
    - type: respond
      attributes:
        message: |-
            {{- $segs := scheduleSegments .channel 3 -}}
            {{- $fmtSegs := list -}}
            {{- range $segs -}}
            {{- $fmtSegs = mustAppend $fmtSegs (
              printf "%s @ %s"
                (.Category.Name)
                (dateInZone "02.01. 15:40" .StartTime "Europe/Berlin")
            ) -}}
            {{- end -}}
            Next streams are: {{ $fmtSegs | join ", " }}
            - See more in the Twitch schedule:
            https://www.twitch.tv/{{ fixUsername .channel }}/schedule
  match_message: '!schedule\b'
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

## Notify Discord when stream is live

```yaml
  - actions:
      - type: discordhook
        attributes:
          add_embed: true
          avatar_url: '{{ profileImage .channel }}'
          content: |
            <@&123456789012345678> {{ displayName (fixUsername .channel) (fixUsername .channel) }}
            is now live on https://www.twitch.tv/{{ fixUsername .channel }} - join us!
          embed_author_icon_url: '{{ profileImage .channel }}'
          embed_author_name: '{{ displayName (fixUsername .channel) (fixUsername .channel) }}'
          embed_fields: |
              {{
                toJson (
                  list
                    (dict
                      "name" "Game"
                      "value" (recentGame .channel))
                )
              }}
          embed_image: https://static-cdn.jtvnw.net/previews-ttv/live_user_{{ fixUsername .channel }}-1280x720.jpg
          embed_thumbnail: '{{ profileImage .channel }}'
          embed_title: '{{ recentTitle .channel }}'
          embed_url: https://twitch.tv/{{ fixUsername .channel }}
          hook_url: https://discord.com/api/webhooks/[...]/[...]
          username: 'Stream-Live: {{ displayName (fixUsername .channel) (fixUsername .channel) }}'
    match_event: stream_online
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
        message: '{{ mention .to }}, you will not get timed out for the next {{ .permitTimeout }} seconds.'
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

## VIP of the Month

In this example a channel-point reward for one month of VIP status exists which is automated through the bot. To get the `.reward_id` you can use the [Debug-Overlay]({{< ref "../overlays/_index.md" >}}) or just use the `.reward_title` variable and check against the name of the reward.

```yaml
  - description: 'Channel-Point-Reward: VIP of the month'
    actions:
      - type: vip
        attributes:
          channel: '{{ .channel }}'
          user: '{{ .user }}'
      - type: customevent
        attributes:
          fields: |-
            {{
              toJson (
                dict
                  "targetChannel" .channel
                  "targetUser" .user
                  "type" "timed_unvip"
              )
            }}
          schedule_in: 744h
    match_event: channelpoint_redeem
    disable_on_template: '{{ ne .reward_id "aaa66d18-8dab-46f4-a222-6ff228f2fdfb" }}'
    disable_on: [moderator, vip]

  - description: 'Channel-Point-Reward: Remove VIP of the month'
    actions:
      - type: unvip
        attributes:
          channel: '{{ .targetChannel }}'
          user: '{{ .targetUser }}'
    match_event: custom
    disable_on_template: '{{ ne .type "timed_unvip" }}'
```
