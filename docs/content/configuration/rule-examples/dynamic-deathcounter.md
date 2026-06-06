---
title: Game death counter with dynamic name
---

This command increments a death counter named after the channel's current game and reports the updated count. It is limited to broadcaster and moderator use in `#mychannel`.

<!--more-->

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
