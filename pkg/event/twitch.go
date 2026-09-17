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

// Event implements Event interface
func (AdBreakBegin) Event() *string { return new("adbreak_begin") }

// Event implements Event interface
func (CategoryUpdate) Event() *string { return new("category_update") }

// Event implements Event interface
func (ChannelPointRedeem) Event() *string { return new("channelpoint_redeem") }

// Event implements Event interface
func (Follow) Event() *string { return new("follow") }

// Event implements Event interface
func (HypetrainBegin) Event() *string { return new("hypetrain_begin") }

// Event implements Event interface
func (HypetrainEnd) Event() *string { return new("hypetrain_end") }

// Event implements Event interface
func (HypetrainProgress) Event() *string { return new("hypetrain_progress") }

// Event implements Event interface
func (ModeratorAdd) Event() *string { return new("moderator_add") }

// Event implements Event interface
func (ModeratorRemove) Event() *string { return new("moderator_remove") }

// Event implements Event interface
func (OutboundRaid) Event() *string { return new("outbound_raid") }

// Event implements Event interface
func (PollBegin) Event() *string { return new("poll_begin") }

// Event implements Event interface
func (PollEnd) Event() *string { return new("poll_end") }

// Event implements Event interface
func (PollProgress) Event() *string { return new("poll_progress") }

// Event implements Event interface
func (ShoutoutCreated) Event() *string { return new("shoutout_created") }

// Event implements Event interface
func (ShoutoutReceived) Event() *string { return new("shoutout_received") }

// Event implements Event interface
func (StreamOffline) Event() *string { return new("stream_offline") }

// Event implements Event interface
func (StreamOnline) Event() *string { return new("stream_online") }

// Event implements Event interface
func (SuspiciousUserMessage) Event() *string { return new("sus_user_message") }

// Event implements Event interface
func (SuspiciousUserUpdate) Event() *string { return new("sus_user_update") }

// Event implements Event interface
func (TitleUpdate) Event() *string { return new("title_update") }

// Event implements Event interface
func (VIPAdd) Event() *string { return new("vip_add") }

// Event implements Event interface
func (VIPRemove) Event() *string { return new("vip_remove") }
