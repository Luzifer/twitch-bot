package twitch

const (
	// API Scopes
	ScopeChannelManageRedemptions = "channel:manage:redemptions"
	ScopeChannelReadRedemptions   = "channel:read:redemptions"
	ScopeChannelEditCommercial    = "channel:edit:commercial"
	ScopeChannelManageBroadcast   = "channel:manage:broadcast"
	ScopeChannelManagePolls       = "channel:manage:polls"
	ScopeChannelManagePredictions = "channel:manage:predictions"

	// Deprecated v5 scope but used in chat
	ScopeV5ChannelEditor = "channel_editor"

	// Chat Scopes
	ScopeChannelModerate = "channel:moderate" // Perform moderation actions in a channel. The user requesting the scope must be a moderator in the channel.
	ScopeChatEdit        = "chat:edit"        // Send live stream chat and rooms messages.
	ScopeChatRead        = "chat:read"        // View live stream chat and rooms messages.
	ScopeWhisperRead     = "whispers:read"    // View your whisper messages.
	ScopeWhisperEdit     = "whispers:edit"    // Send whisper messages.
)
