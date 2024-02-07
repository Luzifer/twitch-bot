---
title: "Available Events"
---

## `ban`

Moderator action caused a user to be banned from chat.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:

- `channel` - The channel the event occurred in
- `target_id` - The ID of the user being banned
- `target_name` - The login-name of the user being banned

## `bits`

User spent bits in the channel. The full message is available like in a normal chat message, additionally the `{{ .bits }}` field is added with the total amount of bits spent.

Fields:

- `bits` - Total amount of bits spent in the message
- `channel` - The channel the event occurred in
- `username` - The login-name of the user who spent the bits

## `category_update`

The current category for the channel was changed. (This event has some delay to the real category change!)

Fields:

- `category` - The name of the new game / category
- `channel` - The channel the event occurred in

## `channelpoint_redeem`

A custom channel-point reward was redeemed in the given channel. (Only available when EventSub support is available and streamer granted required permissions!)

Fields:

- `channel` - The channel the event occurred in
- `reward_cost` - Number of points the user paid for the reward
- `reward_id` - ID of the reward the user redeemed
- `reward_title` - Title of the reward the user redeemed
- `status` - Status of the reward (one of `unknown`, `unfulfilled`, `fulfilled`, and `canceled`)
- `user_id` - The ID of the user who redeemed the reward
- `user_input` - The text the user entered into the input for the reward
- `user` - The login-name of the user who redeemed the reward

## `clearchat`

Moderator action caused chat to be cleared.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:

- `channel` - The channel the event occurred in

## `delete`

Moderator action caused a chat message to be deleted.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:

- `channel` - The channel the event occurred in
- `message_id` - The UUID of the message being deleted
- `target_name` - Login name of the author of the deleted message

## `follow`

User followed the channel. This event is not de-duplicated and therefore might be used to spam! (Only available when EventSub support is available!)

Fields:

- `channel` - The channel the event occurred in
- `followed_at` - Time object of the follow date
- `user_id` - ID of the newly following user
- `user` - The login-name of the user who followed

## `giftpaidupgrade`

User upgraded their gifted subscription into a paid one. This event does not contain any details about the tier of the paid subscription.

Fields:

- `channel` - The channel the event occurred in
- `gifter` - The login-name of the user who gifted the subscription
- `username` - The login-name of the user who upgraded their subscription

## `join`

User joined the channel-chat. This is **NOT** an indicator they are viewing, the event is **NOT** reliably sent when the user really joined the chat. The event will be sent with some delay after they join the chat and is sometimes repeated multiple times during their stay. So **DO NOT** use this to greet users!

Fields:

- `channel` - The channel the event occurred in
- `user` - The login-name of the user who joined

## `outbound_raid`

The channel has raided another channel. (The event is issued in the moment the raid is executed, not when the raid timer starts!)

Fields:

- `channel` - The channel the raid originated at
- `to` - The login-name of the channel the viewers are sent to
- `to_id` - The ID of the channel the viewers are sent to
- `viewers` - The number of viewers included in the raid

## `part`

User left the channel-chat. This is **NOT** an indicator they are no longer viewing, the event is **NOT** reliably sent when the user really leaves the chat. The event will be sent with some delay after they leave the chat and is sometimes repeated multiple times during their stay. So this does **NOT** mean they do no longer read the chat!

Fields:

- `channel` - The channel the event occurred in
- `user` - The login-name of the user who left

## `permit`

User received a permit, which means they are no longer affected by rules which are disabled on permit.

Fields:

- `channel` - The channel the event occurred in
- `user` - The login-name of the user who **gave** the permit
- `to` - The username who got the permit

## `poll_begin` / `poll_end` / `poll_progress`

A poll was started / was ended / had changes in the given channel.

Fields:

- `channel` - The channel the event occurred in
- `poll` - The poll object describing the poll, see schema in [`pkg/twitch/eventsub.go#L92`](https://github.com/Luzifer/twitch-bot/blob/master/pkg/twitch/eventsub.go#L92)
- `status` - The status of the poll (one of `completed`, `terminated` or `archived`) - only available in `poll_end`
- `title` - The title of the poll the event was generated for

## `raid`

The channel was raided by another user.

Fields:

- `channel` - The channel the event occurred in
- `username` - The login-name of the user who raided the channel
- `viewercount` - The amount of users who have been raided (this number is not fully accurate)

## `resub`

The user shared their resubscription. (This event is triggered manually by the user using the "Share my Resub" button and does not occur when the user does not actively share their sub!)

Fields:

- `channel` - The channel the event occurred in
- `plan` - The sub-plan they are using (`1000` = T1, `2000` = T2, `3000` = T3, `Prime`)
- `subscribed_months` - How long have they been subscribed
- `username` - The login-name of the user who resubscribed

## `shoutout_created`

The channel gave another streamer a (Twitch native) shoutout

Fields:

- `channel` - The channel the event occurred in
- `to_id` - The ID of the channel who received the shoutout
- `to` - The login-name of the channel who received the shoutout
- `viewers` - The amount of viewers the shoutout was shown to

## `shoutout_received`

The channel received a (Twitch native) shoutout by another channel.

Fields:

- `channel` - The channel the event occurred in
- `from_id` - The ID of the channel who issued the shoutout
- `from` - The login-name of the channel who issued the shoutout
- `viewers` - The amount of viewers the shoutout was shown to

## `stream_offline`

The channels stream went offline. (This event has some delay to the real category change!)

Fields:

- `channel` - The channel the event occurred in

## `stream_online`

The channels stream went offline. (This event has some delay to the real category change!)

Fields:

- `channel` - The channel the event occurred in

## `sub`

The user newly subscribed on their own. (This event is triggered automatically and does not need to be shared actively!)

Fields:

- `channel` - The channel the event occurred in
- `plan` - The sub-plan they are using (`1000` = T1, `2000` = T2, `3000` = T3, `Prime`)
- `username` - The login-name of the user who subscribed

## `subgift`

The user gifted the subscription to a specific user. (This event **DOES** occur multiple times after `submysterygift` events!)

Fields:

- `channel` - The channel the event occurred in
- `gifted_months` - Number of months the user gifted
- `origin_id` - ID unique to the gift-event (can be used to match `subgift` events to corresponding `submysterygift` event)
- `plan` - The sub-plan they are using (`1000` = T1, `2000` = T2, `3000` = T3, `Prime`)
- `subscribed_months` - How long the recipient has been subscribed
- `to` - The user who received the sub
- `username` - The login-name of the user who gifted the subscription

## `submysterygift`

The user gifted multiple subs to the community. (This event is followed by `number x subgift` events.)

Fields:

- `channel` - The channel the event occurred in
- `number` - The amount of gifted subs
- `origin_id` - ID unique to the gift-event (can be used to match `subgift` events to corresponding `submysterygift` event)
- `plan` - The sub-plan they are using (`1000` = T1, `2000` = T2, `3000` = T3, `Prime`)
- `username` - The login-name of the user who gifted the subscription

## `timeout`

Moderator action caused a user to be timed out from chat.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:

- `channel` - The channel the event occurred in
- `duration` - The timeout duration (`time.Duration`, nanoseconds)
- `seconds` - The timeout duration (`int`, seconds)
- `target_id` - The ID of the user being timed out 
- `target_name` - The login-name of the user being timed out 

## `title_update`

The current title for the channel was changed. (This event has some delay to the real category change!)

Fields:

- `channel` - The channel the event occurred in
- `title` - The title of the stream

## `whisper`

The bot received a whisper message. (You can use `(.*)` as message match and `{{ group 1 }}` as template to get the content of the whisper.)

Fields:

- `username` - The login-name of the user who sent the message
