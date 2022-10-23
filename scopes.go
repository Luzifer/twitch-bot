package main

import "github.com/Luzifer/twitch-bot/v2/pkg/twitch"

var (
	channelDefaultScopes = []string{
		twitch.ScopeChannelEditCommercial,
		twitch.ScopeChannelManageBroadcast,
		twitch.ScopeChannelReadRedemptions,
		twitch.ScopeChannelManageRaids,
	}

	botDefaultScopes = append(channelDefaultScopes,
		twitch.ScopeChatEdit,
		twitch.ScopeChatRead,
		twitch.ScopeModeratorManageAnnoucements,
		twitch.ScopeModeratorManageBannedUsers,
		twitch.ScopeModeratorManageChatMessages,
		twitch.ScopeModeratorManageChatSettings,
		twitch.ScopeWhisperRead,
	)
)
