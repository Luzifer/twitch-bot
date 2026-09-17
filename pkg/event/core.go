package event

import "time"

type (
	// Announcement represents an announcement sent in chat
	Announcement struct {
		BaseChannel
		BaseUser
		Color   string `field:"color" description:"Announcement color"`
		Message string `field:"message" description:"Announcement text"`
	}

	// BaseChannel is a mixin to provide the definition for the Channel field
	BaseChannel struct {
		Channel string `field:"channel" description:"Channel the event occurred in"`
	}

	// BaseUser is a mixin to provide the definitions of user related fields
	BaseUser struct {
		User   string `field:"user" description:"Login name of the user associated with the event"`
		UserID string `field:"user_id,omitzero" description:"ID of the user associated with the event"`
	}

	// Ban represents a user being banned from chat
	Ban struct {
		BaseChannel
		TargetID   string `field:"target_id" description:"ID of the user being banned"`
		TargetName string `field:"target_name" description:"Login name of the user being banned"`
	}

	// Bits represents a user spending bits in a chat message
	Bits struct {
		BaseChannel
		BaseUser
		Bits    int64  `field:"bits" description:"Total amount of bits spent in the message"`
		Message string `field:"message" description:"Chat message containing the bits"`
	}

	// ClearChat represents a channel chat being cleared
	ClearChat struct {
		BaseChannel
	}

	// ClearMessage represents a message deletion event
	ClearMessage struct {
		BaseChannel
		MessageID  string `field:"message_id" description:"UUID of the message being deleted"`
		TargetName string `field:"target_name" description:"Login name of the author of the deleted message"`
	}

	// GiftPaidUpgrade represents a user upgrading a gifted subscription to a paid subscription
	GiftPaidUpgrade struct {
		BaseChannel
		BaseUser
		Gifter string `field:"gifter" description:"Login name of the user who gifted the subscription"`
	}

	// Join represents a user joining a channel chat
	Join struct {
		BaseChannel
		BaseUser
	}

	// Message represents a normal chat message
	Message struct{}

	// Part represents a user leaving a channel chat
	Part struct {
		BaseChannel
		BaseUser
	}

	// Permit represents a user receiving a permit
	Permit struct {
		BaseChannel
		BaseUser
		To string `field:"to" description:"Login name of the user who received the permit"`
	}

	// PrimePaidUpgrade represents a user upgrading a Prime subscription to a paid subscription
	PrimePaidUpgrade struct {
		BaseChannel
		BaseUser
		Plan string `field:"plan" description:"Paid subscription plan"`
	}

	// Raid represents an incoming raid
	Raid struct {
		BaseChannel
		BaseUser
		From        string `field:"from" description:"Login name of the user who raided the channel"`
		ViewerCount int64  `field:"viewercount" description:"Number of viewers included in the raid"`
	}

	// Resub represents a user sharing their resubscription
	Resub struct {
		BaseChannel
		BaseUser
		From             string `field:"from" description:"Login name of the user who resubscribed"`
		Message          string `field:"message" description:"Message shared with the resubscription"`
		MultiMonth       int64  `field:"multi_month" description:"Multi-month duration in months reported by Twitch"`
		Plan             string `field:"plan" description:"Subscription plan"`
		SubscribedMonths int64  `field:"subscribed_months" description:"Number of months the user has been subscribed"`
	}

	// Sub represents a new subscription
	Sub struct {
		BaseChannel
		BaseUser
		From       string `field:"from" description:"Login name of the user who subscribed"`
		MultiMonth int64  `field:"multi_month" description:"Multi-month duration in months reported by Twitch"`
		Plan       string `field:"plan" description:"Subscription plan"`
	}

	// SubGift represents a subscription gifted to a specific user
	SubGift struct {
		BaseChannel
		BaseUser
		From             string `field:"from" description:"Login name of the user who gifted the subscription"`
		GiftedMonths     int64  `field:"gifted_months" description:"Number of months the user gifted"`
		MultiMonth       int64  `field:"multi_month" description:"Multi-month duration in months reported by Twitch"`
		OriginID         string `field:"origin_id" description:"ID unique to the gift event"`
		Plan             string `field:"plan" description:"Subscription plan"`
		SubscribedMonths int64  `field:"subscribed_months" description:"Number of months the recipient has been subscribed"`
		To               string `field:"to" description:"Login name of the user who received the subscription"`
		TotalGifted      int64  `field:"total_gifted" description:"Total number of subscriptions gifted by the user"`
	}

	// SubMysteryGift represents multiple subscriptions gifted to the community
	SubMysteryGift struct {
		BaseChannel
		BaseUser
		From        string `field:"from" description:"Login name of the user who gifted the subscriptions"`
		MultiMonth  int64  `field:"multi_month" description:"Multi-month duration in months reported by Twitch"`
		Number      int64  `field:"number" description:"Number of gifted subscriptions"`
		OriginID    string `field:"origin_id" description:"ID unique to the gift event"`
		Plan        string `field:"plan" description:"Subscription plan"`
		TotalGifted int64  `field:"total_gifted" description:"Total number of subscriptions gifted by the user"`
	}

	// Timeout represents a user being timed out from chat
	Timeout struct {
		BaseChannel
		Duration   time.Duration `field:"duration" description:"Timeout duration in nanoseconds"`
		Seconds    int           `field:"seconds" description:"Timeout duration in seconds"`
		TargetID   string        `field:"target_id" description:"ID of the user being timed out"`
		TargetName string        `field:"target_name" description:"Login name of the user being timed out"`
	}

	// WatchStreak represents a user sharing a watch-streak milestone
	WatchStreak struct {
		BaseChannel
		BaseUser
		Message string `field:"message" description:"Message shared with the milestone"`
		Streak  int64  `field:"streak" description:"Watch-streak value"`
	}

	// Whisper represents an incoming whisper message
	Whisper struct{}
)

// Description implements DocumentedEvent interface
func (Announcement) Description() string { return "An announcement was sent in chat." }

// Event implements Event interface
func (Announcement) Event() *string { return new("announcement") }

// Description implements DocumentedEvent interface
func (Ban) Description() string {
	return "Moderator action caused a user to be banned from chat.\n\nNote: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable."
}

// Event implements Event interface
func (Ban) Event() *string { return new("ban") }

// Description implements DocumentedEvent interface
func (Bits) Description() string {
	return "User spent bits in the channel. The full message is available like in a normal chat message, additionally the `{{ .bits }}` field is added with the total amount of bits spent."
}

// Event implements Event interface
func (Bits) Event() *string { return new("bits") }

// Description implements DocumentedEvent interface
func (ClearChat) Description() string {
	return "Moderator action caused chat to be cleared.\n\nNote: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable."
}

// Event implements Event interface
func (ClearChat) Event() *string { return new("clearchat") }

// Description implements DocumentedEvent interface
func (ClearMessage) Description() string {
	return "Moderator action caused a chat message to be deleted.\n\nNote: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable."
}

// Event implements Event interface
func (ClearMessage) Event() *string { return new("delete") }

// Description implements DocumentedEvent interface
func (GiftPaidUpgrade) Description() string {
	return "User upgraded their gifted subscription into a paid one. This event does not contain any details about the tier of the paid subscription."
}

// Event implements Event interface
func (GiftPaidUpgrade) Event() *string { return new("giftpaidupgrade") }

// Description implements DocumentedEvent interface
func (Join) Description() string {
	return "User joined the channel-chat. This is **NOT** an indicator they are viewing, the event is **NOT** reliably sent when the user really joined the chat. The event will be sent with some delay after they join the chat and is sometimes repeated multiple times during their stay. So **DO NOT** use this to greet users!"
}

// Event implements Event interface
func (Join) Event() *string { return new("join") }

// Event implements Event interface
func (Message) Event() *string { return nil }

// Description implements DocumentedEvent interface
func (Part) Description() string {
	return "User left the channel-chat. This is **NOT** an indicator they are no longer viewing, the event is **NOT** reliably sent when the user really leaves the chat. The event will be sent with some delay after they leave the chat and is sometimes repeated multiple times during their stay. So this does **NOT** mean they do no longer read the chat!"
}

// Event implements Event interface
func (Part) Event() *string { return new("part") }

// Description implements DocumentedEvent interface
func (Permit) Description() string {
	return "User received a permit, which means they are no longer affected by rules which are disabled on permit."
}

// Event implements Event interface
func (Permit) Event() *string { return new("permit") }

// Description implements DocumentedEvent interface
func (PrimePaidUpgrade) Description() string {
	return "User upgraded their Prime subscription into a paid one."
}

// Event implements Event interface
func (PrimePaidUpgrade) Event() *string { return new("primepaidupgrade") }

// Description implements DocumentedEvent interface
func (Raid) Description() string { return "The channel was raided by another user." }

// Event implements Event interface
func (Raid) Event() *string { return new("raid") }

// Description implements DocumentedEvent interface
func (Resub) Description() string {
	return "The user shared their resubscription. (This event is triggered manually by the user using the \"Share my Resub\" button and does not occur when the user does not actively share their sub!)"
}

// Event implements Event interface
func (Resub) Event() *string { return new("resub") }

// Description implements DocumentedEvent interface
func (Sub) Description() string {
	return "The user newly subscribed on their own. (This event is triggered automatically and does not need to be shared actively!)"
}

// Event implements Event interface
func (Sub) Event() *string { return new("sub") }

// Description implements DocumentedEvent interface
func (SubGift) Description() string {
	return "The user gifted the subscription to a specific user. (This event **DOES** occur multiple times after `submysterygift` events!)"
}

// Event implements Event interface
func (SubGift) Event() *string { return new("subgift") }

// Description implements DocumentedEvent interface
func (SubMysteryGift) Description() string {
	return "The user gifted multiple subs to the community. (This event is followed by `number x subgift` events.)"
}

// Event implements Event interface
func (SubMysteryGift) Event() *string { return new("submysterygift") }

// Description implements DocumentedEvent interface
func (Timeout) Description() string {
	return "Moderator action caused a user to be timed out from chat.\n\nNote: This event does **not** contain the acting user! You cannot use the `{{.user}}` variable."
}

// Event implements Event interface
func (Timeout) Event() *string { return new("timeout") }

// Description implements DocumentedEvent interface
func (WatchStreak) Description() string { return "The user shared a watch-streak milestone." }

// Event implements Event interface
func (WatchStreak) Event() *string { return new("watch_streak") }

// Description implements DocumentedEvent interface
func (Whisper) Description() string {
	return "The bot received a whisper message. (You can use `(.*)` as message match and `{{ group 1 }}` as template to get the content of the whisper.)"
}

// Event implements Event interface
func (Whisper) Event() *string { return new("whisper") }
