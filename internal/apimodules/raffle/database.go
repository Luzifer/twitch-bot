package raffle

import (
	"strings"
	"sync"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type (
	dbClient struct {
		activeRaffles map[string]uint64
		db            database.Connector
		lock          sync.RWMutex
	}

	raffle struct {
		ID uint64 `gorm:"primaryKey" json:"id"`

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

		TextEntry            string        `json:"textEntry"`
		TextEntryPost        bool          `json:"textEntryPost"`
		TextEntryFail        string        `json:"textEntryFail"`
		TextEntryFailPost    bool          `json:"textEntryFailPost"`
		TextWin              string        `json:"textWin"`
		TextWinPost          bool          `json:"textWinPost"`
		TextReminder         string        `json:"textReminder"`
		TextReminderInterval time.Duration `json:"textReminderInterval"`
		TextReminderNextSend time.Time     `json:"-"`
		TextReminderPost     bool          `json:"textReminderPost"`

		Entries []raffleEntry `gorm:"foreignKey:RaffleID" json:"entries,omitempty"`
	}

	raffleEntry struct {
		ID       uint64 `gorm:"primaryKey" json:"id"`
		RaffleID uint64 `gorm:"uniqueIndex:user_per_raffle" json:"-"`

		UserID          string `gorm:"uniqueIndex:user_per_raffle" json:"userID"`
		UserLogin       string `json:"userLogin"`
		UserDisplayName string `json:"userDisplayName"`

		EnteredAt  time.Time `json:"enteredAt"`
		EnteredAs  string    `json:"enteredAs"`
		Multiplier float64   `json:"multiplier"`

		WasPicked  bool `json:"wasPicked"`
		WasRedrawn bool `json:"wasRedrawn"`

		DrawResponse string `json:"drawResponse"`
	}

	raffleMessageEvent uint8
	raffleStatus       string
)

const (
	raffleMessageEventEntryFailed raffleMessageEvent = iota
	raffleMessageEventEntry
	raffleMessageEventReminder
	raffleMessageEventWin
)

const (
	raffleStatusPlanned raffleStatus = "planned"
	raffleStatusActive  raffleStatus = "active"
	raffleStatusEnded   raffleStatus = "ended"
)

var errRaffleNotFound = errors.New("raffle not found")

func newDBClient(db database.Connector) *dbClient {
	return &dbClient{
		activeRaffles: make(map[string]uint64),
		db:            db,
	}
}

// AutoCloseExpired collects all active raffles which have overdue
// close_at dates and closes them
func (d *dbClient) AutoCloseExpired() (err error) {
	var rr []raffle

	if err = d.db.DB().
		Where("status = ? AND close_at IS NOT NULL AND close_at < ?", raffleStatusActive, time.Now().UTC()).
		Find(&rr).
		Error; err != nil {
		return errors.Wrap(err, "fetching raffles to close")
	}

	for _, r := range rr {
		if err = d.Close(r.ID); err != nil {
			return errors.Wrapf(err, "closing raffle %d", r.ID)
		}
	}

	return nil
}

// AutoSendReminders collects all active raffles which have enabled
// reminders which are overdue and posts the reminder for them
func (d *dbClient) AutoSendReminders() (err error) {
	var rr []raffle

	if err = d.db.DB().
		Where("status = ? AND text_reminder_post = ? AND text_reminder_next_send < ?", raffleStatusActive, true, time.Now().UTC()).
		Find(&rr).
		Error; err != nil {
		return errors.Wrap(err, "fetching raffles to send reminders")
	}

	for _, r := range rr {
		if err = r.SendEvent(raffleMessageEventReminder, nil); err != nil {
			return errors.Wrapf(err, "sending reminder for raffle %d", r.ID)
		}
	}

	return nil
}

// AutoStart collects planned and overdue raffles and starts them
func (d *dbClient) AutoStart() (err error) {
	var rr []raffle

	if err = d.db.DB().
		Where("status = ? AND auto_start_at IS NOT NULL AND auto_start_at < ?", raffleStatusPlanned, time.Now().UTC()).
		Find(&rr).
		Error; err != nil {
		return errors.Wrap(err, "fetching raffles to start")
	}

	for _, r := range rr {
		if err = d.Start(r.ID); err != nil {
			return errors.Wrapf(err, "starting raffle %d", r.ID)
		}
	}

	return nil
}

// Clone duplicates a raffle into a new draft resetting some
// parameters into their default state
func (d *dbClient) Clone(raffleID uint64) error {
	raffle, err := d.Get(raffleID)
	if err != nil {
		return errors.Wrap(err, "getting raffle")
	}

	raffle.CloseAt = nil
	raffle.Entries = nil
	raffle.ID = 0
	raffle.Status = raffleStatusPlanned
	raffle.Title = strings.Join([]string{"Copy of", raffle.Title}, " ")

	return errors.Wrap(d.Create(raffle), "creating copy")
}

// Close marks the raffle as closed and removes it from the active
// raffle cache
func (d *dbClient) Close(raffleID uint64) error {
	r, err := d.Get(raffleID)
	if err != nil {
		return errors.Wrap(err, "getting raffle")
	}

	if err = d.db.DB().Model(&raffle{}).
		Where("id = ?", raffleID).
		Update("status", raffleStatusEnded).
		Error; err != nil {
		return errors.Wrap(err, "setting status closed")
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.activeRaffles, strings.Join([]string{r.Channel, r.Keyword}, "::"))

	return nil
}

// Create creates a new raffle. The record will be written to
// the database without modification and therefore need to be filled
// before calling this function
func (d *dbClient) Create(r raffle) error {
	return errors.Wrap(
		d.db.DB().Create(&r).Error,
		"creating database record",
	)
}

// Delete removes all entries for the given raffle and afterwards
// deletes the raffle itself
func (d *dbClient) Delete(raffleID uint64) (err error) {
	if err = d.db.DB().
		Where("raffle_id = ?", raffleID).
		Delete(&raffleEntry{}).
		Error; err != nil {
		return errors.Wrap(err, "deleting raffle entries")
	}

	return errors.Wrap(
		d.db.DB().
			Where("id = ?", raffleID).
			Delete(&raffle{}).Error,
		"creating database record",
	)
}

// Enter creates a new raffle entry. The entry will be written to
// the database without modification and therefore need to be filled
// before calling this function
func (d *dbClient) Enter(re raffleEntry) error {
	return errors.Wrap(
		d.db.DB().Create(&re).Error,
		"creating database record",
	)
}

// Get retrieves a raffle from the database
func (d *dbClient) Get(raffleID uint64) (out raffle, err error) {
	return out, errors.Wrap(
		d.db.DB().
			Where("raffles.id = ?", raffleID).
			Preload("Entries").
			First(&out).
			Error,
		"getting raffle from database",
	)
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
func (d *dbClient) List() (raffles []raffle, _ error) {
	return raffles, errors.Wrap(
		d.db.DB().Model(&raffle{}).
			Order("id DESC").
			Find(&raffles).
			Error,
		"updating column",
	)
}

// PatchNextReminderSend updates the time another reminder shall be
// sent for the given raffle ID. No other fields are modified
func (d *dbClient) PatchNextReminderSend(raffleID uint64, next time.Time) error {
	return errors.Wrap(
		d.db.DB().Model(&raffle{}).
			Where("id = ?", raffleID).
			Update("text_reminder_next_send", next).
			Error,
		"updating column",
	)
}

// PickWinner fetches the given raffle and picks a random winner
// based on entries and their multiplier
func (d *dbClient) PickWinner(raffleID uint64) error {
	r, err := d.Get(raffleID)
	if err != nil {
		return errors.Wrap(err, "getting raffle")
	}

	winner, err := pickWinnerFromRaffle(r)
	if err != nil {
		return errors.Wrap(err, "picking winner")
	}

	if err = d.db.DB().Model(&raffleEntry{}).
		Where("id = ?", winner.ID).
		Update("was_picked", true).
		Error; err != nil {
		return errors.Wrap(err, "updating winner")
	}

	fields := plugins.FieldCollectionFromData(map[string]any{
		"user_id": winner.UserID,
		"user":    winner.UserLogin,
		"winner":  winner,
	})

	return errors.Wrap(
		r.SendEvent(raffleMessageEventWin, fields),
		"sending win-message",
	)
}

// RedrawWinner marks the previous winner as redrawn (and therefore
// crossed out as winner in the interface) and picks a new one
func (d *dbClient) RedrawWinner(raffleID, winnerID uint64) error {
	if err := d.db.DB().Model(&raffleEntry{}).
		Where("id = ?", winnerID).
		Update("was_redrawn", true).
		Error; err != nil {
		return errors.Wrap(err, "updating previous winner")
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
		tmp     = map[string]uint64{}
	)

	if err := d.db.DB().
		Where("status = ?", raffleStatusActive).
		Find(&actives).
		Error; err != nil {
		return errors.Wrap(err, "fetching active raffles")
	}

	for _, r := range actives {
		tmp[strings.Join([]string{r.Channel, r.Keyword}, "::")] = r.ID
	}

	d.activeRaffles = tmp
	return nil
}

// Start fetches the given raffle, updates its CloseAt attribute
// in case it is not already set, sets the raffle to active, updates
// the raffle in the database and notes its channel/keyword combo
// into the activeRaffles cache for use with irc handling
func (d *dbClient) Start(raffleID uint64) error {
	r, err := d.Get(raffleID)
	if err != nil {
		return errors.Wrap(err, "getting specified raffle")
	}

	if r.CloseAt == nil {
		end := time.Now().UTC().Add(r.CloseAfter)
		r.CloseAt = &end
	}

	r.Status = raffleStatusActive
	if err = d.Update(r); err != nil {
		return errors.Wrap(err, "updating raffle")
	}

	// Store ID to active-raffle cache
	d.lock.Lock()
	defer d.lock.Unlock()
	d.activeRaffles[strings.Join([]string{r.Channel, r.Keyword}, "::")] = r.ID

	return errors.Wrap(
		r.SendEvent(raffleMessageEventReminder, nil),
		"sending first reminder",
	)
}

// Update stores the given raffle to the database. The ID within the
// raffle object must be set in order to update it. The object must
// be completely filled.
func (d *dbClient) Update(r raffle) error {
	return errors.Wrap(
		d.db.DB().
			Model(&raffle{}).
			Where("id = ?", r.ID).
			Updates(&r).
			Error,
		"updating raffle",
	)
}

// SendEvent processes the text template and sends the message if
// enabled given through the event
func (r raffle) SendEvent(evt raffleMessageEvent, fields *plugins.FieldCollection) (err error) {
	if fields == nil {
		fields = plugins.NewFieldCollection()
	}

	fields.Set("raffle", r) // Make raffle available to templating

	var sendTextTpl string

	switch evt {
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
			return errors.Wrap(err, "updating next reminder for raffle")
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
		return errors.Wrap(err, "formatting message to send")
	}

	return errors.Wrap(
		send(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				"#" + strings.TrimLeft(r.Channel, "#"),
				msg,
			},
		}),
		"sending message",
	)
}
