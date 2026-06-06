---
title: Notify Discord when stream is live
---

This rule sends a Discord webhook message when the channel goes live. Replace the role mention and `hook_url` with the Discord role and webhook for your server.

<!--more-->

```yaml
  - actions:
      - type: discordhook
        attributes:
          add_embed: true
          avatar_url: '{{ profileImage .channel }}'
          content: |
            <@&123456789012345678> {{ displayName (fixUsername .channel) (fixUsername .channel) }}
            is now live on https://www.twitch.tv/{{ fixUsername .channel }} - join us!
          embed_author_icon_url: '{{ profileImage .channel }}'
          embed_author_name: '{{ displayName (fixUsername .channel) (fixUsername .channel) }}'
          embed_fields: |
              {{
                toJson (
                  list
                    (dict
                      "name" "Game"
                      "value" (recentGame .channel))
                )
              }}
          embed_image: https://static-cdn.jtvnw.net/previews-ttv/live_user_{{ fixUsername .channel }}-1280x720.jpg
          embed_thumbnail: '{{ profileImage .channel }}'
          embed_title: '{{ recentTitle .channel }}'
          embed_url: https://twitch.tv/{{ fixUsername .channel }}
          hook_url: https://discord.com/api/webhooks/[...]/[...]
          username: 'Stream-Live: {{ displayName (fixUsername .channel) (fixUsername .channel) }}'
    match_event: stream_online
```
