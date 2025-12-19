---
title: Script Actions
---

> [!TIP]
> In order to maximize the flexibility of the bot you can trigger external scripts / commands in rules. These scripts are provided with extensive data to act on.

Your command will get a JSON object passed through `stdin` you can parse to gain details about the message. It is expected to yield an array of actions on `stdout` and exit with status `0`. If it does not the action will be marked failed. In case you need to output debug output you can use `stderr` which is directly piped to the bots `stderr`.

This is an example input you might get on `stdin`:

```json
{
  "badges": {
    "glhf-pledge": 1,
    "moderator": 1
  },
  "channel": "#tezrian",
  "message": "!test",
  "tags": {
    "badge-info": "",
    "badges": "moderator/1,glhf-pledge/1",
    "client-nonce": "6801c82a341f728dbbaad87ef30eae49",
    "color": "#A72920",
    "display-name": "Luziferus",
    "emotes": "",
    "flags": "",
    "id": "dca06466-3741-4b22-8339-4cb5b07a02cc",
    "mod": "1",
    "room-id": "485884564",
    "subscriber": "0",
    "tmi-sent-ts": "1610313040489",
    "turbo": "0",
    "user-id": "69699328",
    "user-type": "mod"
  },
  "username": "luziferus"
}
```

The example was dumped using this action:

```yaml
  - actions:
    - type: script
      attributes:
        command: [/usr/bin/bash, -c, "jq . >&2"]
    match_channels: ['#tezrian']
    match_message: '^!test'
```
