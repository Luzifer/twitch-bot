---
title: Raffle
---

> [!TIP]
> Using the raffle module you can create giveaways with various settings, timers and pick one or multiple winners. You just have to send the good yourself…

## General Overview

![]({{< static "raffle-overview.png" >}})

In the overview you can see a list of your raffles and their status. You can edit, start / stop, copy, delete them or access the list of entrants from here.

The screenshot above shows one draft of a raffle together with one currently active.

![]({{< static "raffle-entrants.png" >}})

You can access the entrants list through the "group of people" button in the raffle overview. This becomes available as soon as the raffle has started.

In this list you can see the status, the nickname and the time of entry for each entrant. The status will be a person (<i class="fas fa-user"></i>) for someone joined through the **Everyone** allowance, a heart (<i class="fas fa-heart"></i>) for a follower, a star (<i class="fas fa-star"></i>) for a subscriber, a diamond (<i class="fas fa-gem"></i>) for a VIP and a coin (<i class="fas fa-coins"></i>) for someone who joined through a channel-point redeem. The list will update itself when there are changes in the entree-list.

![]({{< static "raffle-entrants-closed.png" >}})

After the raffle has been closed (either through the timer or by clicking the button) a winner can be picked through the "Pick Winner" button. A winner will display a crown before their status, the first chat message after being picked below the name and a "recycle" button to re-draw them. If you choose to re-draw a winner the crown will get striked and greyed out. A re-drawn winner can not be picked again (re-rdrawing without any candidates will cause an error and void the slot)!

### Recommendations

- As you can see below there are many options to configure and you probably will use the same texts (and maybe even some other options too) in every raffle. That's why you can see a `TEMPLATE` raffle in the screenshot above. That's a fully configured and never started raffle I keep around. When starting a new raffle I will just use the "copy" button to create a copy of that raffle, edit the copy, adjust the title and maybe options I don't like and can start the raffle saving a lot of work in the process especially creating the texts. You can have as many templates as you like if you're doing different raffles over and over.

## Raffle Configuration

### General Settings

![]({{< static "raffle-general-config.png" >}})

Within the general settings you will configure how your raffle behaves once started:

- **Channel** configures where it will take place: Straight forward, put your channel without the leading `#`.
- **Keyword** is what users must type in order to participate. In general they are used to type commands like `!enter` or `!key` for Steam-Key giveaways so I'd advice to stick to a command format. You should ensure not to use the same command in two raffles active at the same time though it is possible: if two raffles are using the same keyword, the user writing the keyword once will enter **both** raffles.
- **Title** should reflect what you're giving away and is available in the texts (more below). So for example this could be `Steam-Key: Starfield`.
- **Allowed Entries** configure who can take part in your giveaway. Pay attention these conditions are **or**-connected, so the chatter must only have one condition matching and not all of them!
  - **Everyone** is straight forward: No conditions are imposed.
  - **Followers, since `X` min** means all followers can participate if they are followed at least `X` minutes ago.
  - **Subscribers** is straight forward again: If they have a subscriber-badge, they can join.
  - **VIPs** is the same just they do need a VIP badge.
- **Luck Modifiers** are kinda tricky as they configure the size of each ticket and therefore manipulate the probability to be chosen. As stated in the text below the base modifier for **Everyone** is `1.0` and you can modifiy the "luck" for all others. Pay attention with this: if you for example disable the VIPs checkbox in **Allowed Entries** the VIPs luck modifier will **not** be used! That VIP will then enter as a subscriber or follower or even as "everyone" and get the respective modifier.
- **Times** configure when and for how long the raffle will take place:
  - **Auto-Start** can be configured and if it is, the raffle will open itself at that point of time (within 1 minute). If you don't set this you need to start the raffle yourself.
  - **Duration** configures how long the raffle will run. This is only relevant if **Close At** is unset. (Internally on starting the raffle **Close At** will be set to `now + duration` if **Close At** is empty.)
  - **Close At** marks the end of the raffle. The raffle will automatically get closed at that point of time (within 1 minute).
  - **Respond in** adds a time window where the bot will record the first message of the picked user after they have been picked. You will see that message in the entrants list: Useful for channels with a lot going on in the chat so you don't miss their response. After this time window is over no response will be recorded.

### Texts

![]({{< static "raffle-texts.png" >}})

The texts do support templating and do have the same format like other templates i.e. in rules. You can enable or disable each of them though I'd recommend to keep all of them enabled (maybe except the "failed entry" message).

- **Message on successful entry** will be posted as soon as the chatter is added to the entrants list.
- **Message on failed entry** will be posted in case the chatter is not entered (could be they are already entered or the bot encountered any other error while adding them).
- **Message on winner draw** will be posted for the chatter getting picked when drawing the winner: if you disable this you still can tell them they won when picking them.
- **Periodic reminder every `X` min** is a message to remember chatters (and tell new ones) there is a raffle open. It will be posted every `X` minutes, first time when opening the raffle.
- **Message on raffle close** will be posted when the raffle closes (either you closed it manually or the **Close At** time is reached).

Within the templates you do have access to the variables `.user` and `.raffle` (which represents the raffle object). Have a look at the default templates for examples what you can do with them.

## Using Channel-Point Rewards to join

To create a raffle to be entered through channel-point rewards you'll do the basic setup of your raffle as usual but you'll do some special adjustments:

- Set the raffle **Keyword** to something no user will ever use in chat (must be one word, can be a bunch of random characters), if a user can guess this, they can enter without using the channel points
- Doesn't matter what you select for **Allowed Entries** (the channel-point actor will ignore that setting)
- Ensure no text contains the `{{ .raffle.Keyword }}` template directive (you don't want to "leak" your keyword)
- Create a Channel-Point reward:
  - Name it as you like (but make the name unique among all your rewards as we will use that to determine whether to trigger the rule), set the points to the amount of channel points you like, put limits on it as you like
  - You can enable "Skip Queue" but in that case points will be lost when no raffle is active or if any user redeems it more than once per raffle, if you don't set this you can refund the points manually but also you need to mark all raffle entries completed manually.
- Create a new rule:
  - Channel: Limit to your channel
  - Event: `channelpoint_redeem`
  - Disable on template: `{{ ne .reward_title "<the name you chose for the reward>" }}`
  - Action: **Enter User to Raffle**, for the keyword enter the same as in the raffle

When an user redeems that reward, the rule will be triggered and if a raffle is active with that keyword, the user will be entered into that raffle as if they triggered the keyword themselves.

**Tip:** If no raffle is active disable / pause the reward to prevent users to waste points on it while there is no raffle active.
