---
title: Chat-addable generic text-respond-commands
---

These rules let moderators create, use, and delete simple text commands from chat. The command text is stored in variables keyed by channel and command name.

<!--more-->

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
