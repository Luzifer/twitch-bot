---
title: Available Events
---

## `adbreak_begin`

Ad-break has begun and ads are playing now in mentioned channel.

Fields:

- `channel` _string_ - The channel the event occurred in
- `duration` _int64_ - Duration of the ads in seconds
- `is_automatic` _bool_ - Were the ads started by the ad-manager?
- `started_at` _time.Time_ - When did the ad-break start

## `ban`

Moderator action caused a user to be banned from chat.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:

- `channel` _string_ - The channel the event occurred in
- `target_id` _string_ - The ID of the user being banned
- `target_name` _string_ - The login-name of the user being banned

## `bits`

User spent bits in the channel. The full message is available like in a normal chat message, additionally the `{{ .bits }}` field is added with the total amount of bits spent.

Fields:

- `bits` _int64_ - Total amount of bits spent in the message
- `channel` _string_ - The channel the event occurred in
- `username` _string_ - The login-name of the user who spent the bits

## `category_update`

The current category for the channel was changed. (This event has some delay to the real category change!)

Fields:

- `category` _string_ - The name of the new game / category
- `channel` _string_ - The channel the event occurred in

## `channelpoint_redeem`

A custom channel-point reward was redeemed in the given channel. (Only available when EventSub support is available and streamer granted required permissions!)

Fields:

- `channel` _string_ - The channel the event occurred in
- `reward_cost` _int64_ - Number of points the user paid for the reward
- `reward_id` _string_ - ID of the reward the user redeemed
- `reward_title` _string_ - Title of the reward the user redeemed
- `status` _string_ - Status of the reward (one of `unknown`, `unfulfilled`, `fulfilled`, and `canceled`)
- `user_id` _string_ - The ID of the user who redeemed the reward
- `user_input` _string_ - The text the user entered into the input for the reward
- `user` _string_ - The login-name of the user who redeemed the reward

## `clearchat`

Moderator action caused chat to be cleared.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:

- `channel` _string_ - The channel the event occurred in

## `delete`

Moderator action caused a chat message to be deleted.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:

- `channel` _string_ - The channel the event occurred in
- `message_id` _string_ - The UUID of the message being deleted
- `target_name` _string_ - Login name of the author of the deleted message

## `follow`

User followed the channel. This event is not de-duplicated and therefore might be used to spam! (Only available when EventSub support is available!)

Fields:

- `channel` _string_ - The channel the event occurred in
- `followed_at` _time.Time_ - Time object of the follow date
- `user_id` _string_ - ID of the newly following user
- `user` _string_ - The login-name of the user who followed

## `giftpaidupgrade`

User upgraded their gifted subscription into a paid one. This event does not contain any details about the tier of the paid subscription.

Fields:

- `channel` _string_ - The channel the event occurred in
- `gifter` _string_ - The login-name of the user who gifted the subscription
- `username` _string_ - The login-name of the user who upgraded their subscription

## `hypetrain_begin`, `hypetrain_end`, `hypetrain_progress`

An Hype-Train has begun, ended or progressed in the given channel.

Fields:

- `channel` _string_ - The channel the event occurred in
- `level` _int64_ - The current level of the Hype-Train
- `levelProgress` _float64_ - Percentage of reached "points" in the current level to complete the level (not available on `hypetrain_end`)
- `event` _EventSubEventHypetrain_ - Raw Hype-Train event, see schema in [`pkg/twitch/eventsub.go#L92`](https://github.com/Luzifer/twitch-bot/blob/master/pkg/twitch/eventsub.go#L121) 

## `join`

User joined the channel-chat. This is **NOT** an indicator they are viewing, the event is **NOT** reliably sent when the user really joined the chat. The event will be sent with some delay after they join the chat and is sometimes repeated multiple times during their stay. So **DO NOT** use this to greet users!

Fields:

- `channel` _string_ - The channel the event occurred in
- `user` _string_ - The login-name of the user who joined

## `kofi_donation`

A Ko-fi donation was received through the API-Webhook.

Fields:

- `channel` _string_ - The channel the event occurred for
- `from` _string_ - The name submitted by Ko-fi (can be arbitrarily entered)
- `amount` _float64_ - The amount donated as submitted by Ko-fi (i.e. 27.95)
- `currency` _string_ - The currency of the amount (i.e. USD)
- `isSubscription` _bool_ - true on monthly subscriptions, false on single-donations
- `isFirstSubPayment` _bool_ - true on first montly payment, false otherwise
- `message` _string_ - The message entered by the donator (**not** present when donation was marked as private!)
- `tier` _string_ - The tier the subscriber subscribed to (seems not to be filled on the first transaction?)

## `outbound_raid`

The channel has raided another channel. (The event is issued in the moment the raid is executed, not when the raid timer starts!)

Fields:

- `channel` _string_ - The channel the raid originated at
- `to` _string_ - The login-name of the channel the viewers are sent to
- `to_id` _string_ - The ID of the channel the viewers are sent to
- `viewers` _int64_ - The number of viewers included in the raid

## `part`

User left the channel-chat. This is **NOT** an indicator they are no longer viewing, the event is **NOT** reliably sent when the user really leaves the chat. The event will be sent with some delay after they leave the chat and is sometimes repeated multiple times during their stay. So this does **NOT** mean they do no longer read the chat!

Fields:

- `channel` _string_ - The channel the event occurred in
- `user` _string_ - The login-name of the user who left

## `permit`

User received a permit, which means they are no longer affected by rules which are disabled on permit.

Fields:

- `channel` _string_ - The channel the event occurred in
- `user` _string_ - The login-name of the user who **gave** the permit
- `to` _string_ - The username who got the permit

## `poll_begin` / `poll_end` / `poll_progress`

A poll was started / was ended / had changes in the given channel.

Fields:

- `channel` _string_ - The channel the event occurred in
- `poll` _EventSubEventPoll_ - The poll object describing the poll, see schema in [`pkg/twitch/eventsub.go#L92`](https://github.com/Luzifer/twitch-bot/blob/master/pkg/twitch/eventsub.go#L152)
- `status` _string_ - The status of the poll (one of `completed`, `terminated` or `archived`) - only available in `poll_end`
- `title` _string_ - The title of the poll the event was generated for

## `raid`

The channel was raided by another user.

Fields:

- `channel` _string_ - The channel the event occurred in
- `username` _string_ - The login-name of the user who raided the channel
- `viewercount` _int64_ - The amount of users who have been raided (this number is not fully accurate)

## `resub`

The user shared their resubscription. (This event is triggered manually by the user using the "Share my Resub" button and does not occur when the user does not actively share their sub!)

Fields:

- `channel` _string_ - The channel the event occurred in
- `plan` _string_ - The sub-plan they are using (`1000` = T1, `2000` = T2, `3000` = T3, `Prime`)
- `subscribed_months` _int64_ - How long have they been subscribed
- `username` _string_ - The login-name of the user who resubscribed

## `shoutout_created`

The channel gave another streamer a (Twitch native) shoutout

Fields:

- `channel` _string_ - The channel the event occurred in
- `to_id` _string_ - The ID of the channel who received the shoutout
- `to` _string_ - The login-name of the channel who received the shoutout
- `viewers` _int64_ - The amount of viewers the shoutout was shown to

## `shoutout_received`

The channel received a (Twitch native) shoutout by another channel.

Fields:

- `channel` _string_ - The channel the event occurred in
- `from_id` _string_ - The ID of the channel who issued the shoutout
- `from` _string_ - The login-name of the channel who issued the shoutout
- `viewers` _int64_ - The amount of viewers the shoutout was shown to

## `stream_offline`

The channels stream went offline. (This event has some delay to the real category change!)

Fields:

- `channel` _string_ - The channel the event occurred in

## `stream_online`

The channels stream went offline. (This event has some delay to the real category change!)

Fields:

- `channel` _string_ - The channel the event occurred in

## `sub`

The user newly subscribed on their own. (This event is triggered automatically and does not need to be shared actively!)

Fields:

- `channel` _string_ - The channel the event occurred in
- `plan` _string_ - The sub-plan they are using (`1000` = T1, `2000` = T2, `3000` = T3, `Prime`)
- `username` _string_ - The login-name of the user who subscribed

## `subgift`

The user gifted the subscription to a specific user. (This event **DOES** occur multiple times after `submysterygift` events!)

Fields:

- `channel` _string_ - The channel the event occurred in
- `gifted_months` _int64_ - Number of months the user gifted
- `origin_id` _string_ - ID unique to the gift-event (can be used to match `subgift` events to corresponding `submysterygift` event)
- `plan` _string_ - The sub-plan they are using (`1000` = T1, `2000` = T2, `3000` = T3, `Prime`)
- `subscribed_months` _int64_ - How long the recipient has been subscribed
- `to` _string_ - The user who received the sub
- `total_gifted` _int64_ - How many subs has the user given in total (might be zero due to users preferences)
- `username` _string_ - The login-name of the user who gifted the subscription

## `submysterygift`

The user gifted multiple subs to the community. (This event is followed by `number x subgift` events.)

Fields:

- `channel` _string_ - The channel the event occurred in
- `number` _int64_ - The amount of gifted subs
- `origin_id` _string_ - ID unique to the gift-event (can be used to match `subgift` events to corresponding `submysterygift` event)
- `plan` _string_ - The sub-plan they are using (`1000` = T1, `2000` = T2, `3000` = T3, `Prime`)
- `username` _string_ - The login-name of the user who gifted the subscription

## `sus_user_message`

A suspicious (monitored / restricted) user sent a message in the given channel

- `ban_evasion` _string_ - Status of the ban-evasion detection: `unknown`, `possible`, `likely`
- `channel` _string_ - The channel in which the event occurred
- `message` _string_ - The message the user sent in plain text
- `shared_ban_channels` _[]string_ - IDs of channels with shared ban-info in which the user is also banned
- `status` _string_ - Restriction status: `active_monitoring`, `restricted`
- `user_id` _string_ - ID of the user sending the message
- `user_type` _[]string_ - How the user ended being on the naughty-list: `manually_added`, `ban_evader_detector`, or `shared_channel_ban`
- `username` _string_ - The login-name of the user sending the message

## `sus_user_update`

The status of suspicious user was changed by a moderator

- `channel` _string_ - The channel in which the event occurred
- `moderator` _string_ - The login-name of the moderator changing the status
- `status` _string_ - Restriction status: `no_treatment`, `active_monitoring`, `restricted`
- `user_id` _string_ - ID of the suspicious user
- `username` _string_ - Login-name of the suspicious user

## `timeout`

Moderator action caused a user to be timed out from chat.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:

- `channel` _string_ - The channel the event occurred in
- `duration` _time.Duration_ - The timeout duration (nanoseconds)
- `seconds` _int_ - The timeout duration (seconds)
- `target_id` _string_ - The ID of the user being timed out 
- `target_name` _string_ - The login-name of the user being timed out 

## `title_update`

The current title for the channel was changed. (This event has some delay to the real category change!)

Fields:

- `channel` _string_ - The channel the event occurred in
- `title` _string_ - The title of the stream

## `whisper`

The bot received a whisper message. (You can use `(.*)` as message match and `{{ group 1 }}` as template to get the content of the whisper.)

Fields:

- `username` _string_ - The login-name of the user who sent the message
