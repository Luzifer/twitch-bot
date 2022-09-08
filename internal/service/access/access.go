package access

import (
	"database/sql"
	"strings"

	"github.com/pkg/errors"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/internal/database"
	"github.com/Luzifer/twitch-bot/twitch"
)

const (
	coreMetaKeyBotToken        = "bot_access_token"
	coreMetaKeyBotRefreshToken = "bot_refresh_token"
)

type (
	ClientConfig struct {
		TwitchClient       string
		TwitchClientSecret string

		FallbackToken string

		TokenUpdateHook func()
	}

	Store struct{ db database.Connector }
)

func New(db database.Connector) *Store {
	return &Store{db}
}

func (s Store) GetBotTwitchClient(cfg ClientConfig) (*twitch.Client, error) {
	var botAccessToken, botRefreshToken string

	err := s.db.ReadCoreMeta(coreMetaKeyBotToken, &botAccessToken)
	switch {
	case errors.Is(err, nil):
		// This is fine

	case errors.Is(err, database.ErrCoreMetaNotFound):
		botAccessToken = cfg.FallbackToken

	default:
		return nil, errors.Wrap(err, "getting bot access token from database")
	}

	if err = s.db.ReadCoreMeta(coreMetaKeyBotToken, &botAccessToken); err != nil && !errors.Is(err, database.ErrCoreMetaNotFound) {
		return nil, errors.Wrap(err, "getting bot refresh token from database")
	}

	twitchClient := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, botAccessToken, botRefreshToken)
	twitchClient.SetTokenUpdateHook(s.SetBotTwitchCredentials)

	return twitchClient, nil
}

func (s Store) GetTwitchClientForChannel(channel string, cfg ClientConfig) (*twitch.Client, error) {
	row := s.db.DB().QueryRow(
		`SELECT access_token, refresh_token, scopes
			FROM extended_permissions
			WHERE channel = $1`,
		channel,
	)

	var accessToken, refreshToken, scopeStr string
	if err := row.Scan(&accessToken, &refreshToken, &scopeStr); err != nil {
		return nil, errors.Wrap(err, "getting twitch credentials from database")
	}

	scopes := strings.Split(scopeStr, " ")

	tc := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, accessToken, refreshToken)
	tc.SetTokenUpdateHook(func(at, rt string) error {
		return errors.Wrap(s.SetExtendedTwitchCredentials(channel, at, rt, scopes), "updating extended permissions token")
	})

	return tc, nil
}

func (s Store) HasAnyPermissionForChannel(channel string, scopes ...string) (bool, error) {
	row := s.db.DB().QueryRow(
		`SELECT scopes
			FROM extended_permissions
			WHERE channel = $1`,
		channel,
	)

	var scopeStr string
	if err := row.Scan(&scopeStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.Wrap(err, "getting scopes from database")
	}

	storedScopes := strings.Split(scopeStr, " ")

	for _, scope := range scopes {
		if str.StringInSlice(scope, storedScopes) {
			return true, nil
		}
	}

	return false, nil
}

func (s Store) HasPermissionsForChannel(channel string, scopes ...string) (bool, error) {
	row := s.db.DB().QueryRow(
		`SELECT scopes
			FROM extended_permissions
			WHERE channel = $1`,
		channel,
	)

	var scopeStr string
	if err := row.Scan(&scopeStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.Wrap(err, "getting scopes from database")
	}

	storedScopes := strings.Split(scopeStr, " ")

	for _, scope := range scopes {
		if !str.StringInSlice(scope, storedScopes) {
			return false, nil
		}
	}

	return true, nil
}

func (s Store) RemoveExendedTwitchCredentials(channel string) error {
	_, err := s.db.DB().Exec(
		`DELETE FROM extended_permissions
			WHERE channel = $1`,
		channel,
	)

	return errors.Wrap(err, "deleting data from table")
}

func (s Store) SetBotTwitchCredentials(accessToken, refreshToken string) (err error) {
	if err = s.db.StoreCoreMeta(coreMetaKeyBotToken, accessToken); err != nil {
		return errors.Wrap(err, "storing bot access token")
	}

	if err = s.db.StoreCoreMeta(coreMetaKeyBotRefreshToken, refreshToken); err != nil {
		return errors.Wrap(err, "storing bot refresh token")
	}

	return nil
}

func (s Store) SetExtendedTwitchCredentials(channel, accessToken, refreshToken string, scope []string) (err error) {
	_, err = s.db.DB().Exec(
		`INSERT INTO extended_permissions
			(channel, access_token, refresh_token, scopes)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO UPDATE SET
				access_token=excluded.access_token,
				refresh_token=excluded.refresh_token,
				scopes=excluded.scopes;`,
		channel, accessToken, refreshToken, strings.Join(scope, " "),
	)

	return errors.Wrap(err, "inserting data into table")
}
