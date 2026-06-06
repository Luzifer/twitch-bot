---
title: Auto-Shoutout on Raid
---

This example combines manual, native, and raid-triggered shoutouts. Moderators can trigger a shoutout with `!so`, native Twitch shoutouts are mirrored into chat with a text response, and incoming raids automatically get a short welcome message followed by a delayed shoutout.

<!--more-->

```yaml
- uuid: 5063bc07-8da1-4731-aed3-4ce06819b90f
  description: 'Shoutout: Manual Trigger'
  actions:
    - type: shoutout
      attributes:
        user: '{{ group 1 }}'
  match_message: ^!so (.*)
  disable_on_offline: true
  enable_on:
    - broadcaster
    - moderator
    
- uuid: fdd024d2-a292-5c8f-8d90-f0abfb2880d7
  description: 'Shoutout: Native SO to Text'
  actions:
    - type: respond
      attributes:
        message: >-
          Check out {{ mention .to }} at https://twitch.tv/{{ .to }}
          and leave a follow. They were last streaming
          {{ recentGame .to "something mysterious" }}…
  match_channels:
    - '#luziferus'
  match_event: shoutout_created

- uuid: 58c4c1b9-ea3d-5b40-9a93-041c3f078c73
  description: 'Shoutout: Auto-Shoutout on Raid'
  actions:
    - type: respond
      attributes:
        message: >-
          Ahhhhhh! A raid! To arms! Protect the treasure
          chamber! luzife4Knife luzife4Loot luzife4Knife
    - type: delay
      attributes:
        delay: 30s
    - type: shoutout
      attributes:
        user: '{{ tag "login" }}'
  match_channels:
    - '#luziferus'
  match_event: raid
  disable_on_offline: true
```
