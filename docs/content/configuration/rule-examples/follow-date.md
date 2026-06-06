---
title: Post follow date for an user
---

This command replies with the date the current chatter followed the channel. It uses the `followDate` template helper and formats the result as `YYYY-MM-DD`.

<!--more-->

```yaml
  - actions:
    - type: respond
      attributes:
        message: 'You followed on {{ ( followDate .username ( fixUsername .channel ) ).Format "2006-01-02" }}'
    match_channels: ['#mychannel']
    match_message: '^!followage'
```
