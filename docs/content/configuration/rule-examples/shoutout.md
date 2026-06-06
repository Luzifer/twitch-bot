---
title: Shoutout command with game query
---

This moderator command posts a shoutout for another Twitch channel and includes the last known game for that channel. The target user is taken from the first command argument.

<!--more-->

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
