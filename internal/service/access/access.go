package access

import (
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v2/pkg/database"
	"github.com/Luzifer/twitch-bot/v2/pkg/twitch"
)

const (
	coreMetaKeyBotToken        = "bot_access_token"
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

func (s Service) GetBotTwitchClient(cfg ClientConfig) (*twitch.Client, error) {
	var botAccessToken, botRefreshToken string

	err := s.db.ReadEncryptedCoreMeta(coreMetaKeyBotToken, &botAccessToken)
	switch {
	case errors.Is(err, nil):
		// This is fine

	case errors.Is(err, database.ErrCoreMetaNotFound):
		botAccessToken = cfg.FallbackToken

	default:
		return nil, errors.Wrap(err, "getting bot access token from database")
	}

	if err = s.db.ReadEncryptedCoreMeta(coreMetaKeyBotRefreshToken, &botRefreshToken); err != nil && !errors.Is(err, database.ErrCoreMetaNotFound) {
		return nil, errors.Wrap(err, "getting bot refresh token from database")
	}

	twitchClient := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, botAccessToken, botRefreshToken)
	twitchClient.SetTokenUpdateHook(s.SetBotTwitchCredentials)

	return twitchClient, nil
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
	var (
		err  error
		perm extendedPermission
	)

	if err = s.db.DB().First(&perm, "channel = ?", channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "getting twitch credential from database")
	}

	storedScopes := strings.Split(perm.Scopes, " ")

	for _, scope := range scopes {
		if str.StringInSlice(scope, storedScopes) {
			return true, nil
		}
	}

	return false, nil
}

func (s Service) HasPermissionsForChannel(channel string, scopes ...string) (bool, error) {
	var (
		err  error
		perm extendedPermission
	)

	if err = s.db.DB().First(&perm, "channel = ?", channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "getting twitch credential from database")
	}

	storedScopes := strings.Split(perm.Scopes, " ")

	for _, scope := range scopes {
		if !str.StringInSlice(scope, storedScopes) {
			return false, nil
		}
	}

	return true, nil
}

func (s Service) RemoveExendedTwitchCredentials(channel string) error {
	return errors.Wrap(
		s.db.DB().Delete(&extendedPermission{}, "channel = ?", channel).Error,
		"deleting data from table",
	)
}

func (s Service) SetBotTwitchCredentials(accessToken, refreshToken string) (err error) {
	if err = s.db.StoreEncryptedCoreMeta(coreMetaKeyBotToken, accessToken); err != nil {
		return errors.Wrap(err, "storing bot access token")
	}

	if err = s.db.StoreEncryptedCoreMeta(coreMetaKeyBotRefreshToken, refreshToken); err != nil {
		return errors.Wrap(err, "storing bot refresh token")
	}

	return nil
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
