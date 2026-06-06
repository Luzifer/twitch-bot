---
title: Hype Train chat alert
---

These rules announce the start, progress, and end of a Hype Train in chat. A variable tracks whether a Hype Train is active so progress messages are skipped after it has ended.

<!--more-->

```yaml
- uuid: 20796a4d-8bb2-43c4-a6be-bf9531c16c7e
  description: 'EVENTS: Hypetrain start'
  actions:
    - type: delay
      attributes:
        delay: 1s
    - type: respond
      attributes:
        message: >-
          The Hype Train started!
          {{ printf "%.0f" (mulf .levelProgress 100) }}% towards
          level {{ .level }} are done
    - type: setvariable
      attributes:
        set: "1"
        variable: hypetrain_active
  match_event: hypetrain_begin

- uuid: 0eca6fe7-f1a6-4c93-9deb-87ac79238bd7
  description: 'EVENTS: Hypetrain progress'
  actions:
    - type: stopexec
      attributes:
        when: '{{ eq (variable "hypetrain_active" "0") "0"  }}'
    - type: respond
      attributes:
        message: >-
          The Hype Train is still going!
          {{ printf "%.0f" (mulf .levelProgress 100) }}% towards
          level {{ .level }} are done!
  channel_cooldown: 1m0s
  match_event: hypetrain_progress

- uuid: 8bb27234-76a2-49e7-ba18-2af8b8715ab1
  description: 'EVENTS: Hypetrain end'
  actions:
    - type: respond
      attributes:
        message: >-
          {{ if (eq (add .level -1) 0)}} The Hype Train ended! It
          didn't work this time, maybe next time {{ else }} The Hype
          Train ended! We finished level {{ (add .level -1) }}! Thanks
          for your support {{ end }}
    - type: setvariable
      attributes:
        set: "0"
        variable: hypetrain_active
  match_event: hypetrain_end
```
