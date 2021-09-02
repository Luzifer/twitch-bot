package main

import (
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type (
	twitchChannelState struct {
		Category string
		IsLive   bool
		Title    string
	}

	twitchWatcher struct {
		ChannelStatus map[string]*twitchChannelState

		lock sync.RWMutex
	}
)

func (t twitchChannelState) Equals(c twitchChannelState) bool {
	return t.Category == c.Category &&
		t.IsLive == c.IsLive &&
		t.Title == c.Title
}

func newTwitchWatcher() *twitchWatcher {
	return &twitchWatcher{
		ChannelStatus: make(map[string]*twitchChannelState),
	}
}

func (r *twitchWatcher) AddChannel(channel string) error {
	r.lock.RLock()
	if _, ok := r.ChannelStatus[channel]; ok {
		r.lock.RUnlock()
		return nil
	}
	r.lock.RUnlock()

	return r.updateChannelFromAPI(channel, false)
}

func (r *twitchWatcher) Check() {
	var channels []string
	r.lock.RLock()
	for c := range r.ChannelStatus {
		channels = append(channels, c)
	}
	r.lock.RUnlock()

	for _, ch := range channels {
		if err := r.updateChannelFromAPI(ch, true); err != nil {
			log.WithError(err).WithField("channel", ch).Error("Unable to update channel status")
		}
	}
}

func (r *twitchWatcher) RemoveChannel(channel string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.ChannelStatus, channel)
	return nil
}

func (r *twitchWatcher) updateChannelFromAPI(channel string, sendUpdate bool) error {
	var (
		err    error
		status twitchChannelState
	)

	status.IsLive, err = twitchClient.HasLiveStream(channel)
	if err != nil {
		return errors.Wrap(err, "getting live status")
	}

	status.Category, status.Title, err = twitchClient.GetRecentStreamInfo(channel)
	if err != nil {
		return errors.Wrap(err, "getting stream info")
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	if r.ChannelStatus[channel] != nil && r.ChannelStatus[channel].Equals(status) {
		return nil
	}

	if sendUpdate && r.ChannelStatus[channel] != nil {
		if r.ChannelStatus[channel].Category != status.Category {
			log.WithFields(log.Fields{
				"channel":  channel,
				"category": status.Category,
			}).Debug("Twitch metadata changed")
			go handleMessage(nil, nil, eventTypeTwitchCategoryUpdate, map[string]interface{}{
				"channel":  channel,
				"category": status.Category,
			})
		}

		if r.ChannelStatus[channel].Title != status.Title {
			log.WithFields(log.Fields{
				"channel": channel,
				"title":   status.Title,
			}).Debug("Twitch metadata changed")
			go handleMessage(nil, nil, eventTypeTwitchTitleUpdate, map[string]interface{}{
				"channel": channel,
				"title":   status.Title,
			})
		}

		if r.ChannelStatus[channel].IsLive != status.IsLive {
			log.WithFields(log.Fields{
				"channel": channel,
				"isLive":  status.IsLive,
			}).Debug("Twitch metadata changed")

			evt := eventTypeTwitchStreamOnline
			if !status.IsLive {
				evt = eventTypeTwitchStreamOffline
			}

			go handleMessage(nil, nil, evt, map[string]interface{}{
				"channel": channel,
			})
		}
	}

	r.ChannelStatus[channel] = &status
	return nil
}
