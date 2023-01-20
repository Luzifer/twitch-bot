package twitch

const (
	// API Scopes
	ScopeChannelEditCommercial       = "channel:edit:commercial"
	ScopeChannelManageBroadcast      = "channel:manage:broadcast"
	ScopeChannelManageModerators     = "channel:manage:moderators"
	ScopeChannelManagePolls          = "channel:manage:polls"
	ScopeChannelManagePredictions    = "channel:manage:predictions"
	ScopeChannelManageRaids          = "channel:manage:raids"
	ScopeChannelManageRedemptions    = "channel:manage:redemptions"
	ScopeChannelManageVIPS           = "channel:manage:vips"
	ScopeChannelManageWhispers       = "user:manage:whispers"
	ScopeChannelReadRedemptions      = "channel:read:redemptions"
	ScopeModeratorManageAnnoucements = "moderator:manage:announcements"
	ScopeModeratorManageBannedUsers  = "moderator:manage:banned_users"
	ScopeModeratorManageChatMessages = "moderator:manage:chat_messages"
	ScopeModeratorManageChatSettings = "moderator:manage:chat_settings"
	ScopeModeratorManageShoutouts    = "moderator:manage:shoutouts"
	ScopeUserManageChatColor         = "user:manage:chat_color"

	// Deprecated v5 scope but used in chat
	ScopeV5ChannelEditor = "channel_editor"

	// Chat Scopes
	ScopeChatEdit    = "chat:edit"     // Send live stream chat and rooms messages.
	ScopeChatRead    = "chat:read"     // View live stream chat and rooms messages.
	ScopeWhisperRead = "whispers:read" // View your whisper messages.
)
