---
title: 'Deathcounter: Query, Set, Increase and Decrease in one single rule'
---

This rule handles querying, setting, incrementing, and decrementing a game-specific death counter through one `!dc` command. Permission checks restrict setting and lowering the counter to trusted chat roles.

<!--more-->

```yaml
- description: Category DeathCounter
  actions:
    - type: eventmod
      attributes:
        fields: |-
          {{
            dict
              "game" (
                default
                  (recentGame .channel)
                  (
                    variable (
                      list .channel "game-override" | join ":"
                    ) ""
                  )
              )
            | mustToJson
          }}
    - type: eventmod
      attributes:
        fields: |-
          {{
            dict
              "counterName" (
                list .channel "deathCounter" .game | join ":"
              )
            | mustToJson
          }}
    - type: counter
      attributes:
        counter: '{{ .counterName }}'
        counter_set: |-
          {{-
            if and
              (eq (group 1 "") "set")
              (ne (group 3 "") "")
          -}}
          {{- group 3 -}}
          {{- end -}}
        counter_step: |
          {{-
            if and
              (ne (group 1 "") "set")
              (ne (group 2 "") "")
          -}}
          {{- group 2 }}{{ group 3 "1" -}}
          {{- else -}}
          0
          {{- end -}}
    - type: eventmod
      attributes:
        fields: |
          {{- $action := "" -}}
          {{-
            if and
              (eq (group 1 "") "set")
              (ne (group 3 "") "")
          -}}
            {{- $action = "set" -}}
          {{-
            else if and
              (ne (group 1 "") "set")
              (ne (group 2 "") "")
          -}}
            {{- $action = printf "%s%s" (group 2) (group 3 "1") -}}
          {{- end -}}
          {{-
            dict
              "action" $action
              "chanName" (displayName
                .channel
                (fixUsername .channel))
              "value" (counterValue .counterName)
            | mustToJson
          -}}
    - type: respond
      attributes:
        message: |
          {{ .chanName }} died in "{{ .game }}"
          {{ .value }} times.
          {{ with .action }}({{ . }}){{ end }}
  match_message: (?i)^!(set)?dc([ +-])?([0-9]+)?(?:\s|$)
  disable_on_template: |-
    {{-
      or
        (and
          (eq (group 1 "") "set")
          (not (chatterHasBadge "broadcaster"))
          (not (chatterHasBadge "moderator"))
        )
        (and
          (eq (group 2 "") "-")
          (not (chatterHasBadge "broadcaster"))
          (not (chatterHasBadge "moderator"))
          (not (chatterHasBadge "vip"))
        )
        (and
          (eq (group 2 "") "+")
          (not (chatterHasBadge "broadcaster"))
          (not (chatterHasBadge "moderator"))
          (not (chatterHasBadge "vip"))
          (not (chatterHasBadge "subscriber"))
        )
    -}}
```
