---
title: VIP of the Month
---

This example automates a channel-point reward that grants VIP status for one month and later removes it again. To get the `.reward_id` you can use the [Debug-Overlay]({{< ref "../../overlays/_index.md" >}}) or just use the `.reward_title` variable and check against the name of the reward.

<!--more-->

```yaml
  - description: 'Channel-Point-Reward: VIP of the month'
    actions:
      - type: vip
        attributes:
          channel: '{{ .channel }}'
          user: '{{ .user }}'
      - type: customevent
        attributes:
          fields: |-
            {{
              toJson (
                dict
                  "targetChannel" .channel
                  "targetUser" .user
                  "type" "timed_unvip"
              )
            }}
          schedule_in: 744h
    match_event: channelpoint_redeem
    disable_on_template: '{{ ne .reward_id "aaa66d18-8dab-46f4-a222-6ff228f2fdfb" }}'
    disable_on: [moderator, vip]

  - description: 'Channel-Point-Reward: Remove VIP of the month'
    actions:
      - type: unvip
        attributes:
          channel: '{{ .targetChannel }}'
          user: '{{ .targetUser }}'
    match_event: custom
    disable_on_template: '{{ ne .type "timed_unvip" }}'
```
