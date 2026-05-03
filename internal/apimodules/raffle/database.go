package raffle

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"
	"gorm.io/gorm"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

const (
	frontendNotifyEventRaffleChange      = "raffleChanged"
	frontendNotifyEventRaffleEntryChange = "raffleEntryChanged"
)

const (
	raffleMessageEventEntryFailed raffleMessageEvent = iota
	raffleMessageEventEntry
	raffleMessageEventReminder
	raffleMessageEventWin
	raffleMessageEventClose
)

const (
	raffleStatusPlanned raffleStatus = "planned"
	raffleStatusActive  raffleStatus = "active"
	raffleStatusEnded   raffleStatus = "ended"
)

type (
	dbClient struct {
		activeRaffles map[string]uint64
		speakUp       map[string]*speakUpWait
		db            database.Connector
		lock          sync.RWMutex
	}

	raffle struct {
		ID uint64 `gorm:"primaryKey" json:"id,string"`

		Channel string       `json:"channel"`
		Keyword string       `json:"keyword"`
		Title   string       `json:"title"`
		Status  raffleStatus `gorm:"default:planned" json:"status"`

		AllowEveryone   bool          `json:"allowEveryone"`
		AllowFollower   bool          `json:"allowFollower"`
		AllowSubscriber bool          `json:"allowSubscriber"`
		AllowVIP        bool          `gorm:"column:allow_vip" json:"allowVIP"`
		MinFollowAge    time.Duration `json:"minFollowAge"`

		MultiFollower   float64 `json:"multiFollower"`
		MultiSubscriber float64 `json:"multiSubscriber"`
		MultiVIP        float64 `gorm:"column:multi_vip" json:"multiVIP"`

		AutoStartAt     *time.Time    `json:"autoStartAt"`
		CloseAfter      time.Duration `json:"closeAfter"`
		CloseAt         *time.Time    `json:"closeAt"`
		WaitForResponse time.Duration `json:"waitForResponse"`

		TextClose            string        `json:"textClose"`
		TextClosePost        bool          `json:"textClosePost"`
		TextEntry            string        `json:"textEntry"`
		TextEntryPost        bool          `json:"textEntryPost"`
		TextEntryFail        string        `json:"textEntryFail"`
		TextEntryFailPost    bool          `json:"textEntryFailPost"`
		TextReminder         string        `json:"textReminder"`
		TextReminderInterval time.Duration `json:"textReminderInterval"`
		TextReminderNextSend *time.Time    `json:"-"`
		TextReminderPost     bool          `json:"textReminderPost"`
		TextWin              string        `json:"textWin"`
		TextWinPost          bool          `json:"textWinPost"`

		Entries []raffleEntry `gorm:"foreignKey:RaffleID" json:"entries,omitempty"`
	}

	raffleEntry struct {
		ID       uint64 `gorm:"primaryKey" json:"id,string"`
		RaffleID uint64 `gorm:"uniqueIndex:user_per_raffle" json:"-"`

		UserID          string `gorm:"size:128;uniqueIndex:user_per_raffle" json:"userID"`
		UserLogin       string `json:"userLogin"`
		UserDisplayName string `json:"userDisplayName"`

		EnteredAt  time.Time `json:"enteredAt"`
		EnteredAs  string    `json:"enteredAs"`
		Multiplier float64   `json:"multiplier"`

		WasPicked  bool `json:"wasPicked"`
		WasRedrawn bool `json:"wasRedrawn"`

		DrawResponse string     `json:"drawResponse,omitempty"`
		SpeakUpUntil *time.Time `json:"speakUpUntil,omitempty"`
	}

	raffleMessageEvent uint8
	raffleStatus       string

	speakUpWait struct {
		RaffleEntryID uint64
		Until         time.Time
	}
)

var errRaffleNotFound = errors.New("raffle not found")

func newDBClient(db database.Connector) *dbClient {
	return &dbClient{
		activeRaffles: make(map[string]uint64),
		speakUp:       make(map[string]*speakUpWait),
		db:            db,
	}
}

// AutoCloseExpired collects all active raffles which have overdue
// close_at dates and closes them
func (d *dbClient) AutoCloseExpired() (err error) {
	var rr []raffle

	if err = helpers.Retry(func() error {
		return d.db.DB().
			Where("status = ? AND close_at IS NOT NULL AND close_at < ?", raffleStatusActive, time.Now().UTC()).
			Find(&rr).
			Error
	}); err != nil {
		return fmt.Errorf("fetching raffles to close: %w", err)
	}

	for _, r := range rr {
		if err = d.Close(r.ID); err != nil {
			return fmt.Errorf("closing raffle %d: %w", r.ID, err)
		}
	}

	return nil
}

// AutoSendReminders collects all active raffles which have enabled
// reminders which are overdue and posts the reminder for them
func (d *dbClient) AutoSendReminders() (err error) {
	var rr []raffle

	if err = helpers.Retry(func() error {
		return d.db.DB().
			Where("status = ? AND text_reminder_post = ? AND (text_reminder_next_send IS NULL OR text_reminder_next_send < ?)", raffleStatusActive, true, time.Now().UTC()).
			Find(&rr).
			Error
	}); err != nil {
		return fmt.Errorf("fetching raffles to send reminders: %w", err)
	}

	for _, r := range rr {
		if err = r.SendEvent(raffleMessageEventReminder, nil); err != nil {
			return fmt.Errorf("sending reminder for raffle %d: %w", r.ID, err)
		}
	}

	return nil
}

// AutoStart collects planned and overdue raffles and starts them
func (d *dbClient) AutoStart() (err error) {
	var rr []raffle

	if err = helpers.Retry(func() error {
		return d.db.DB().
			Where("status = ? AND auto_start_at IS NOT NULL AND auto_start_at < ?", raffleStatusPlanned, time.Now().UTC()).
			Find(&rr).
			Error
	}); err != nil {
		return fmt.Errorf("fetching raffles to start: %w", err)
	}

	for _, r := range rr {
		if err = d.Start(r.ID); err != nil {
			return fmt.Errorf("starting raffle %d: %w", r.ID, err)
		}
	}

	return nil
}

// Clone duplicates a raffle into a new draft resetting some
// parameters into their default state
func (d *dbClient) Clone(raffleID uint64) error {
	raffle, err := d.Get(raffleID)
	if err != nil {
		return fmt.Errorf("getting raffle: %w", err)
	}

	raffle.AutoStartAt = nil
	raffle.CloseAt = nil
	raffle.Entries = nil
	raffle.ID = 0
	raffle.Status = raffleStatusPlanned
	raffle.Title = strings.Join([]string{"Copy of", raffle.Title}, " ")

	if err = d.Create(raffle); err != nil {
		return fmt.Errorf("creating copy: %w", err)
	}

	frontendNotify(frontendNotifyEventRaffleChange)
	return nil
}

// Close marks the raffle as closed and removes it from the active
// raffle cache
func (d *dbClient) Close(raffleID uint64) error {
	r, err := d.Get(raffleID)
	if err != nil {
		return fmt.Errorf("getting raffle: %w", err)
	}

	if err = helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		return tx.Model(&raffle{}).
			Where("id = ?", raffleID).
			Update("status", raffleStatusEnded).
			Error
	}); err != nil {
		return fmt.Errorf("setting status closed: %w", err)
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.activeRaffles, strings.Join([]string{r.Channel, r.Keyword}, "::"))

	frontendNotify(frontendNotifyEventRaffleChange)

	if err = r.SendEvent(raffleMessageEventClose, nil); err != nil {
		return fmt.Errorf("sending close-message: %w", err)
	}

	return nil
}

// Create creates a new raffle. The record will be written to
// the database without modification and therefore need to be filled
// before calling this function
func (d *dbClient) Create(r raffle) error {
	if err := helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		return tx.Create(&r).Error
	}); err != nil {
		return fmt.Errorf("creating database record: %w", err)
	}

	frontendNotify(frontendNotifyEventRaffleChange)
	return nil
}

// Delete removes all entries for the given raffle and afterwards
// deletes the raffle itself
func (d *dbClient) Delete(raffleID uint64) (err error) {
	if err = helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		if err = tx.
			Where("raffle_id = ?", raffleID).
			Delete(&raffleEntry{}).
			Error; err != nil {
			return fmt.Errorf("deleting raffle entries: %w", err)
		}

		if err = tx.
			Where("id = ?", raffleID).
			Delete(&raffle{}).Error; err != nil {
			return fmt.Errorf("creating database record: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("deleting raffle: %w", err)
	}

	frontendNotify(frontendNotifyEventRaffleChange)
	return nil
}

// Enter creates a new raffle entry. The entry will be written to
// the database without modification and therefore need to be filled
// before calling this function
func (d *dbClient) Enter(re raffleEntry) error {
	if err := helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error { return tx.Create(&re).Error }); err != nil {
		return fmt.Errorf("creating database record: %w", err)
	}

	frontendNotify(frontendNotifyEventRaffleEntryChange)
	return nil
}

// Get retrieves a raffle from the database
func (d *dbClient) Get(raffleID uint64) (out raffle, err error) {
	if err = helpers.Retry(func() error {
		return d.db.DB().
			Where("raffles.id = ?", raffleID).
			Preload("Entries").
			First(&out).
			Error
	}); err != nil {
		return out, fmt.Errorf("getting raffle from database: %w", err)
	}

	return out, nil
}

// GetByChannelAndKeyword resolves an active raffle through channel
// and keyword given in the raffle and returns it through the Get
// function. If the combination is not known errRaffleNotFound is
// returned.
func (d *dbClient) GetByChannelAndKeyword(channel, keyword string) (raffle, error) {
	d.lock.RLock()
	id := d.activeRaffles[strings.Join([]string{channel, keyword}, "::")]
	d.lock.RUnlock()

	if id == 0 {
		return raffle{}, errRaffleNotFound
	}

	return d.Get(id)
}

// List returns a list of all known raffles
func (d *dbClient) List() (raffles []raffle, err error) {
	if err = helpers.Retry(func() error {
		return d.db.DB().Model(&raffle{}).
			Order("id DESC").
			Find(&raffles).
			Error
	}); err != nil {
		return nil, fmt.Errorf("getting raffles: %w", err)
	}

	return raffles, nil
}

// PatchNextReminderSend updates the time another reminder shall be
// sent for the given raffle ID. No other fields are modified
func (d *dbClient) PatchNextReminderSend(raffleID uint64, next time.Time) error {
	if err := helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		return tx.Model(&raffle{}).
			Where("id = ?", raffleID).
			Update("text_reminder_next_send", next).
			Error
	}); err != nil {
		return fmt.Errorf("updating column: %w", err)
	}

	return nil
}

// PickWinner fetches the given raffle and picks a random winner
// based on entries and their multiplier
func (d *dbClient) PickWinner(raffleID uint64) error {
	r, err := d.Get(raffleID)
	if err != nil {
		return fmt.Errorf("getting raffle: %w", err)
	}

	winner, err := pickWinnerFromRaffle(r)
	if err != nil {
		return fmt.Errorf("picking winner: %w", err)
	}

	speakUpUntil := time.Now().UTC().Add(r.WaitForResponse)
	if err = helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		return tx.Model(&raffleEntry{}).
			Where("id = ?", winner.ID).
			Updates(map[string]any{"was_picked": true, "speak_up_until": speakUpUntil}).
			Error
	}); err != nil {
		return fmt.Errorf("updating winner: %w", err)
	}

	d.lock.Lock()
	d.speakUp[strings.Join([]string{r.Channel, winner.UserLogin}, ":")] = &speakUpWait{RaffleEntryID: winner.ID, Until: speakUpUntil}
	d.lock.Unlock()

	fields := fieldcollection.FieldCollectionFromData(map[string]any{
		"user_id": winner.UserID,
		"user":    winner.UserLogin,
		"winner":  winner,
	})

	frontendNotify(frontendNotifyEventRaffleEntryChange)

	if err = r.SendEvent(raffleMessageEventWin, fields); err != nil {
		return fmt.Errorf("sending win-message: %w", err)
	}

	return nil
}

// RedrawWinner marks the previous winner as redrawn (and therefore
// crossed out as winner in the interface) and picks a new one
func (d *dbClient) RedrawWinner(raffleID, winnerID uint64) error {
	if err := helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		return tx.Model(&raffleEntry{}).
			Where("id = ?", winnerID).
			Update("was_redrawn", true).
			Error
	}); err != nil {
		return fmt.Errorf("updating previous winner: %w", err)
	}

	return d.PickWinner(raffleID)
}

// RefreshActiveRaffles loads all active raffles and populates the
// activeRaffles cache
func (d *dbClient) RefreshActiveRaffles() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	var (
		actives []raffle
		tmp     = make(map[string]uint64)
	)

	if err := helpers.Retry(func() error {
		return d.db.DB().
			Where("status = ?", raffleStatusActive).
			Find(&actives).
			Error
	}); err != nil {
		return fmt.Errorf("fetching active raffles: %w", err)
	}

	for _, r := range actives {
		tmp[strings.Join([]string{r.Channel, r.Keyword}, "::")] = r.ID
	}

	d.activeRaffles = tmp
	return nil
}

// RefreshSpeakUp seeks all still active speak-up entries and
// populates the speak-up cache with them
func (d *dbClient) RefreshSpeakUp() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	var (
		res []raffleEntry
		tmp = make(map[string]*speakUpWait)
	)

	if err := helpers.Retry(func() error {
		return d.db.DB().Debug().
			Where("speak_up_until IS NOT NULL AND speak_up_until > ?", time.Now().UTC()).
			Find(&res).
			Error
	}); err != nil {
		return fmt.Errorf("querying active entries: %w", err)
	}

	for _, e := range res {
		var r raffle
		if err := helpers.Retry(func() error {
			return d.db.DB().
				Where("id = ?", e.RaffleID).
				First(&r).
				Error
		}); err != nil {
			return fmt.Errorf("fetching raffle for entry: %w", err)
		}
		tmp[strings.Join([]string{r.Channel, e.UserLogin}, ":")] = &speakUpWait{RaffleEntryID: e.ID, Until: *e.SpeakUpUntil}
	}

	d.speakUp = tmp
	return nil
}

// RegisterSpeakUp sets the speak-up message if there was a
// speakUpWait for that user and channel
func (d *dbClient) RegisterSpeakUp(channel, user, message string) error {
	d.lock.RLock()
	w := d.speakUp[strings.Join([]string{channel, user}, ":")]
	d.lock.RUnlock()

	if w == nil || w.Until.Before(time.Now()) {
		// No speak-up-request for that user or expired
		return nil
	}

	if err := helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		return tx.Model(&raffleEntry{}).
			Where("id = ?", w.RaffleEntryID).
			Updates(map[string]any{
				"DrawResponse": message,
				"SpeakUpUntil": nil,
			}).
			Error
	}); err != nil {
		return fmt.Errorf("registering speak-up: %w", err)
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.speakUp, strings.Join([]string{channel, user}, ":"))

	frontendNotify(frontendNotifyEventRaffleEntryChange)
	return nil
}

// Reopen updates the CloseAt attribute and status to active to
// prolong the raffle
func (d *dbClient) Reopen(raffleID uint64, duration time.Duration) error {
	r, err := d.Get(raffleID)
	if err != nil {
		return fmt.Errorf("getting specified raffle: %w", err)
	}

	if err = helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		return tx.Model(&raffle{}).
			Where("id = ?", raffleID).
			Updates(map[string]any{
				"CloseAt": time.Now().UTC().Add(duration),
				"status":  raffleStatusActive,
			}).
			Error
	}); err != nil {
		return fmt.Errorf("updating raffle: %w", err)
	}

	// Store ID to active-raffle cache
	d.lock.Lock()
	defer d.lock.Unlock()
	d.activeRaffles[strings.Join([]string{r.Channel, r.Keyword}, "::")] = r.ID

	frontendNotify(frontendNotifyEventRaffleChange)
	return nil
}

// Reset resets the raffle participants, the status, start / close
// times while preserving the rest of the settings
func (d *dbClient) Reset(raffleID uint64) error {
	raffle, err := d.Get(raffleID)
	if err != nil {
		return fmt.Errorf("getting raffle: %w", err)
	}

	raffle.AutoStartAt = nil
	raffle.CloseAt = nil
	raffle.Entries = nil
	raffle.Status = raffleStatusPlanned

	if err = helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		if err = tx.Delete(&raffleEntry{}, "raffle_id = ?", raffleID).Error; err != nil {
			return fmt.Errorf("deleting raffle entries: %w", err)
		}

		return tx.Save(raffle).Error
	}); err != nil {
		return fmt.Errorf("saving cleaned raffle: %w", err)
	}

	frontendNotify(frontendNotifyEventRaffleChange)
	return nil
}

// Start fetches the given raffle, updates its CloseAt attribute
// in case it is not already set, sets the raffle to active, updates
// the raffle in the database and notes its channel/keyword combo
// into the activeRaffles cache for use with irc handling
func (d *dbClient) Start(raffleID uint64) error {
	r, err := d.Get(raffleID)
	if err != nil {
		return fmt.Errorf("getting specified raffle: %w", err)
	}

	if r.CloseAt == nil {
		end := time.Now().UTC().Add(r.CloseAfter)
		r.CloseAt = &end
	}

	r.Status = raffleStatusActive
	if err = d.Update(r); err != nil {
		return fmt.Errorf("updating raffle: %w", err)
	}

	// Store ID to active-raffle cache
	d.lock.Lock()
	defer d.lock.Unlock()
	d.activeRaffles[strings.Join([]string{r.Channel, r.Keyword}, "::")] = r.ID

	frontendNotify(frontendNotifyEventRaffleChange)

	if err = r.SendEvent(raffleMessageEventReminder, nil); err != nil {
		return fmt.Errorf("sending first reminder: %w", err)
	}

	return nil
}

// Update stores the given raffle to the database. The ID within the
// raffle object must be set in order to update it. The object must
// be completely filled.
func (d *dbClient) Update(r raffle) error {
	old, err := d.Get(r.ID)
	if err != nil {
		return fmt.Errorf("getting previous version: %w", err)
	}

	// These information must not be changed after raffle has been started
	if old.Status != raffleStatusPlanned {
		r.Channel = old.Channel
		r.Keyword = old.Keyword
		r.Status = old.Status

		r.AllowEveryone = old.AllowEveryone
		r.AllowFollower = old.AllowFollower
		r.AllowSubscriber = old.AllowSubscriber
		r.AllowVIP = old.AllowVIP
		r.MinFollowAge = old.MinFollowAge

		r.MultiFollower = old.MultiFollower
		r.MultiSubscriber = old.MultiSubscriber
		r.MultiVIP = old.MultiVIP

		r.AutoStartAt = old.AutoStartAt
	}

	// This info must be preserved
	r.Entries = nil
	r.TextReminderNextSend = old.TextReminderNextSend

	if err := helpers.RetryTransaction(d.db.DB(), func(tx *gorm.DB) error {
		return tx.Model(&raffle{}).
			Where("id = ?", r.ID).
			Updates(&r).
			Error
	}); err != nil {
		return fmt.Errorf("updating raffle: %w", err)
	}

	frontendNotify(frontendNotifyEventRaffleChange)
	return nil
}

// SendEvent processes the text template and sends the message if
// enabled given through the event
func (r raffle) SendEvent(evt raffleMessageEvent, fields *fieldcollection.FieldCollection) (err error) {
	if fields == nil {
		fields = fieldcollection.NewFieldCollection()
	}

	fields.Set("raffle", r) // Make raffle available to templating

	var sendTextTpl string

	switch evt {
	case raffleMessageEventClose:
		if !r.TextClosePost {
			return nil
		}
		sendTextTpl = r.TextClose

	case raffleMessageEventEntryFailed:
		if !r.TextEntryFailPost {
			return nil
		}
		sendTextTpl = r.TextEntryFail

	case raffleMessageEventEntry:
		if !r.TextEntryPost {
			return nil
		}
		sendTextTpl = r.TextEntry

	case raffleMessageEventReminder:
		if !r.TextReminderPost {
			return nil
		}
		sendTextTpl = r.TextReminder
		if err = dbc.PatchNextReminderSend(r.ID, time.Now().UTC().Add(r.TextReminderInterval)); err != nil {
			return fmt.Errorf("updating next reminder for raffle: %w", err)
		}

	case raffleMessageEventWin:
		if !r.TextWinPost {
			return nil
		}
		sendTextTpl = r.TextWin

	default:
		// How?
		return errors.New("unexpected event")
	}

	msg, err := formatMessage(sendTextTpl, nil, nil, fields)
	if err != nil {
		return fmt.Errorf("formatting message to send: %w", err)
	}

	if err = send(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			"#" + strings.TrimLeft(r.Channel, "#"),
			msg,
		},
	}); err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	return nil
}
