package main

import "github.com/Luzifer/twitch-bot/twitch"

var (
	botDefaultScopes = []string{
		twitch.ScopeChatRead,
		twitch.ScopeChatEdit,
		twitch.ScopeWhisperRead,
		twitch.ScopeWhisperEdit,
		twitch.ScopeChannelModerate,
		twitch.ScopeChannelManageBroadcast,
		twitch.ScopeChannelEditCommercial,
		twitch.ScopeV5ChannelEditor,
	}

	channelDefaultScopes = []string{
		twitch.ScopeChannelReadRedemptions,
	}
)
