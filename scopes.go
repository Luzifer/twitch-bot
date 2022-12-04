package main

import "github.com/Luzifer/twitch-bot/v3/pkg/twitch"

var (
	channelExtendedScopes = map[string]string{
		twitch.ScopeChannelEditCommercial:    "run commercial",
		twitch.ScopeChannelManageBroadcast:   "modify category / title",
		twitch.ScopeChannelManagePolls:       "manage polls",
		twitch.ScopeChannelManagePredictions: "manage predictions",
		twitch.ScopeChannelManageRaids:       "start raids",
		twitch.ScopeChannelManageVIPS:        "manage VIPs",
		twitch.ScopeChannelReadRedemptions:   "see channel-point redemptions",
	}

	botDefaultScopes = []string{
		// API Scopes
		twitch.ScopeModeratorManageAnnoucements,
		twitch.ScopeModeratorManageBannedUsers,
		twitch.ScopeModeratorManageChatMessages,
		twitch.ScopeModeratorManageChatSettings,

		// Chat Scopes
		twitch.ScopeChatEdit,
		twitch.ScopeChatRead,
		twitch.ScopeWhisperRead,
	}
)
