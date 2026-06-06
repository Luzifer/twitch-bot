---
title: Respond to a message after random delay
---

This rule responds to greetings after a randomized delay, which makes the reply feel less immediate and automated. The configured delay waits 30 seconds plus up to 10 seconds of jitter.

<!--more-->

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
