package twitch

const (
	// API Scopes
	ScopeChannelBot                  = "channel:bot"
	ScopeChannelEditCommercial       = "channel:edit:commercial"
	ScopeChannelManageAds            = "channel:manage:ads"
	ScopeChannelManageBroadcast      = "channel:manage:broadcast"
	ScopeChannelManageModerators     = "channel:manage:moderators"
	ScopeChannelManagePolls          = "channel:manage:polls"
	ScopeChannelManagePredictions    = "channel:manage:predictions"
	ScopeChannelManageRaids          = "channel:manage:raids"
	ScopeChannelManageRedemptions    = "channel:manage:redemptions"
	ScopeChannelManageVIPS           = "channel:manage:vips"
	ScopeChannelManageWhispers       = "user:manage:whispers"
	ScopeChannelReadAds              = "channel:read:ads"
	ScopeChannelReadPolls            = "channel:read:polls"
	ScopeChannelReadRedemptions      = "channel:read:redemptions"
	ScopeChannelReadSubscriptions    = "channel:read:subscriptions"
	ScopeClipsEdit                   = "clips:edit"
	ScopeModeratorManageAnnoucements = "moderator:manage:announcements"
	ScopeModeratorManageBannedUsers  = "moderator:manage:banned_users"
	ScopeModeratorManageChatMessages = "moderator:manage:chat_messages"
	ScopeModeratorManageChatSettings = "moderator:manage:chat_settings"
	ScopeModeratorManageShieldMode   = "moderator:manage:shield_mode"
	ScopeModeratorManageShoutouts    = "moderator:manage:shoutouts"
	ScopeModeratorReadFollowers      = "moderator:read:followers"
	ScopeModeratorReadShoutouts      = "moderator:read:shoutouts"
	ScopeUserBot                     = "user:bot"
	ScopeUserManageChatColor         = "user:manage:chat_color"
	ScopeUserManageWhispers          = "user:manage:whispers"
	ScopeUserReadChat                = "user:read:chat"

	// Deprecated v5 scope but used in chat
	ScopeV5ChannelEditor = "channel_editor"

	// Chat Scopes
	ScopeChatEdit    = "chat:edit"     // Send live stream chat and rooms messages.
	ScopeChatRead    = "chat:read"     // View live stream chat and rooms messages.
	ScopeWhisperRead = "whispers:read" // View your whisper messages.
)
