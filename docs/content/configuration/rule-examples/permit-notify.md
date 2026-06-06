---
title: Send a notification on successful permit
---

This rule confirms a successful permit event in chat and tells the target user how long the permit is valid. It relies on the permit event fields `.to` and `.permitTimeout`.

<!--more-->

```yaml
  - actions:
    - type: respond
      attributes:
        message: '{{ mention .to }}, you will not get timed out for the next {{ .permitTimeout }} seconds.'
    match_channels: ['#mychannel']
    match_event: 'permit'
```
