package access

import (
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

const (
	coreMetaKeyBotToken        = "bot_access_token"
	coreMetaKeyBotUsername     = "bot_username"
	coreMetaKeyBotRefreshToken = "bot_refresh_token" //#nosec:G101 // That's a key, not a credential
)

type (
	ClientConfig struct {
		TwitchClient       string
		TwitchClientSecret string

		FallbackToken string

		TokenUpdateHook func()
	}

	extendedPermission struct {
		Channel      string `gorm:"primaryKey"`
		AccessToken  string
		RefreshToken string
		Scopes       string
	}

	Service struct{ db database.Connector }
)

func New(db database.Connector) (*Service, error) {
	return &Service{db}, errors.Wrap(
		db.DB().AutoMigrate(&extendedPermission{}),
		"migrating database schema",
	)
}

func (s Service) GetBotUsername() (string, error) {
	var botUsername string
	err := s.db.ReadCoreMeta(coreMetaKeyBotUsername, &botUsername)
	return botUsername, errors.Wrap(err, "reading bot username")
}

func (s Service) GetChannelPermissions(channel string) ([]string, error) {
	var (
		err  error
		perm extendedPermission
	)

	if err = s.db.DB().First(&perm, "channel = ?", channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "getting twitch credential from database")
	}

	return strings.Split(perm.Scopes, " "), nil
}

func (s Service) GetBotTwitchClient(cfg ClientConfig) (*twitch.Client, error) {
	botUsername, err := s.GetBotUsername()
	switch {
	case errors.Is(err, nil):
		// This is fine, we have a username
		return s.GetTwitchClientForChannel(botUsername, cfg)

	case errors.Is(err, database.ErrCoreMetaNotFound):
		// The bot has no username stored, we try to auto-migrate below

	default:
		return nil, errors.Wrap(err, "getting bot username from database")
	}

	// Bot username is not set, either we're running from fallback token
	// or did not yet execute the v3.5.0 migration

	var botAccessToken, botRefreshToken string
	err = s.db.ReadEncryptedCoreMeta(coreMetaKeyBotToken, &botAccessToken)
	switch {
	case errors.Is(err, nil):
		// This is fine, we do have a pre-v3.5.0 config, lets do the migration

	case errors.Is(err, database.ErrCoreMetaNotFound):
		// We're don't have a stored pre-v3.5.0 token either, so we're
		// running from the fallback token (which might be empty)
		return twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, cfg.FallbackToken, ""), nil

	default:
		return nil, errors.Wrap(err, "getting bot access token from database")
	}

	if err = s.db.ReadEncryptedCoreMeta(coreMetaKeyBotRefreshToken, &botRefreshToken); err != nil {
		return nil, errors.Wrap(err, "getting bot refresh token from database")
	}

	// Now we do have (hopefully valid) tokens for the bot and therefore
	// can determine who the bot is. That means we can set the username
	// for later reference and afterwards delete the duplicated tokens.

	_, botUser, err := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, botAccessToken, botRefreshToken).GetAuthorizedUser()
	if err != nil {
		return nil, errors.Wrap(err, "validating stored access token")
	}

	if err = s.db.StoreCoreMeta(coreMetaKeyBotUsername, botUser); err != nil {
		return nil, errors.Wrap(err, "setting bot username")
	}

	if _, err = s.GetTwitchClientForChannel(botUser, cfg); errors.Is(err, gorm.ErrRecordNotFound) {
		// There is no extended permission for that channel, we probably
		// are in a state created by the v2 migrator. Lets just store the
		// token without any permissions as we cannot know the permissions
		// assigned to that token
		if err = s.SetExtendedTwitchCredentials(botUser, botAccessToken, botRefreshToken, nil); err != nil {
			return nil, errors.Wrap(err, "moving bot access token")
		}
	}

	if err = s.db.DeleteCoreMeta(coreMetaKeyBotToken); err != nil {
		return nil, errors.Wrap(err, "deleting deprecated bot token")
	}

	if err = s.db.DeleteCoreMeta(coreMetaKeyBotRefreshToken); err != nil {
		return nil, errors.Wrap(err, "deleting deprecated bot refresh-token")
	}

	return s.GetTwitchClientForChannel(botUser, cfg)
}

func (s Service) GetTwitchClientForChannel(channel string, cfg ClientConfig) (*twitch.Client, error) {
	var (
		err  error
		perm extendedPermission
	)

	if err = s.db.DB().First(&perm, "channel = ?", channel).Error; err != nil {
		return nil, errors.Wrap(err, "getting twitch credential from database")
	}

	if perm.AccessToken, err = s.db.DecryptField(perm.AccessToken); err != nil {
		return nil, errors.Wrap(err, "decrypting access token")
	}

	if perm.RefreshToken, err = s.db.DecryptField(perm.RefreshToken); err != nil {
		return nil, errors.Wrap(err, "decrypting refresh token")
	}

	scopes := strings.Split(perm.Scopes, " ")

	tc := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, perm.AccessToken, perm.RefreshToken)
	tc.SetTokenUpdateHook(func(at, rt string) error {
		return errors.Wrap(s.SetExtendedTwitchCredentials(channel, at, rt, scopes), "updating extended permissions token")
	})

	return tc, nil
}

func (s Service) HasAnyPermissionForChannel(channel string, scopes ...string) (bool, error) {
	storedScopes, err := s.GetChannelPermissions(channel)
	if err != nil {
		return false, errors.Wrap(err, "getting channel scopes")
	}

	for _, scope := range scopes {
		if str.StringInSlice(scope, storedScopes) {
			return true, nil
		}
	}

	return false, nil
}

func (s Service) HasPermissionsForChannel(channel string, scopes ...string) (bool, error) {
	storedScopes, err := s.GetChannelPermissions(channel)
	if err != nil {
		return false, errors.Wrap(err, "getting channel scopes")
	}

	for _, scope := range scopes {
		if !str.StringInSlice(scope, storedScopes) {
			return false, nil
		}
	}

	return true, nil
}

func (s Service) ListPermittedChannels() ([]string, error) {
	var perms []extendedPermission
	if err := s.db.DB().Find(&perms).Error; err != nil {
		return nil, errors.Wrap(err, "listing permissions")
	}

	var out []string
	for _, perm := range perms {
		out = append(out, perm.Channel)
	}

	return out, nil
}

func (s Service) RemoveExendedTwitchCredentials(channel string) error {
	return errors.Wrap(
		s.db.DB().Delete(&extendedPermission{}, "channel = ?", channel).Error,
		"deleting data from table",
	)
}

// Deprecated: Use SetBotUsername and SetExtendedTwitchCredentials
// instead. This function is only required for the v2 migration tool.
func (s Service) SetBotTwitchCredentials(accessToken, refreshToken string) (err error) {
	if err = s.db.StoreEncryptedCoreMeta(coreMetaKeyBotToken, accessToken); err != nil {
		return errors.Wrap(err, "storing bot access token")
	}

	if err = s.db.StoreEncryptedCoreMeta(coreMetaKeyBotRefreshToken, refreshToken); err != nil {
		return errors.Wrap(err, "storing bot refresh token")
	}

	return nil
}

func (s Service) SetBotUsername(channel string) (err error) {
	return errors.Wrap(
		s.db.StoreCoreMeta(coreMetaKeyBotUsername, channel),
		"storing bot username",
	)
}

func (s Service) SetExtendedTwitchCredentials(channel, accessToken, refreshToken string, scope []string) (err error) {
	if accessToken, err = s.db.EncryptField(accessToken); err != nil {
		return errors.Wrap(err, "encrypting access token")
	}

	if refreshToken, err = s.db.EncryptField(refreshToken); err != nil {
		return errors.Wrap(err, "encrypting refresh token")
	}

	return errors.Wrap(
		s.db.DB().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "channel"}},
			DoUpdates: clause.AssignmentColumns([]string{"access_token", "refresh_token", "scopes"}),
		}).Create(extendedPermission{
			Channel:      channel,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Scopes:       strings.Join(scope, " "),
		}).Error,
		"inserting data into table",
	)
}
