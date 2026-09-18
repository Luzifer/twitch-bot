---
title: Available Events
---

## `adbreak_begin`

Ad-break has begun and ads are playing now in mentioned channel.

Fields:


- `channel` _string_ - Channel the event occurred in
- `duration` _int64_ - Duration of the ad break in seconds
- `is_automatic` _bool_ - Whether the ad break was started automatically
- `started_at` _time.Time_ - Time the ad break started

## `announcement`

An announcement was sent in chat.

Fields:


- `channel` _string_ - Channel the event occurred in
- `color` _string_ - Announcement color
- `message` _string_ - Announcement text
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `ban`

Moderator action caused a user to be banned from chat.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:


- `channel` _string_ - Channel the event occurred in
- `target_id` _string_ - ID of the user being banned
- `target_name` _string_ - Login name of the user being banned

## `bits`

User spent bits in the channel. The full message is available like in a normal chat message, additionally the `{{ .bits }}` field is added with the total amount of bits spent.

Fields:


- `bits` _int64_ - Total amount of bits spent in the message
- `channel` _string_ - Channel the event occurred in
- `message` _string_ - Chat message containing the bits
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `category_update`

The current category for the channel was changed. (This event has some delay to the real category change!)

Fields:


- `category` _string_ - New stream category
- `channel` _string_ - Channel the event occurred in

## `channelpoint_redeem`

A custom channel-point reward was redeemed in the given channel. (Only available when EventSub support is available and streamer granted required permissions!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `reward_cost` _int64_ - Number of points paid for the reward
- `reward_id` _string_ - ID of the redeemed reward
- `reward_title` _string_ - Title of the redeemed reward
- `status` _string_ - Status of the redemption
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event
- `user_input` _string_ - Text entered for the reward

## `clearchat`

Moderator action caused chat to be cleared.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:


- `channel` _string_ - Channel the event occurred in

## `custom`

A custom event was created through the `customevent` action or API.

Fields:

- `channel` _string_ - Channel the event occurred in

- Any additional fields passed to the custom event

## `delete`

Moderator action caused a chat message to be deleted.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:


- `channel` _string_ - Channel the event occurred in
- `message_id` _string_ - UUID of the message being deleted
- `target_name` _string_ - Login name of the author of the deleted message

## `follow`

User followed the channel. This event is not de-duplicated and therefore might be used to spam! (Only available when EventSub support is available!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `followed_at` _time.Time_ - Time the user followed the channel
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `giftpaidupgrade`

User upgraded their gifted subscription into a paid one. This event does not contain any details about the tier of the paid subscription.

Fields:


- `channel` _string_ - Channel the event occurred in
- `gifter` _string_ - Login name of the user who gifted the subscription
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `hypetrain_begin`

A Hype-Train has begun in the given channel.

Fields:


- `channel` _string_ - Channel the event occurred in
- `event` _twitch.EventSubEventHypetrain_ - Raw Hype-Train event
- `level` _int64_ - Current Hype-Train level
- `levelProgress` _float64_ - Progress towards the next Hype-Train level

## `hypetrain_end`

A Hype-Train has ended in the given channel.

Fields:


- `channel` _string_ - Channel the event occurred in
- `event` _twitch.EventSubEventHypetrain_ - Raw Hype-Train event
- `level` _int64_ - Current Hype-Train level

## `hypetrain_progress`

A Hype-Train has progressed in the given channel.

Fields:


- `channel` _string_ - Channel the event occurred in
- `event` _twitch.EventSubEventHypetrain_ - Raw Hype-Train event
- `level` _int64_ - Current Hype-Train level
- `levelProgress` _float64_ - Progress towards the next Hype-Train level

## `join`

User joined the channel-chat. This is **NOT** an indicator they are viewing, the event is **NOT** reliably sent when the user really joined the chat. The event will be sent with some delay after they join the chat and is sometimes repeated multiple times during their stay. So **DO NOT** use this to greet users!

Fields:


- `channel` _string_ - Channel the event occurred in
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `kofi_donation`

A Ko-fi donation was received through the API-Webhook.

Fields:


- `amount` _float64_ - Amount donated as submitted by Ko-fi
- `channel` _string_ - Channel the event occurred for
- `currency` _string_ - Currency of the donated amount
- `from` _string_ - Name submitted by the donor
- `isFirstSubPayment` _bool_ - Whether this is the first subscription payment
- `isSubscription` _bool_ - Whether this is a subscription payment
- `message` _*string_ _(optional)_ - Message entered by the donor
- `tier` _*string_ _(optional)_ - Subscription tier

## `moderator_add`

A user was added as a moderator to the channel. (Only available when EventSub support is available and the streamer granted the required permission!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `moderator_remove`

A user was removed as a moderator from the channel. (Only available when EventSub support is available and the streamer granted the required permission!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `outbound_raid`

The channel has raided another channel. (The event is issued in the moment the raid is executed, not when the raid timer starts!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `to` _string_ - Login name of the channel receiving the raid
- `to_id` _string_ - ID of the channel receiving the raid
- `viewers` _int64_ - Number of viewers included in the raid

## `part`

User left the channel-chat. This is **NOT** an indicator they are no longer viewing, the event is **NOT** reliably sent when the user really leaves the chat. The event will be sent with some delay after they leave the chat and is sometimes repeated multiple times during their stay. So this does **NOT** mean they do no longer read the chat!

Fields:


- `channel` _string_ - Channel the event occurred in
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `permit`

User received a permit, which means they are no longer affected by rules which are disabled on permit.

Fields:


- `channel` _string_ - Channel the event occurred in
- `to` _string_ - Login name of the user who received the permit
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `poll_begin`

A poll was started in the given channel.

Fields:


- `channel` _string_ - Channel the event occurred in
- `hasChannelPointVoting` _bool_ - Whether channel-point voting is enabled
- `poll` _twitch.EventSubEventPoll_ - Raw poll event
- `title` _string_ - Poll title

## `poll_end`

A poll ended in the given channel.

Fields:


- `channel` _string_ - Channel the event occurred in
- `hasChannelPointVoting` _bool_ - Whether channel-point voting is enabled
- `poll` _twitch.EventSubEventPoll_ - Raw poll event
- `status` _string_ - Poll status
- `title` _string_ - Poll title

## `poll_progress`

A poll changed in the given channel.

Fields:


- `channel` _string_ - Channel the event occurred in
- `hasChannelPointVoting` _bool_ - Whether channel-point voting is enabled
- `poll` _twitch.EventSubEventPoll_ - Raw poll event
- `title` _string_ - Poll title

## `primepaidupgrade`

User upgraded their Prime subscription into a paid one.

Fields:


- `channel` _string_ - Channel the event occurred in
- `plan` _string_ - Paid subscription plan
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `raid`

The channel was raided by another user.

Fields:


- `channel` _string_ - Channel the event occurred in
- `from` _string_ - Login name of the user who raided the channel
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event
- `viewercount` _int64_ - Number of viewers included in the raid

## `resub`

The user shared their resubscription. (This event is triggered manually by the user using the "Share my Resub" button and does not occur when the user does not actively share their sub!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `from` _string_ - Login name of the user who resubscribed
- `message` _string_ - Message shared with the resubscription
- `multi_month` _int64_ - Multi-month duration in months reported by Twitch
- `plan` _string_ - Subscription plan
- `subscribed_months` _int64_ - Number of months the user has been subscribed
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `shoutout_created`

The channel gave another streamer a (Twitch native) shoutout

Fields:


- `channel` _string_ - Channel the event occurred in
- `to` _string_ - Login name of the channel receiving the shoutout
- `to_id` _string_ - ID of the channel receiving the shoutout
- `viewers` _int64_ - Number of viewers shown the shoutout

## `shoutout_received`

The channel received a (Twitch native) shoutout by another channel.

Fields:


- `channel` _string_ - Channel the event occurred in
- `from` _string_ - Login name of the channel issuing the shoutout
- `from_id` _string_ - ID of the channel issuing the shoutout
- `viewers` _int64_ - Number of viewers shown the shoutout

## `stream_offline`

The channels stream went offline. (This event has some delay to the real button-press to "stop stream"!)

Fields:


- `channel` _string_ - Channel the event occurred in

## `stream_online`

The channels stream went online. (This event has some delay to the real button-press to "start stream"!)

Fields:


- `channel` _string_ - Channel the event occurred in

## `sub`

The user newly subscribed on their own. (This event is triggered automatically and does not need to be shared actively!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `from` _string_ - Login name of the user who subscribed
- `multi_month` _int64_ - Multi-month duration in months reported by Twitch
- `plan` _string_ - Subscription plan
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `subgift`

The user gifted the subscription to a specific user. (This event **DOES** occur multiple times after `submysterygift` events!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `from` _string_ - Login name of the user who gifted the subscription
- `gifted_months` _int64_ - Number of months the user gifted
- `multi_month` _int64_ - Multi-month duration in months reported by Twitch
- `origin_id` _string_ - ID unique to the gift event
- `plan` _string_ - Subscription plan
- `subscribed_months` _int64_ - Number of months the recipient has been subscribed
- `to` _string_ - Login name of the user who received the subscription
- `total_gifted` _int64_ - Total number of subscriptions gifted by the user
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `submysterygift`

The user gifted multiple subs to the community. (This event is followed by `number x subgift` events.)

Fields:


- `channel` _string_ - Channel the event occurred in
- `from` _string_ - Login name of the user who gifted the subscriptions
- `multi_month` _int64_ - Multi-month duration in months reported by Twitch
- `number` _int64_ - Number of gifted subscriptions
- `origin_id` _string_ - ID unique to the gift event
- `plan` _string_ - Subscription plan
- `total_gifted` _int64_ - Total number of subscriptions gifted by the user
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `sus_user_message`

A suspicious (monitored / restricted) user sent a message in the given channel

Fields:


- `ban_evasion` _string_ - Ban-evasion evaluation
- `channel` _string_ - Channel the event occurred in
- `message` _string_ - Message text
- `shared_ban_channels` _[]string_ - IDs of shared-ban channels
- `status` _string_ - Restriction status
- `user_id` _string_ - ID of the suspicious user
- `user_type` _[]string_ - Suspicious-user types
- `username` _string_ - Login name of the suspicious user

## `sus_user_update`

The status of suspicious user was changed by a moderator

Fields:


- `channel` _string_ - Channel the event occurred in
- `moderator` _string_ - Login name of the acting moderator
- `status` _string_ - Restriction status
- `user_id` _string_ - ID of the suspicious user
- `username` _string_ - Login name of the suspicious user

## `timeout`

Moderator action caused a user to be timed out from chat.

Note: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable.

Fields:


- `channel` _string_ - Channel the event occurred in
- `duration` _time.Duration_ - Timeout duration in nanoseconds
- `seconds` _int_ - Timeout duration in seconds
- `target_id` _string_ - ID of the user being timed out
- `target_name` _string_ - Login name of the user being timed out

## `title_update`

The current title for the channel was changed. (This event has some delay to the real category change!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `title` _string_ - New stream title

## `vip_add`

A user was added as a VIP to the channel. (Only available when EventSub support is available and the streamer granted the required permission!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `vip_remove`

A user was removed as a VIP from the channel. (Only available when EventSub support is available and the streamer granted the required permission!)

Fields:


- `channel` _string_ - Channel the event occurred in
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `watch_streak`

The user shared a watch-streak milestone.

Fields:


- `channel` _string_ - Channel the event occurred in
- `message` _string_ - Message shared with the milestone
- `streak` _int64_ - Watch-streak value
- `user` _string_ - Login name of the user associated with the event
- `user_id` _string_ _(optional)_ - ID of the user associated with the event

## `whisper`

The bot received a whisper message. (You can use `(.*)` as message match and `{{ group 1 }}` as template to get the content of the whisper.)

This event has no event fields.
