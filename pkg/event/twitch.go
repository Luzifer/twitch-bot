package event

import (
	"time"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

type (
	// AdBreakBegin represents the start of an ad break
	AdBreakBegin struct {
		BaseChannel
		Duration    int64     `field:"duration" description:"Duration of the ad break in seconds"`
		IsAutomatic bool      `field:"is_automatic" description:"Whether the ad break was started automatically"`
		StartedAt   time.Time `field:"started_at" description:"Time the ad break started"`
	}

	// CategoryUpdate represents a stream-category change
	CategoryUpdate struct {
		BaseChannel
		Category string `field:"category" description:"New stream category"`
	}

	// ChannelPointRedeem represents a custom channel-point reward redemption
	ChannelPointRedeem struct {
		BaseChannel
		BaseUser
		RewardCost  int64  `field:"reward_cost" description:"Number of points paid for the reward"`
		RewardID    string `field:"reward_id" description:"ID of the redeemed reward"`
		RewardTitle string `field:"reward_title" description:"Title of the redeemed reward"`
		Status      string `field:"status" description:"Status of the redemption"`
		UserInput   string `field:"user_input" description:"Text entered for the reward"`
	}

	// Follow represents a user following a channel
	Follow struct {
		BaseChannel
		BaseUser
		FollowedAt time.Time `field:"followed_at" description:"Time the user followed the channel"`
	}

	// Hypetrain contains the common fields of all Hype-Train events
	Hypetrain struct {
		BaseChannel
		Event twitch.EventSubEventHypetrain `field:"event" nolog:"true" description:"Raw Hype-Train event"`
		Level int64                         `field:"level" description:"Current Hype-Train level"`
	}

	// HypetrainBegin represents the beginning of a Hype-Train
	HypetrainBegin struct {
		Hypetrain
		HypetrainLevelProgress
	}
	// HypetrainEnd represents the end of a Hype-Train
	HypetrainEnd struct{ Hypetrain }

	// HypetrainLevelProgress contains the progress fields available while a Hype-Train is active
	HypetrainLevelProgress struct {
		LevelProgress float64 `field:"levelProgress" description:"Progress towards the next Hype-Train level"`
	}
	// HypetrainProgress represents progress of a Hype-Train
	HypetrainProgress struct {
		Hypetrain
		HypetrainLevelProgress
	}

	// ModeratorAdd represents a user receiving moderator status
	ModeratorAdd struct {
		BaseChannel
		BaseUser
	}

	// ModeratorRemove represents a user losing moderator status
	ModeratorRemove struct {
		BaseChannel
		BaseUser
	}

	// OutboundRaid represents a channel raiding another channel
	OutboundRaid struct {
		BaseChannel
		To      string `field:"to" description:"Login name of the channel receiving the raid"`
		ToID    string `field:"to_id" description:"ID of the channel receiving the raid"`
		Viewers int64  `field:"viewers" description:"Number of viewers included in the raid"`
	}

	// Poll contains the common fields of all poll events
	Poll struct {
		BaseChannel
		HasChannelPointVoting bool                     `field:"hasChannelPointVoting" description:"Whether channel-point voting is enabled"`
		Poll                  twitch.EventSubEventPoll `field:"poll" nolog:"true" description:"Raw poll event"`
		Title                 string                   `field:"title" description:"Poll title"`
	}

	// PollBegin represents the beginning of a poll
	PollBegin struct{ Poll }

	// PollEnd represents the end of a poll
	PollEnd struct {
		Poll
		Status string `field:"status" description:"Poll status"`
	}

	// PollProgress represents changes to a poll
	PollProgress struct{ Poll }

	// ShoutoutCreated represents a shoutout created by the channel
	ShoutoutCreated struct {
		BaseChannel
		To      string `field:"to" description:"Login name of the channel receiving the shoutout"`
		ToID    string `field:"to_id" description:"ID of the channel receiving the shoutout"`
		Viewers int64  `field:"viewers" description:"Number of viewers shown the shoutout"`
	}

	// ShoutoutReceived represents a shoutout received by the channel
	ShoutoutReceived struct {
		BaseChannel
		From    string `field:"from" description:"Login name of the channel issuing the shoutout"`
		FromID  string `field:"from_id" description:"ID of the channel issuing the shoutout"`
		Viewers int64  `field:"viewers" description:"Number of viewers shown the shoutout"`
	}

	// StreamOffline represents the stream going offline
	StreamOffline struct{ BaseChannel }
	// StreamOnline represents the stream going online
	StreamOnline struct{ BaseChannel }

	// SuspiciousUserMessage represents a message from a suspicious user
	SuspiciousUserMessage struct {
		BaseChannel
		BanEvasion        string   `field:"ban_evasion" description:"Ban-evasion evaluation"`
		Message           string   `field:"message" description:"Message text"`
		SharedBanChannels []string `field:"shared_ban_channels" description:"IDs of shared-ban channels"`
		Status            string   `field:"status" description:"Restriction status"`
		UserID            string   `field:"user_id" description:"ID of the suspicious user"`
		UserType          []string `field:"user_type" description:"Suspicious-user types"`
		Username          string   `field:"username" description:"Login name of the suspicious user"`
	}

	// SuspiciousUserUpdate represents a change to a suspicious user
	SuspiciousUserUpdate struct {
		BaseChannel
		Moderator string `field:"moderator" description:"Login name of the acting moderator"`
		Status    string `field:"status" description:"Restriction status"`
		UserID    string `field:"user_id" description:"ID of the suspicious user"`
		Username  string `field:"username" description:"Login name of the suspicious user"`
	}

	// TitleUpdate represents a stream-title change
	TitleUpdate struct {
		BaseChannel
		Title string `field:"title" description:"New stream title"`
	}

	// VIPAdd represents a user receiving VIP status
	VIPAdd struct {
		BaseChannel
		BaseUser
	}

	// VIPRemove represents a user losing VIP status
	VIPRemove struct {
		BaseChannel
		BaseUser
	}
)

// Description implements DocumentedEvent interface
func (AdBreakBegin) Description() string {
	return "Ad-break has begun and ads are playing now in mentioned channel."
}

// Event implements Event interface
func (AdBreakBegin) Event() *string { return new("adbreak_begin") }

// Description implements DocumentedEvent interface
func (CategoryUpdate) Description() string {
	return "The current category for the channel was changed. (This event has some delay to the real category change!)"
}

// Event implements Event interface
func (CategoryUpdate) Event() *string { return new("category_update") }

// Description implements DocumentedEvent interface
func (ChannelPointRedeem) Description() string {
	return "A custom channel-point reward was redeemed in the given channel. (Only available when EventSub support is available and streamer granted required permissions!)"
}

// Event implements Event interface
func (ChannelPointRedeem) Event() *string { return new("channelpoint_redeem") }

// Description implements DocumentedEvent interface
func (Follow) Description() string {
	return "User followed the channel. This event is not de-duplicated and therefore might be used to spam! (Only available when EventSub support is available!)"
}

// Event implements Event interface
func (Follow) Event() *string { return new("follow") }

// Description implements DocumentedEvent interface
func (HypetrainBegin) Description() string {
	return "An Hype-Train has begun, ended or progressed in the given channel."
}

// Event implements Event interface
func (HypetrainBegin) Event() *string { return new("hypetrain_begin") }

// Description implements DocumentedEvent interface
func (HypetrainEnd) Description() string {
	return "An Hype-Train has begun, ended or progressed in the given channel."
}

// Event implements Event interface
func (HypetrainEnd) Event() *string { return new("hypetrain_end") }

// Description implements DocumentedEvent interface
func (HypetrainProgress) Description() string {
	return "An Hype-Train has begun, ended or progressed in the given channel."
}

// Event implements Event interface
func (HypetrainProgress) Event() *string { return new("hypetrain_progress") }

// Description implements DocumentedEvent interface
func (ModeratorAdd) Description() string {
	return "A user was added as a moderator to the channel or removed from its moderators. (Only available when EventSub support is available and the streamer granted the required permission!)"
}

// Event implements Event interface
func (ModeratorAdd) Event() *string { return new("moderator_add") }

// Description implements DocumentedEvent interface
func (ModeratorRemove) Description() string {
	return "A user was added as a moderator to the channel or removed from its moderators. (Only available when EventSub support is available and the streamer granted the required permission!)"
}

// Event implements Event interface
func (ModeratorRemove) Event() *string { return new("moderator_remove") }

// Description implements DocumentedEvent interface
func (OutboundRaid) Description() string {
	return "The channel has raided another channel. (The event is issued in the moment the raid is executed, not when the raid timer starts!)"
}

// Event implements Event interface
func (OutboundRaid) Event() *string { return new("outbound_raid") }

// Description implements DocumentedEvent interface
func (PollBegin) Description() string {
	return "A poll was started / was ended / had changes in the given channel."
}

// Event implements Event interface
func (PollBegin) Event() *string { return new("poll_begin") }

// Description implements DocumentedEvent interface
func (PollEnd) Description() string {
	return "A poll was started / was ended / had changes in the given channel."
}

// Event implements Event interface
func (PollEnd) Event() *string { return new("poll_end") }

// Description implements DocumentedEvent interface
func (PollProgress) Description() string {
	return "A poll was started / was ended / had changes in the given channel."
}

// Event implements Event interface
func (PollProgress) Event() *string { return new("poll_progress") }

// Description implements DocumentedEvent interface
func (ShoutoutCreated) Description() string {
	return "The channel gave another streamer a (Twitch native) shoutout"
}

// Event implements Event interface
func (ShoutoutCreated) Event() *string { return new("shoutout_created") }

// Description implements DocumentedEvent interface
func (ShoutoutReceived) Description() string {
	return "The channel received a (Twitch native) shoutout by another channel."
}

// Event implements Event interface
func (ShoutoutReceived) Event() *string { return new("shoutout_received") }

// Description implements DocumentedEvent interface
func (StreamOffline) Description() string {
	return "The channels stream went offline. (This event has some delay to the real button-press to \"stop stream\"!)"
}

// Event implements Event interface
func (StreamOffline) Event() *string { return new("stream_offline") }

// Description implements DocumentedEvent interface
func (StreamOnline) Description() string {
	return "The channels stream went online. (This event has some delay to the real button-press to \"start stream\"!)"
}

// Event implements Event interface
func (StreamOnline) Event() *string { return new("stream_online") }

// Description implements DocumentedEvent interface
func (SuspiciousUserMessage) Description() string {
	return "A suspicious (monitored / restricted) user sent a message in the given channel"
}

// Event implements Event interface
func (SuspiciousUserMessage) Event() *string { return new("sus_user_message") }

// Description implements DocumentedEvent interface
func (SuspiciousUserUpdate) Description() string {
	return "The status of suspicious user was changed by a moderator"
}

// Event implements Event interface
func (SuspiciousUserUpdate) Event() *string { return new("sus_user_update") }

// Description implements DocumentedEvent interface
func (TitleUpdate) Description() string {
	return "The current title for the channel was changed. (This event has some delay to the real category change!)"
}

// Event implements Event interface
func (TitleUpdate) Event() *string { return new("title_update") }

// Description implements DocumentedEvent interface
func (VIPAdd) Description() string {
	return "A user was added as a VIP to the channel or removed from its VIPs. (Only available when EventSub support is available and the streamer granted the required permission!)"
}

// Event implements Event interface
func (VIPAdd) Event() *string { return new("vip_add") }

// Description implements DocumentedEvent interface
func (VIPRemove) Description() string {
	return "A user was added as a VIP to the channel or removed from its VIPs. (Only available when EventSub support is available and the streamer granted the required permission!)"
}

// Event implements Event interface
func (VIPRemove) Event() *string { return new("vip_remove") }
