package main

import "github.com/Luzifer/twitch-bot/v3/pkg/twitch"

var (
	channelExtendedScopes = map[string]string{
		twitch.ScopeChannelEditCommercial:        "run commercial",
		twitch.ScopeChannelManageBroadcast:       "modify category / title, create markers",
		twitch.ScopeChannelManagePolls:           "manage polls",
		twitch.ScopeChannelManagePredictions:     "manage predictions",
		twitch.ScopeChannelManageRaids:           "start raids",
		twitch.ScopeChannelManageVIPS:            "manage VIPs",
		twitch.ScopeChannelReadAds:               "see when an ad-break starts",
		twitch.ScopeChannelReadHypetrain:         "see Hype-Train events",
		twitch.ScopeChannelReadRedemptions:       "see channel-point redemptions",
		twitch.ScopeChannelReadSubscriptions:     "see subscribed users / sub count / points",
		twitch.ScopeClipsEdit:                    "create clips on behalf of this user",
		twitch.ScopeModeratorReadFollowers:       "see who follows this channel",
		twitch.ScopeModeratorReadShoutouts:       "see shoutouts created / received",
		twitch.ScopeModeratorReadSuspiciousUsers: "see users marked suspicious / restricted",
		twitch.ScopeUserManageWhispers:           "send whispers on behalf of this user",
	}

	botDefaultScopes = []string{
		// API Scopes
		twitch.ScopeModeratorManageAnnoucements,
		twitch.ScopeModeratorManageBannedUsers,
		twitch.ScopeModeratorManageChatMessages,
		twitch.ScopeModeratorManageChatSettings,
		twitch.ScopeModeratorManageShieldMode,
		twitch.ScopeModeratorManageShoutouts,
		twitch.ScopeModeratorReadFollowers,

		// Chat Scopes
		twitch.ScopeChatEdit,
		twitch.ScopeChatRead,
		twitch.ScopeWhisperRead,
	}
)
