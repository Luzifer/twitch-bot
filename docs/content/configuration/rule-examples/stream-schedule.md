---
title: Display Stream-Schedule in Chat
---

This command lists the next three scheduled streams for the current channel. Adjust the timezone in `dateInZone` if your schedule should be displayed for a different region.

<!--more-->

```yaml
- actions:
    - type: respond
      attributes:
        message: |-
            {{- $segs := scheduleSegments .channel 3 -}}
            {{- $fmtSegs := list -}}
            {{- range $segs -}}
            {{- $fmtSegs = mustAppend $fmtSegs (
              printf "%s @ %s"
                (.Category.Name)
                (dateInZone "02.01. 15:40" .StartTime "Europe/Berlin")
            ) -}}
            {{- end -}}
            Next streams are: {{ $fmtSegs | join ", " }}
            - See more in the Twitch schedule:
            https://www.twitch.tv/{{ fixUsername .channel }}/schedule
  match_message: '!schedule\b'
```
