// Package access contains a service to manage Twitch tokens and scopes
package access

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Luzifer/go_helpers/backoff"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

const (
	coreMetaKeyBotUsername = "bot_username"
)

type (
	// ClientConfig contains a configuration to derive new Twitch clients
	// from
	ClientConfig struct {
		TwitchClient       string
		TwitchClientSecret string

		TokenUpdateHook func()
	}

	extendedPermission struct {
		Channel      string `gorm:"primaryKey"`
		AccessToken  string //#nosec:G117 // Intended to handle secrets
		RefreshToken string //#nosec:G117 // Intended to handle secrets
		Scopes       string
	}

	// Service manages the permission database
	Service struct{ db database.Connector }
)

// ErrChannelNotAuthorized denotes there is no valid authoriztion for
// the given channel
var ErrChannelNotAuthorized = errors.New("channel is not authorized")

// New creates a new Service on the given database
func New(db database.Connector) (*Service, error) {
	if err := db.DB().AutoMigrate(&extendedPermission{}); err != nil {
		return nil, fmt.Errorf("migrating database schema: %w", err)
	}

	return &Service{db}, nil
}

// CopyDatabase enables the bot to migrate the access database
func (*Service) CopyDatabase(src, target *gorm.DB) error {
	return database.CopyObjects(src, target, &extendedPermission{}) //nolint:wrapcheck // Internal helper
}

// GetBotTwitchClient returns a twitch.Client configured to act as the
// bot user
func (s Service) GetBotTwitchClient(cfg ClientConfig) (*twitch.Client, error) {
	botUsername, err := s.GetBotUsername()
	if err != nil {
		return nil, fmt.Errorf("getting bot username: %w", err)
	}

	return s.GetTwitchClientForChannel(botUsername, cfg)
}

// GetBotUsername gets the cached bot username
func (s Service) GetBotUsername() (botUsername string, err error) {
	if err = s.db.ReadCoreMeta(coreMetaKeyBotUsername, &botUsername); err != nil {
		return botUsername, fmt.Errorf("reading bot username: %w", err)
	}
	return botUsername, nil
}

// GetChannelPermissions returns the scopes granted for the given channel
func (s Service) GetChannelPermissions(channel string) ([]string, error) {
	var (
		err  error
		perm extendedPermission
	)

	if err = helpers.Retry(func() error {
		if err = s.db.DB().First(&perm, "channel = ?", strings.TrimLeft(channel, "#")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}

			return fmt.Errorf("getting twitch credential from database: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return strings.Split(perm.Scopes, " "), nil
}

// GetTwitchClientForChannel returns a twitch.Client configured to act
// as the owner of the given channel
func (s Service) GetTwitchClientForChannel(channel string, cfg ClientConfig) (*twitch.Client, error) {
	var (
		err  error
		perm extendedPermission
	)

	if err = helpers.Retry(func() error {
		if err = s.db.DB().First(&perm, "channel = ?", strings.TrimLeft(channel, "#")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return backoff.NewErrCannotRetry(ErrChannelNotAuthorized)
			}
			return fmt.Errorf("getting twitch credential from database: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if perm.AccessToken, err = s.db.DecryptField(perm.AccessToken); err != nil {
		return nil, fmt.Errorf("decrypting access token: %w", err)
	}

	if perm.RefreshToken, err = s.db.DecryptField(perm.RefreshToken); err != nil {
		return nil, fmt.Errorf("decrypting refresh token: %w", err)
	}

	if perm.AccessToken == "" && perm.RefreshToken == "" {
		// We have no tokens but an entry in the permission table: Means
		// we still can't do stuff on behalf of that channel so we treat
		// that as an unauthorized channel
		return nil, ErrChannelNotAuthorized
	}

	scopes := strings.Split(perm.Scopes, " ")

	tc := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, perm.AccessToken, perm.RefreshToken)
	tc.SetTokenUpdateHook(func(at, rt string) error {
		if err := s.SetExtendedTwitchCredentials(channel, at, rt, scopes); err != nil {
			return fmt.Errorf("updating extended permissions token: %w", err)
		}

		return nil
	})

	return tc, nil
}

// HasAnyPermissionForChannel checks whether any of the given scopes
// are granted for the given channel
func (s Service) HasAnyPermissionForChannel(channel string, scopes ...string) (bool, error) {
	storedScopes, err := s.GetChannelPermissions(channel)
	if err != nil {
		return false, fmt.Errorf("getting channel scopes: %w", err)
	}

	for _, scope := range scopes {
		if slices.Contains(storedScopes, scope) {
			return true, nil
		}
	}

	return false, nil
}

// HasPermissionsForChannel checks whether all of the given scopes
// are granted for the given channel
func (s Service) HasPermissionsForChannel(channel string, scopes ...string) (bool, error) {
	storedScopes, err := s.GetChannelPermissions(channel)
	if err != nil {
		return false, fmt.Errorf("getting channel scopes: %w", err)
	}

	for _, scope := range scopes {
		if !slices.Contains(storedScopes, scope) {
			return false, nil
		}
	}

	return true, nil
}

// HasTokensForChannel retrieves and decrypts stored access- and
// refresh-tokens to evaluate whether tokens are available. Those
// tokens are NOT validated in this request, it's just checked whether
// they are present
func (s Service) HasTokensForChannel(channel string) (bool, error) {
	var (
		err  error
		perm extendedPermission
	)

	if err = helpers.Retry(func() error {
		if err = s.db.DB().First(&perm, "channel = ?", strings.TrimLeft(channel, "#")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return backoff.NewErrCannotRetry(ErrChannelNotAuthorized)
			}
			return fmt.Errorf("getting twitch credential from database: %w", err)
		}

		return nil
	}); err != nil {
		if errors.Is(err, ErrChannelNotAuthorized) {
			return false, nil
		}
		return false, err
	}

	if perm.AccessToken, err = s.db.DecryptField(perm.AccessToken); err != nil {
		return false, fmt.Errorf("decrypting access token: %w", err)
	}

	if perm.RefreshToken, err = s.db.DecryptField(perm.RefreshToken); err != nil {
		return false, fmt.Errorf("decrypting refresh token: %w", err)
	}

	return perm.AccessToken != "" && perm.RefreshToken != "", nil
}

// ListPermittedChannels returns a list of all channels having a token
// for the channels owner
func (s Service) ListPermittedChannels() (out []string, err error) {
	var perms []extendedPermission
	if err = helpers.Retry(func() error {
		if err := s.db.DB().Find(&perms).Error; err != nil {
			return fmt.Errorf("listing permissions: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err //nolint:wrapcheck // is already wrapped on the inside
	}

	for _, perm := range perms {
		out = append(out, perm.Channel)
	}

	return out, nil
}

// RemoveAllExtendedTwitchCredentials wipes the access database
func (s Service) RemoveAllExtendedTwitchCredentials() error {
	if err := helpers.RetryTransaction(s.db.DB(), func(tx *gorm.DB) error {
		return tx.Delete(&extendedPermission{}, "1 = 1").Error
	}); err != nil {
		return fmt.Errorf("deleting data from table: %w", err)
	}

	return nil
}

// RemoveExendedTwitchCredentials wipes the access database for a given
// channel
func (s Service) RemoveExendedTwitchCredentials(channel string) error {
	if err := helpers.RetryTransaction(s.db.DB(), func(tx *gorm.DB) error {
		return tx.Delete(&extendedPermission{}, "channel = ?", strings.TrimLeft(channel, "#")).Error
	}); err != nil {
		return fmt.Errorf("deleting data from table: %w", err)
	}

	return nil
}

// SetBotUsername stores the username of the bot
func (s Service) SetBotUsername(channel string) (err error) {
	if err = s.db.StoreCoreMeta(coreMetaKeyBotUsername, strings.TrimLeft(channel, "#")); err != nil {
		return fmt.Errorf("storing bot username: %w", err)
	}

	return nil
}

// SetExtendedTwitchCredentials stores tokens and scopes for the given
// channel into the access database
func (s Service) SetExtendedTwitchCredentials(channel, accessToken, refreshToken string, scope []string) (err error) {
	if accessToken, err = s.db.EncryptField(accessToken); err != nil {
		return fmt.Errorf("encrypting access token: %w", err)
	}

	if refreshToken, err = s.db.EncryptField(refreshToken); err != nil {
		return fmt.Errorf("encrypting refresh token: %w", err)
	}

	if err = helpers.RetryTransaction(s.db.DB(), func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "channel"}},
			DoUpdates: clause.AssignmentColumns([]string{"access_token", "refresh_token", "scopes"}),
		}).Create(extendedPermission{
			Channel:      strings.TrimLeft(channel, "#"),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Scopes:       strings.Join(scope, " "),
		}).Error
	}); err != nil {
		return fmt.Errorf("inserting data into table: %w", err)
	}

	return nil
}
