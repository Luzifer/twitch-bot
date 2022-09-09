package main

import "github.com/Luzifer/twitch-bot/pkg/twitch"

var (
	channelDefaultScopes = []string{
		twitch.ScopeChannelEditCommercial,
		twitch.ScopeChannelManageBroadcast,
		twitch.ScopeChannelReadRedemptions,
	}

	botDefaultScopes = append(channelDefaultScopes,
		twitch.ScopeChatRead,
		twitch.ScopeChatEdit,
		twitch.ScopeWhisperRead,
		twitch.ScopeWhisperEdit,
		twitch.ScopeChannelModerate,
	)
)
