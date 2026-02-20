package main

import (
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/gofrs/uuid/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/irc.v4"
	"gopkg.in/yaml.v3"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	expectedMinConfigVersion = 2
	rawLogDirPerm            = 0o755
	rawLogFilePerm           = 0o644
)

var (
	//go:embed default_config.yaml
	defaultConfigurationYAML []byte

	hashstructUUIDNamespace = uuid.Must(uuid.FromString("3a0ccc46-d3ba-46b5-ac07-27528c933174"))

	errSaveNotRequired = errors.New("save not required")
)

type (
	configAuthToken struct {
		Hash    string   `json:"-" yaml:"hash"`
		Modules []string `json:"modules" yaml:"modules"`
		Name    string   `json:"name" yaml:"name"`
		Token   string   `json:"token" yaml:"-"`
	}

	configFileVersioner struct {
		ConfigVersion int64 `yaml:"config_version"`
	}

	configFile struct {
		AuthTokens           map[string]configAuthToken `yaml:"auth_tokens"`
		AutoMessages         []*autoMessage             `yaml:"auto_messages"`
		BotEditors           []string                   `yaml:"bot_editors"`
		Channels             []string                   `yaml:"channels"`
		GitTrackConfig       bool                       `yaml:"git_track_config"`
		HTTPListen           string                     `yaml:"http_listen"`
		PermitAllowModerator bool                       `yaml:"permit_allow_moderator"`
		PermitTimeout        time.Duration              `yaml:"permit_timeout"`
		RawLog               string                     `yaml:"raw_log"`
		ModuleConfig         plugins.ModuleConfig       `yaml:"module_config"`
		Rules                []*plugins.Rule            `yaml:"rules"`
		Variables            map[string]interface{}     `yaml:"variables"`

		rawLogWriter io.WriteCloser

		configFileVersioner `yaml:",inline"`
	}
)

func newConfigFile() *configFile {
	return &configFile{
		AuthTokens:    map[string]configAuthToken{},
		PermitTimeout: time.Minute,
	}
}

func loadConfig(filename string) error {
	var (
		configVersion = &configFileVersioner{}
		err           error
		tmpConfig     = newConfigFile()
	)

	if err = parseConfigFromYAML(filename, configVersion, false); err != nil {
		return errors.Wrap(err, "parsing config version")
	}

	if configVersion.ConfigVersion < expectedMinConfigVersion {
		return errors.Errorf("config version too old: %d < %d - Please have a look at the documentation", configVersion.ConfigVersion, expectedMinConfigVersion)
	}

	if err = parseConfigFromYAML(filename, tmpConfig, true); err != nil {
		return errors.Wrap(err, "parsing config")
	}

	if err = tmpConfig.runLoadChecks(); err != nil {
		return errors.Wrap(err, "running load-checks on config")
	}

	configLock.Lock()
	defer configLock.Unlock()

	tmpConfig.updateAutoMessagesFromConfig(config)
	tmpConfig.fixDurations()
	tmpConfig.fixMissingUUIDs()
	if err = tmpConfig.fixTokenHashStorage(); err != nil {
		return errors.Wrap(err, "applying token hash fixes")
	}

	switch {
	case config != nil && config.RawLog == tmpConfig.RawLog:
		tmpConfig.rawLogWriter = config.rawLogWriter

	case tmpConfig.RawLog == "":
		if err = config.CloseRawMessageWriter(); err != nil {
			return errors.Wrap(err, "closing old raw log writer")
		}

		tmpConfig.rawLogWriter = writeNoOpCloser{io.Discard}

	default:
		if err = config.CloseRawMessageWriter(); err != nil {
			return errors.Wrap(err, "closing old raw log writer")
		}
		if err = os.MkdirAll(path.Dir(tmpConfig.RawLog), rawLogDirPerm); err != nil {
			return errors.Wrap(err, "creating directories for raw log")
		}
		if tmpConfig.rawLogWriter, err = os.OpenFile(tmpConfig.RawLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, rawLogFilePerm); err != nil {
			return errors.Wrap(err, "opening raw log for appending")
		}
	}

	config = tmpConfig
	timerService.UpdatePermitTimeout(tmpConfig.PermitTimeout)

	logrus.WithFields(logrus.Fields{
		"auto_messages": len(config.AutoMessages),
		"rules":         len(config.Rules),
		"channels":      len(config.Channels),
	}).Info("Config file (re)loaded")

	// Notify listener config has changed
	frontendNotifyHooks.Ping(frontendNotifyTypeReload)

	return nil
}

func parseConfigFromYAML(filename string, obj interface{}, strict bool) error {
	f, err := os.Open(filename) //#nosec:G304 // This is intended to open a variable file
	if err != nil {
		return errors.Wrap(err, "open config file")
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.WithError(err).Error("closing config file (leaked fd)")
		}
	}()

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(strict)

	return errors.Wrap(decoder.Decode(obj), "decode config file")
}

func patchConfig(filename, authorName, authorEmail, summary string, patcher func(*configFile) error) error {
	var (
		cfgFile = newConfigFile()
		err     error
	)

	if err = parseConfigFromYAML(filename, cfgFile, true); err != nil {
		return errors.Wrap(err, "loading current config")
	}

	cfgFile.fixMissingUUIDs()
	if err = cfgFile.fixTokenHashStorage(); err != nil {
		return errors.Wrap(err, "applying token hash fixes")
	}

	err = patcher(cfgFile)
	switch {
	case errors.Is(err, nil):
		// This is fine

	case errors.Is(err, errSaveNotRequired):
		// This is also fine but we don't need to save
		return nil

	default:
		return errors.Wrap(err, "patching config")
	}

	if err = cfgFile.runLoadChecks(); err != nil {
		return errors.Wrap(err, "checking config after patch")
	}

	return errors.Wrap(
		writeConfigToYAML(filename, authorName, authorEmail, summary, cfgFile),
		"replacing config",
	)
}

func writeConfigToYAML(filename, authorName, authorEmail, summary string, obj *configFile) error {
	tmpFile, err := os.CreateTemp(path.Dir(filename), "twitch-bot-*.yaml")
	if err != nil {
		return errors.Wrap(err, "opening tempfile")
	}
	tmpFileName := tmpFile.Name()

	if _, err = fmt.Fprintf(tmpFile, "# Automatically updated by %s using Config-Editor frontend, last update: %s\n", authorName, time.Now().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("writing file header: %w", err)
	}

	if err = yaml.NewEncoder(tmpFile).Encode(obj); err != nil {
		tmpFile.Close() //nolint:errcheck,gosec
		return errors.Wrap(err, "encoding config")
	}

	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp config: %w", err)
	}

	if err = os.Rename(tmpFileName, filename); err != nil {
		return errors.Wrap(err, "moving config to location")
	}

	if !obj.GitTrackConfig {
		return nil
	}

	git := newGitHelper(path.Dir(filename))
	if !git.HasRepo() {
		logrus.Error("Instructed to track changes using Git, but config not in repo")
		return nil
	}

	return errors.Wrap(
		git.CommitChange(path.Base(filename), authorName, authorEmail, summary),
		"committing config changes",
	)
}

func writeDefaultConfigFile(filename string) error {
	f, err := os.Create(filename) //#nosec:G304 // This is intended to open a variable file
	if err != nil {
		return errors.Wrap(err, "creating config file")
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.WithError(err).Error("closing config file (leaked fd)")
		}
	}()

	_, err = f.Write(defaultConfigurationYAML)
	return errors.Wrap(err, "writing default config")
}

func (c configAuthToken) validate(token string) error {
	switch {
	case strings.HasPrefix(c.Hash, "$2a$"):
		return errors.Wrap(
			bcrypt.CompareHashAndPassword([]byte(c.Hash), []byte(token)),
			"validating bcrypt",
		)

	case strings.HasPrefix(c.Hash, "$argon2id$"):
		var (
			flds = strings.Split(c.Hash, "$")
			t, m uint32
			p    uint8
		)

		if _, err := fmt.Sscanf(flds[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
			return errors.Wrap(err, "scanning argon2id hash params")
		}

		salt, err := base64.RawStdEncoding.DecodeString(flds[4])
		if err != nil {
			return errors.Wrap(err, "decoding salt")
		}

		if flds[5] == base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(token), salt, t, m, p, argonHashLen)) {
			return nil
		}

		return errors.New("hash does not match")

	default:
		return errors.New("unknown hash format found")
	}
}

func (c *configFile) CloseRawMessageWriter() (err error) {
	if c == nil || c.rawLogWriter == nil {
		return nil
	}

	if err = c.rawLogWriter.Close(); err != nil {
		return fmt.Errorf("closing raw-log writer: %w", err)
	}

	return nil
}

func (c configFile) GetMatchingRules(m *irc.Message, event *string, eventData *fieldcollection.FieldCollection) []*plugins.Rule {
	configLock.RLock()
	defer configLock.RUnlock()

	var out []*plugins.Rule

	for _, r := range c.Rules {
		if r.Matches(m, event, timerService, formatMessage, twitchClient, eventData) {
			out = append(out, r)
		}
	}

	return out
}

func (c configFile) LogRawMessage(m *irc.Message) error {
	_, err := fmt.Fprintln(c.rawLogWriter, m.String())
	return errors.Wrap(err, "writing raw log message")
}

func (c *configFile) fixDurations() {
	// General fields
	c.PermitTimeout = c.fixedDuration(c.PermitTimeout)

	// Fix rules
	for _, r := range c.Rules {
		r.ChannelCooldown = c.fixedDurationPtr(r.ChannelCooldown)
		r.Cooldown = c.fixedDurationPtr(r.Cooldown)
		r.UserCooldown = c.fixedDurationPtr(r.UserCooldown)
	}
}

func (configFile) fixedDuration(d time.Duration) time.Duration {
	if d > time.Second {
		return d
	}
	return d * time.Second //nolint:durationcheck // Error is handled before
}

func (configFile) fixedDurationPtr(d *time.Duration) *time.Duration {
	if d == nil || *d >= time.Second {
		return d
	}
	fd := *d * time.Second //nolint:durationcheck // Error is handled before
	return &fd
}

func (c *configFile) fixMissingUUIDs() {
	for i := range c.AutoMessages {
		if c.AutoMessages[i].UUID != "" {
			continue
		}
		c.AutoMessages[i].UUID = uuid.NewV5(hashstructUUIDNamespace, c.AutoMessages[i].ID()).String()
	}

	for i := range c.Rules {
		if c.Rules[i].UUID != "" {
			continue
		}
		c.Rules[i].UUID = uuid.NewV5(hashstructUUIDNamespace, c.Rules[i].MatcherID()).String()
	}
}

func (c *configFile) fixTokenHashStorage() (err error) {
	for key := range c.AuthTokens {
		auth := c.AuthTokens[key]

		if strings.HasPrefix(auth.Hash, "$") {
			continue
		}

		rawHash, err := hex.DecodeString(auth.Hash)
		if err != nil {
			return errors.Wrap(err, "reading hash")
		}

		auth.Hash = string(rawHash)
		c.AuthTokens[key] = auth
	}

	return nil
}

func (c *configFile) runLoadChecks() (err error) {
	if len(c.Channels) == 0 {
		logrus.Warn("Loaded config with empty channel list")
	}

	if len(c.Rules) == 0 {
		logrus.Warn("Loaded config with empty ruleset")
	}

	var seen []string
	for _, r := range c.Rules {
		if r.UUID != "" && slices.Contains(seen, r.UUID) {
			return errors.New("duplicate rule UUIDs found")
		}
		seen = append(seen, r.UUID)
	}

	if err = c.validateRuleActions(); err != nil {
		return errors.Wrap(err, "validating rule actions")
	}

	return nil
}

func (c *configFile) updateAutoMessagesFromConfig(old *configFile) {
	for idx, nam := range c.AutoMessages {
		// By default assume last message to be sent now
		// in order not to spam messages at startup
		nam.lastMessageSent = time.Now()

		if !nam.IsValid() {
			logrus.WithField("index", idx).Warn("Auto-Message configuration is invalid and therefore disabled")
		}

		if old == nil {
			// Initial config load, do not update timers
			continue
		}

		for _, oam := range old.AutoMessages {
			if nam.ID() != oam.ID() {
				continue
			}

			// We disable the old message as executing it would
			// mess up the constraints of the new message
			oam.lock.Lock()
			oam.disabled = true

			nam.lastMessageSent = oam.lastMessageSent
			nam.linesSinceLastMessage = oam.linesSinceLastMessage
			oam.lock.Unlock()
		}
	}
}

func (c configFile) validateRuleActions() error {
	var hasError bool

	for _, r := range c.Rules {
		logger := logrus.WithField("rule", r.MatcherID())

		if err := r.Validate(validateTemplate); err != nil {
			logger.WithError(err).Error("Rule reported invalid config")
			hasError = true
		}

		for idx, a := range r.Actions {
			actor, err := getActorByName(a.Type)
			if err != nil {
				logger.WithField("index", idx).WithError(err).Error("Cannot get actor by type")
				hasError = true
				continue
			}

			if err = actor.Validate(validateTemplate, a.Attributes); err != nil {
				logger.WithField("index", idx).WithError(err).Error("Actor reported invalid config")
				hasError = true
			}
		}
	}

	if hasError {
		return errors.New("config validation reported errors, see log")
	}

	return nil
}
