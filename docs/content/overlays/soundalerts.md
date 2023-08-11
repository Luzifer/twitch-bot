---
title: Sound-Alerts
weight: 50
---

{{< lead >}}
The Sound-Alerts overlay provides you with a way to let your viewers trigger sounds through channel-points and chat commands.
{{< /lead >}}

This overlay utilizes the custom events actor to generate events specific to the overlay in order to trigger sound alerts.

## OBS Setup

To use it

- generate an API token with the `overlays` permission and note it in a secure place
- add a Browser-Source to your OBS scenes with at least 1×1 px in size
- set the URL to `https://your-bot.example.com/overlays/sounds.html?token=[your-token]&channel=[your-channel]`

After you've done this you're already done with the setup inside your OBS.

## Add the sound-files to the overlays directory

In order to be able to play those sounds the sounds need to be added to overlays directory. You cannot use remote sounds hosted somewhere else because of security limitations.

The Linux OBS browser-source does not allow for proprietary codecs to be used so the sounds should be converted to sound files containing the Opus Codec. (Though this isn't strictly required for Windows OBS it might be a good idea to use this codec in general as it is a non-proprietary codec.)

If you got a sound file in `mp3`, `wav`, `aac` or whatever format you can use [`ffmpeg`](https://ffmpeg.org/) to convert it to a `webm` file which automatically will use the Opus codec:

```console
ffmpeg mysoundfile.mp3 mysoundfile.webm
```

In the following example rules I'll assume you created a `sounds` directory inside your overlays directory and stored some sound files in there.

## Creating Rules to trigger Sounds

### Using Channel-Points

```yaml
- description: Channelpoint SoundAlert
  actions:
    - type: customevent
      attributes:
        fields: |
            {{
              $sound := get ( dict
                "b8afa5fa-2441-43ef-a634-77d0ad4f2691" "applaus"
                "63c8d002-89f4-41fc-b6c9-03b339f439cd" "badum"
                "39c34f65-85e3-4792-9e16-ff3d186aaf6f" "bonk"
                "b9f41ea4-3fcb-41d8-9072-f0923089d097" "tick"
                "f67c792e-2555-486e-b441-8247cb8ac97e" "typewriter"
                "c9ca69a6-9d2e-4b39-90c9-26a4b62064fc" "zirp"
              ) .reward_id
            }}
            {
              "type": "soundalert",
              "soundUrl": "sounds/{{ $sound }}.webm"
            }
  match_channels:
    - '#luziferus'
  match_event: channelpoint_redeem
  disable_on_offline: true
  disable_on_template: '{{ not (hasPrefix "Sound: " .reward_title) }}'
```

On the first glimps this seems rather complicated so lets look at it step-by-step:

- We defined a new rule for our Channel-Points Sound-Alerts
- This rule reacts on redeemed Channel-Points: `match_event: channelpoint_redeem`
- It does not trigger when the channel is offline and it does not trigger when the Channel-Points reward is not named `Sound: [...]`
- It only triggers in my channel: `match_channels: ['#luziferus']`
- When all the conditions mentioned above are met the rule will create a `customevent` containing the `"type": "soundalert"` which is used by the overlay we've set up in OBS before
- It gives the `"soundUrl": "sounds/{{ $sound }}.webm"`, therefore telling the overlay to play the file `sounds/{{ $sound }}.webm` when the event is received
- Now to the "complicated" part: In order not to define a rule for each single sound I've fused all the rewards together into one rule with a mapping of the `.reward_id` to the name of the sound file to be played. This is done through creating a `dict` with the reward id and the name of the file and a `get` function around it selecting the `.reward_id` from the given `dict`.

Before adding this to your bot create at least one **Custom Reward** in the channel-points section in your streamer dashboard. The **Reward Name** should be `Sound: [...]` in order to be matched by the rule above. The **Reward Description** is optional, the viewer should not enter a text, the **Cost** is something you can freely define (I'm using 250 points), the **Reward Icon** too. I'd advice to enable **Skip Reward Requests Queue** as the bot will automatically trigger and disabling this would clutter your Reward Requests Queue. Also you might want to enable the **Cooldown**: I'm using a 5min cooldown on each sound for them not to be spammed permanently.

You now can simply copy the rule into your bot configuration and just adjust the channel, the reward-ids and the sound file names. To get the `reward_id` of the event have a look into the Debug-Overlay described on the [Overlays]({{< ref "_index.md" >}}) page and trigger the reward once. (If you're doing this being offline just disable the cooldown for a moment to be able to trigger the reward while being offline.)

### Using Chat-Commands

```yaml
- description: Command SoundAlert
  actions:
    - type: customevent
      attributes:
        fields: |
            {
              "type": "soundalert",
              "soundUrl": "sounds/{{ group 1 }}.webm"
            }
  cooldown: 5m
  match_channels:
    - '#luziferus'
  match_message: '(?i)^!sound (applaus|badum|bonk|tick)$'
  disable_on_offline: true
```

As we don't have to deal with the Twitch reward-id stuff and just define a command this rule is rather simple but lets have a quick walkthrough too:

- We defined a rule for the `!sound` command: `match_message: '(?i)^!sound (applaus|badum|bonk|tick)$'`
- To add a little bit of security here the command also defines which sounds can be triggered: `(applaus|badum|bonk|tick)`
- Also this rule does not trigger when the channel is offline and defines a 5 minute cooldown (this is not per sound but for all sounds so triggering `!sound applaus` will set the cooldown also for `!sound tick`)
- Finally the rule also creates the `customevent` of `"type": "soundalert"` passing a `soundUrl` where to find the sound
