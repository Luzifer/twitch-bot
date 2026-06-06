---
title: Announce a new poll started / summarize an ended poll
---

These rules announce Twitch polls in chat when they start and summarize the choices with vote counts when completed polls end. Non-completed poll-end events are ignored.

<!--more-->

```yaml
- uuid: 512dd769-a5dc-4506-ac35-e5f5b0db7613
  description: 'Poll: Announce New Poll'
  actions:
    - type: respond
      attributes:
        message: |
          {{- $opts := list -}}
          {{- range $idx, $opt := .poll.Choices -}}
          {{-
            $opts = (
              append $opts
                (printf "/vote %d -> %q" (add $idx 1) $opt.Title)
              )
          -}}
          {{- end -}}
          /me -> Eine Umfrage ist gestartet: "{{ .poll.Title }}"
          luzife4Note
          {{ $opts | join " | " }}
  match_event: poll_begin

- uuid: 68b021b3-2f1b-4ef8-bd52-d2ffe52f7932
  description: 'Poll: Announce Poll End'
  actions:
    - type: respond
      attributes:
        message: |
          {{- $opts := list -}}
          {{- range $opt := .poll.Choices -}}
          {{-
            $opts = (
              append $opts
                (printf "%q (%d)" $opt.Title $opt.Votes)
              )
          -}}
          {{- end -}}
          /me -> Die Umfrage "{{ .poll.Title }}" ist beendet:
          {{ $opts | join " | " }}
          luzife4Note
  match_event: poll_end
  disable_on_template: '{{ ne .status "completed" }}'
```
