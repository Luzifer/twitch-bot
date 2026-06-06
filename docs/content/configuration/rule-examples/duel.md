---
title: 'Viewer-Interaction: Let a viewer challenge another to a duel'
---

This example lets viewers challenge each other to a duel with `!duel`, then accept or deny the challenge from chat. A scheduled custom event clears unanswered challenges after one minute.

<!--more-->

```yaml
- uuid: 78d340e8-fa5b-42f7-bcf2-07063daa4aed
  description: 'Duel: Challenge'
  actions:
    # Store the challenger and therefore enable duelling
    - type: setvariable
      attributes:
        set: '{{ .username }}'
        variable: '{{ list "duel" .channel (lower (fixUsername (group 1))) | join ":" }}'
    # Tell the challenged what they need to do
    - type: respond
      attributes:
        message: |
          {{ mention (group 1) }} you've been challenged to a duel
          by {{ mention .username }} - Type !accept or !deny to
          respond within 1 minute
    # Create a custom event to clear the challenge
    - type: customevent
      attributes:
        fields: |-
          {{
            dict
              "type" "clear-duel"
              "var" (list "duel" .channel (lower (fixUsername (group 1))) | join ":")
              "challenged" (group 1)
            | mustToJson
          }}
        schedule_in: 1m # How long should the challenge be active?
  user_cooldown: 5m0s # Prevent spam
  match_message: ^!duel @?([^\s]+)

- uuid: 5dfaa0f9-436b-4bbb-92ef-269c37deeb35
  description: 'Duel: Clear Challenge'
  actions:
    # Do nothing of the variable is no longer set (challenge was executed)
    - type: stopexec
      attributes:
        when: '{{ eq (variable .var "") "" }}'
    # Notify the challenger the duel timed out
    - type: respond
      attributes:
        message: |-
          {{ mention .challenged }} chickened out, therefore I
          declare {{ mention (variable .var) }} as the winner
          of the duel!
    # Clear the variable and therefore end the duel
    - type: setvariable
      attributes:
        clear: true
        variable: '{{ .var }}'
  match_event: custom
  disable_on_template: '{{ ne .type "clear-duel" }}'

- uuid: 3516859f-dfd4-4d6a-bac0-99e319fb3a19
  description: 'Duel: Respond'
  actions:
    # Retrieve the challenger
    - type: eventmod
      attributes:
        fields: |-
          {{
            dict
              "challenger" (variable (list "duel" .channel (lower (fixUsername .username)) | join ":") "")
            | mustToJson
          }}
    # If there is no challenger, there is no challenge: Stop!
    - type: stopexec
      attributes:
        when: '{{ eq .challenger "" }}'
    # Get a random int and print the result of the challenge
    - type: respond
      attributes:
        message: |
          {{ if eq (group 1) "accept" }}
          {{ mention .username }} and {{ mention .challenger }}
          take their positions and on 3… 2… 1…
          {{ if gt (randInt 1 101) 50 }}
          {{ mention .username }} lands the wining hit! Congrats!
          {{- else -}}
          {{ mention .challenger }} leaves {{ mention .username }}
          no chance and wins! Congrats!
          {{- end -}}
          {{- else -}}
          {{ mention .username }} doesn't feel like duelling and
          silently backs up from the challenge
          {{- end -}}
    # Clear the variable and therefore end the duel
    - type: setvariable
      attributes:
        clear: true
        variable: '{{ list "duel" .channel (lower (fixUsername .username)) | join ":" }}'
  match_message: ^!(accept|deny)\b
```
