package v2migrator

import (
	"compress/gzip"
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/pkg/database"
	"github.com/Luzifer/twitch-bot/internal/v2migrator/crypt"
	"github.com/Luzifer/twitch-bot/plugins"
)

const eventSubSecretLength = 32

var errExtendedPermissionsMissing = errors.New("no extended permissions greanted")

type (
	Migrator interface {
		Load(filename, encryptionPass string) error
		Migrate(db database.Connector) error
	}

	storageExtendedPermission struct {
		AccessToken  string   `encrypt:"true" json:"access_token,omitempty"`
		RefreshToken string   `encrypt:"true" json:"refresh_token,omitempty"`
		Scopes       []string `json:"scopes,omitempty"`
	}

	storageFile struct {
		Counters  map[string]int64              `json:"counters"`
		Timers    map[string]plugins.TimerEntry `json:"timers"`
		Variables map[string]string             `json:"variables"`

		ModuleStorage struct {
			ModPunish   storageModPunish   `json:"44ab4646-ce50-4e16-9353-c1f0eb68962b"`
			ModOverlays storageModOverlays `json:"f9ca2b3a-baf6-45ea-a347-c626168665e8"`
			ModQuoteDB  storageModQuoteDB  `json:"917c83ee-ed40-41e4-a558-1c2e59fdf1f5"`
		} `json:"module_storage"`

		ExtendedPermissions map[string]*storageExtendedPermission `json:"extended_permissions"`

		EventSubSecret string `encrypt:"true" json:"event_sub_secret,omitempty"`

		BotAccessToken  string `encrypt:"true" json:"bot_access_token,omitempty"`
		BotRefreshToken string `encrypt:"true" json:"bot_refresh_token,omitempty"`
	}
)

func NewStorageFile() Migrator {
	return &storageFile{
		Counters:  map[string]int64{},
		Timers:    map[string]plugins.TimerEntry{},
		Variables: map[string]string{},

		ExtendedPermissions: map[string]*storageExtendedPermission{},
	}
}

func (s *storageFile) Load(filename, encryptionPass string) error {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Store init state
			return nil
		}
		return errors.Wrap(err, "open storage file")
	}
	defer f.Close()

	zf, err := gzip.NewReader(f)
	if err != nil {
		return errors.Wrap(err, "create gzip reader")
	}
	defer zf.Close()

	if err = json.NewDecoder(zf).Decode(s); err != nil {
		return errors.Wrap(err, "decode storage object")
	}

	if err = crypt.DecryptFields(s, encryptionPass); err != nil {
		return errors.Wrap(err, "decrypting storage object")
	}

	return nil
}

func (s storageFile) Migrate(db database.Connector) error {
	var bat string
	err := db.ReadCoreMeta("bot_access_token", &bat)
	switch {
	case err == nil:
		return errors.New("Access token is set, database already initialized")

	case errors.Is(err, database.ErrCoreMetaNotFound):
		// This is the expected state

	default:
		return errors.Wrap(err, "checking for bot access token")
	}

	for name, fn := range map[string]func(database.Connector) error{
		// Core
		"core":        s.migrateCoreKV,
		"counter":     s.migrateCounters,
		"permissions": s.migratePermissions,
		"timers":      s.migrateTimers,
		"variables":   s.migrateVariables,
		// Modules
		"mod_punish":   s.ModuleStorage.ModPunish.migrate,
		"mod_overlays": s.ModuleStorage.ModOverlays.migrate,
		"mod_quotedb":  s.ModuleStorage.ModQuoteDB.migrate,
	} {
		logrus.WithField("module", name).Info("Starting migration...")
		if err = fn(db); err != nil {
			return errors.Wrapf(err, "executing %q migration", name)
		}
	}

	return nil
}
