---
title: Simple 8ball without any means of remembering the answer
---

This command replies to `!8ball` with one random Magic 8 Ball style answer. It does not store any state, so repeated questions are answered independently.

<!--more-->

```yaml
- description: 8ball
  actions:
    - type: respond
      attributes:
        as_reply: true
        message: |-
          {{
            randomString
              "It is certain."
              "It is decidedly so."
              "Without a doubt."
              "Yes definitely."
              "You may rely on it."
              "As I see it, yes."
              "Most likely."
              "Outlook good."
              "Yes."
              "Signs point to yes."

              "Reply hazy, try again."
              "Ask again later."
              "Better not tell you now."
              "Cannot predict now."
              "Concentrate and ask again."

              "Don't count on it."
              "My reply is no."
              "My sources say no."
              "Outlook not so good."
              "Very doubtful."
          }}
  match_message: (?i)^!8ball\b
```
