package v2migrator

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/crypt"
	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
)

const eventSubSecretLength = 32

var errExtendedPermissionsMissing = errors.New("no extended permissions greanted")

type (
	storageExtendedPermission struct {
		AccessToken  string   `encrypt:"true" json:"access_token,omitempty"`
		RefreshToken string   `encrypt:"true" json:"refresh_token,omitempty"`
		Scopes       []string `json:"scopes,omitempty"`
	}

	storageFile struct {
		Counters  map[string]int64              `json:"counters"`
		Timers    map[string]plugins.TimerEntry `json:"timers"`
		Variables map[string]string             `json:"variables"`

		ModuleStorage map[string]json.RawMessage `json:"module_storage"`

		GrantedScopes       map[string][]string                   `json:"granted_scopes,omitempty"` // Deprecated, Read-Only
		ExtendedPermissions map[string]*storageExtendedPermission `json:"extended_permissions"`

		EventSubSecret string `encrypt:"true" json:"event_sub_secret,omitempty"`

		BotAccessToken  string `encrypt:"true" json:"bot_access_token,omitempty"`
		BotRefreshToken string `encrypt:"true" json:"bot_refresh_token,omitempty"`

		inMem bool
		lock  *sync.RWMutex
	}
)

func newStorageFile(inMemStore bool) *storageFile {
	return &storageFile{
		Counters:  map[string]int64{},
		Timers:    map[string]plugins.TimerEntry{},
		Variables: map[string]string{},

		ModuleStorage: map[string]json.RawMessage{},

		GrantedScopes:       map[string][]string{},
		ExtendedPermissions: map[string]*storageExtendedPermission{},

		inMem: inMemStore,
		lock:  new(sync.RWMutex),
	}
}

func (s *storageFile) DeleteExtendedPermissions(user string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.ExtendedPermissions, user)

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) DeleteModuleStore(moduleUUID string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.ModuleStorage, moduleUUID)

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) GetBotToken(fallback string) string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if v := s.BotAccessToken; v != "" {
		return v
	}
	return fallback
}

func (s *storageFile) GetCounterValue(counter string) int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Counters[counter]
}

func (s *storageFile) GetOrGenerateEventSubSecret() (string, string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.EventSubSecret != "" {
		return s.EventSubSecret, s.EventSubSecret[:5], nil
	}

	key := make([]byte, eventSubSecretLength)
	n, err := rand.Read(key)
	if err != nil {
		return "", "", errors.Wrap(err, "generating random secret")
	}
	if n != eventSubSecretLength {
		return "", "", errors.Errorf("read only %d of %d byte", n, eventSubSecretLength)
	}

	s.EventSubSecret = hex.EncodeToString(key)

	return s.EventSubSecret, s.EventSubSecret[:5], errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) GetModuleStore(moduleUUID string, storedObject plugins.StorageUnmarshaller) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return errors.Wrap(
		storedObject.UnmarshalStoredObject(s.ModuleStorage[moduleUUID]),
		"unmarshalling stored object",
	)
}

func (s *storageFile) GetTwitchClientForChannel(channel string) (*twitch.Client, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	perms := s.ExtendedPermissions[channel]
	if perms == nil {
		return nil, errExtendedPermissionsMissing
	}

	tc := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, perms.AccessToken, perms.RefreshToken)
	tc.SetTokenUpdateHook(func(at, rt string) error {
		return errors.Wrap(s.SetExtendedPermissions(channel, storageExtendedPermission{
			AccessToken:  at,
			RefreshToken: rt,
		}, true), "updating extended permissions token")
	})

	return tc, nil
}

func (s *storageFile) GetVariable(key string) string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Variables[key]
}

func (s *storageFile) HasTimer(id string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Timers[id].Time.After(time.Now())
}

func (s *storageFile) Load() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.inMem {
		// In-Memory store is active, do not load from disk
		// for testing purposes only!
		return nil
	}

	f, err := os.Open(cfg.StorageFile)
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

	if err = crypt.DecryptFields(s, cfg.StorageEncryptionPass); err != nil {
		return errors.Wrap(err, "decrypting storage object")
	}

	return errors.Wrap(s.migrate(), "migrating storage")
}

func (s *storageFile) Save() error {
	// NOTE(kahlers): DO NOT LOCK THIS, all calling functions are
	// modifying functions and must have locks in place

	if s.inMem {
		// In-Memory store is active, do not store to disk
		// for testing purposes only!
		return nil
	}

	// Cleanup timers
	var timerIDs []string
	for id := range s.Timers {
		timerIDs = append(timerIDs, id)
	}

	for _, i := range timerIDs {
		if s.Timers[i].Time.Before(time.Now()) {
			delete(s.Timers, i)
		}
	}

	// Encrypt fields in memory before writing
	if err := crypt.EncryptFields(s, cfg.StorageEncryptionPass); err != nil {
		return errors.Wrap(err, "encrypting storage object")
	}

	// Write store to disk
	f, err := os.Create(cfg.StorageFile)
	if err != nil {
		return errors.Wrap(err, "create storage file")
	}
	defer f.Close()

	zf := gzip.NewWriter(f)
	defer zf.Close()

	if err = json.NewEncoder(zf).Encode(s); err != nil {
		return errors.Wrap(err, "encode storage object")
	}

	// Decrypt the values to make them accessible again
	if err = crypt.DecryptFields(s, cfg.StorageEncryptionPass); err != nil {
		return errors.Wrap(err, "decrypting storage object")
	}

	return nil
}

func (s *storageFile) SetExtendedPermissions(user string, data storageExtendedPermission, merge bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	prev := s.ExtendedPermissions[user]
	if merge && prev != nil {
		for _, sc := range prev.Scopes {
			if !str.StringInSlice(sc, data.Scopes) {
				data.Scopes = append(data.Scopes, sc)
			}
		}

		if data.AccessToken == "" && prev.AccessToken != "" {
			data.AccessToken = prev.AccessToken
		}

		if data.RefreshToken == "" && prev.RefreshToken != "" {
			data.RefreshToken = prev.RefreshToken
		}
	}

	s.ExtendedPermissions[user] = &data

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) SetModuleStore(moduleUUID string, storedObject plugins.StorageMarshaller) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	data, err := storedObject.MarshalStoredObject()
	if err != nil {
		return errors.Wrap(err, "marshalling stored object")
	}

	s.ModuleStorage[moduleUUID] = data

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) SetTimer(kind plugins.TimerType, id string, expiry time.Time) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Timers[id] = plugins.TimerEntry{Kind: kind, Time: expiry}

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) SetVariable(key, value string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Variables[key] = value

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) RemoveVariable(key string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.Variables, key)

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) UpdateBotToken(accessToken, refreshToken string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.BotAccessToken = accessToken
	s.BotRefreshToken = refreshToken

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) UpdateCounter(counter string, value int64, absolute bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !absolute {
		value = s.Counters[counter] + value
	}

	if value == 0 {
		delete(s.Counters, counter)
	} else {
		s.Counters[counter] = value
	}

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) UserHasExtendedAuth(user string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ep := s.ExtendedPermissions[user]
	return ep != nil && ep.AccessToken != "" && ep.RefreshToken != ""
}

func (s *storageFile) UserHasGrantedAnyScope(user string, scopes ...string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.ExtendedPermissions[user] == nil {
		return false
	}

	grantedScopes := s.ExtendedPermissions[user].Scopes
	for _, scope := range scopes {
		if str.StringInSlice(scope, grantedScopes) {
			return true
		}
	}

	return false
}

func (s *storageFile) UserHasGrantedScopes(user string, scopes ...string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.ExtendedPermissions[user] == nil {
		return false
	}

	grantedScopes := s.ExtendedPermissions[user].Scopes
	for _, scope := range scopes {
		if !str.StringInSlice(scope, grantedScopes) {
			return false
		}
	}

	return true
}

func (s *storageFile) migrate() error {
	// Do NOT lock, use during locked call

	// Migration: Transform GrantedScopes and delete
	for ch, scopes := range s.GrantedScopes {
		if s.ExtendedPermissions[ch] != nil {
			continue
		}
		s.ExtendedPermissions[ch] = &storageExtendedPermission{Scopes: scopes}
	}
	s.GrantedScopes = nil

	return nil
}
