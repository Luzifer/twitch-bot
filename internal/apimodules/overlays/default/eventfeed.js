/**
 * @typedef {Object} Event
 * @property {number} eventId ID of the event as returned by the server
 * @property {Object|undefined} extraData Any additional data specific to this event type
 * @property {string} filterKey Event-Type key
 * @property {string|undefined} originId ID from the Twitch server for de-duplication
 * @property {string|function|undefined} subtext Additional text, usually user-message
 * @property {string|undefined} text Descriptive text of the event
 * @property {Date} time The moment the event occurred
 * @property {string} title The title of the event
 * @property {boolean} hasReplay Whether the replay button should be shown
 * @property {boolean} isMeta Whether not to display event in frontend
 */

import { customFilters, customHandler } from './eventfeed.custom.js'
import { createApp } from 'https://cdn.jsdelivr.net/npm/vue@3.4/dist/vue.esm-browser.prod.js'
import dayjs from 'https://cdn.jsdelivr.net/npm/dayjs@1.11/+esm'
import dayjsLocalizedFormat from 'https://cdn.jsdelivr.net/npm/dayjs@1.11/plugin/localizedFormat.js/+esm'
import dayjsRelativeTime from 'https://cdn.jsdelivr.net/npm/dayjs@1.11/plugin/relativeTime.js/+esm'
import EventClient from './eventclient.mjs'

const STORAGE_KEY = 'io.luzifer.eventfeed'

const defaultFilters = {
  adbreak: { name: 'Adbreaks', visible: true },
  ban: { name: 'Bans / Timeouts', visible: true },
  bits: { name: 'Bits', visible: true },
  channelpoint: { name: 'Channel-Points', visible: true },
  donation: { name: 'Donations', visible: true },
  follow: { name: 'Follows', visible: true },
  hypetrain: { name: 'Hypetrains', visible: true },
  pollEnd: { name: 'Poll-Summary', visible: true },
  raid: { name: 'Raids', visible: true },
  shoutout: { name: 'Shoutouts', visible: true },
  streamOffline: { name: 'Stream-Offline', visible: true },
  streamUpdate: { name: 'Stream-Update', visible: true },
  subs: { name: 'Subs', visible: true },
  watchStreak: { name: 'Watchstreaks', visible: true },
}

const userAnonSubgifter = 'ananonymousgifter'
const userAnonCheerer = 'ananonymouscheerer'

const app = createApp({
  computed: {
    filterCount() {
      const filters = Object.values(this.filters)
      return `${filters.filter(f => f.visible).length} / ${filters.length}`
    },

    filters() {
      return Object.fromEntries(Object.entries({
        ...defaultFilters,
        ...customFilters(),
        ...this.storedData.filters || {},
      })
        .filter(e => Object.keys(defaultFilters).includes(e[0]) || Object.keys(customFilters()).includes(e[0]))
        .sort((a, b) => a[1].name.localeCompare(b[1].name)))
    },

    hypetrain() {
      const evts = [...this.events]
        .filter(evt => evt.filterKey === 'hypetrain')
        .sort((b, a) => a.time.getTime() - b.time.getTime())

      if (evts.length < 1) {
        return {
          active: false,
        }
      }

      return evts[0].extraData
    },

    recentEvents() {
      return [...this.events]
        .filter(evt => !evt.isMeta)
        .filter(evt => this.filters[evt.filterKey]?.visible !== false)
        .filter(evt => !this.knownMultiGiftIDs.includes(evt.originId))
        .sort((b, a) => a.time.getTime() - b.time.getTime())
    },

    sortedStats() {
      const evts = [...this.events]
        .filter(evt => evt.time.getTime() > this.streamOfflineTime.getTime())


      return [
        {
          icon: 'fas fa-gem',
          key: 'bits',
          value: evts
            .filter(evt => evt.filterKey === 'bits')
            .reduce((sum, evt) => sum + evt.extraData.bits, 0),
        },
        {
          icon: 'fas fa-circle-dollar-to-slot',
          key: 'donation',
          value: evts
            .filter(evt => evt.filterKey === 'donation')
            .reduce((sum, evt) => sum + evt.extraData.amount, 0)
            .toFixed(2),
        },
        {
          icon: 'fas fa-heart',
          key: 'follow',
          value: evts
            .filter(evt => evt.filterKey === 'follow')
            .length,
        },
        {
          icon: 'fas fa-parachute-box',
          key: 'raid',
          value: evts
            .filter(evt => evt.filterKey === 'raid')
            .length,
        },
        {
          icon: 'fas fa-star',
          key: 'sub',
          value: evts
            .filter(evt => evt.filterKey === 'subs')
            .filter(evt => !this.knownMultiGiftIDs.includes(evt.originId))
            .reduce((sum, evt) => sum + evt.extraData.count, 0),
        },
      ]
    },
  },

  created() {
    window.setInterval(() => {
      this.now = new Date()
    }, 60000)

    this.eventClient = new EventClient({
      handlers: {
        adbreak_begin: ({ event_id, fields, time }) => this.handleAdBreak(event_id, fields, time),
        ban: ({ event_id, fields, time }) => this.handleBan(event_id, fields, time),
        bits: ({ event_id, fields, time }) => this.handleBits(event_id, fields, time),
        category_update: ({ event_id, fields, time }) => this.handleCategoryUpdate(event_id, fields, time),
        channelpoint_redeem: ({ event_id, fields, time }) => this.handleChannelPoints(event_id, fields, time),
        custom: eventobj => this.handleCustom(eventobj),
        follow: ({ event_id, fields, time }) => this.handleFollow(event_id, fields, time),
        hypetrain_begin: ({ event_id, fields, time }) => this.handleHypetrain(event_id, fields, time, 'start'),
        hypetrain_end: ({ event_id, fields, time }) => this.handleHypetrain(event_id, fields, time, 'end'),
        hypetrain_progress: ({ event_id, fields, time }) => this.handleHypetrain(event_id, fields, time, 'progress'),
        kofi_donation: ({ event_id, fields, time }) => this.handleKoFiDonation(event_id, fields, time),
        poll_end: ({ event_id, fields, time }) => this.handlePollEnd(event_id, fields, time),
        raid: ({ event_id, fields, time }) => this.handleRaid(event_id, fields, time),
        resub: ({ event_id, fields, reason, time, type }) => this.handleSub(type, event_id, fields, time, reason),
        shoutout_created: ({ event_id, fields, time }) => this.handleShoutoutCreated(event_id, fields, time),
        shoutout_received: ({ event_id, fields, time }) => this.handleShoutoutReceived(event_id, fields, time),
        stream_offline: ({ event_id, time }) => this.handleStreamOffline(event_id, time),
        sub: ({ event_id, fields, time, type }) => this.handleSub(type, event_id, fields, time),
        subgift: ({ event_id, fields, time, type }) => this.handleSubgift(type, event_id, fields, time),
        submysterygift: ({ event_id, fields, time, type }) => this.handleSubgift(type, event_id, fields, time),
        timeout: ({ event_id, fields, time }) => this.handleTimeout(event_id, fields, time),
        title_update: ({ event_id, fields, time }) => this.handleTitleUpdate(event_id, fields, time),
        watch_streak: ({ event_id, fields, time }) => this.handleWatchStreak(event_id, fields, time),
      },

      maxReplayAge: 168,
      replay: true,
    })

    this.storageLoad()
    window.addEventListener('storage', ev => {
      if (ev.key !== this.storageKey()) {
        return
      }

      // Our key has been changed, reload stored data
      this.storageLoad()
    })
  },

  data() {
    return {
      eventClient: null,
      events: [],
      now: new Date(),
      storedData: {},

      // Workaround for Twitch not sending hypetrain progress in end-event
      // eslint-disable-next-line sort-keys
      hypetrainProgress: 0,
      knownMultiGiftIDs: [],
      streamOfflineTime: new Date(0),
      subgiftRecipients: {},
    }
  },

  methods: {
    /**
     * @param {Event} event
     */
    addEvent(event) {
      if (!event.eventId || !event.filterKey || !event.time || !event.title) {
        throw new Error(`Event missing fields: ${event}`)
      }

      this.events = [
        ...this.events.filter(evt => evt.eventId !== event.eventId),
        event,
      ]
    },

    eventClass(event) {
      const classes = ['border-event', 'list-group-item']

      if (this.storedData.readDate && this.storedData.readDate > event.time.getTime()) {
        classes.push('disabled')
      }

      if (event.filterKey) {
        classes.push(`event-${event.filterKey}`)
      }

      return classes.join(' ')
    },

    handleAdBreak(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'adbreak',
        icon: 'fas fa-rectangle-ad text-warning',
        text: `${data.duration}s ad-break is now running`,
        time: time ? new Date(time) : null,
        title: 'Ad-Break started',
      })
    },

    handleBan(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'ban',
        icon: 'fas fa-ban',
        time: new Date(time),
        title: `${data.target_name} has been banned`,
      })
    },

    handleBits(eventId, data, time) {
      const from = data.user === userAnonCheerer ? 'Someone' : data.user

      this.addEvent({
        eventId,
        extraData: { bits: data.bits },
        filterKey: 'bits',
        hasReplay: true,
        icon: 'fas fa-gem',
        subtext: data.message,
        text: `${from} just spent ${data.bits} Bits`,
        time: time ? new Date(time) : null,
        title: 'Bits donated',
      })
    },

    handleCategoryUpdate(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'streamUpdate',
        icon: 'fas fa-gamepad',
        text: data.category,
        time: new Date(time),
        title: 'Category updated',
      })
    },

    handleChannelPoints(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'channelpoint',
        hasReplay: true,
        icon: 'fas fa-diamond',
        subtext: data.user_input,
        text: `${data.user} redeemed "${data.reward_title}"`,
        time: new Date(time),
        title: 'Reward Redeemed',
      })
    },

    handleCustom(eventObj) {
      const evt = customHandler(eventObj)
      if (evt !== null) {
        this.addEvent(evt)
      }
    },

    handleFollow(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'follow',
        hasReplay: true,
        icon: 'fas fa-user',
        text: `${data.user} just followed`,
        time: new Date(time),
        title: 'New Follower',
      })
    },

    handleHypetrain(eventId, data, time, phase) {
      const evt = {
        eventId,
        extraData: {
          active: phase !== 'end',
          level: data.level,
          progress: data.levelProgress || this.hypetrainProgress,
        },

        filterKey: 'hypetrain',
        icon: 'fas fa-train',
        time: new Date(time),
      }

      this.hypetrainProgress = evt.extraData.progress

      switch (phase) {
      case 'start':
        this.addEvent({
          ...evt,
          text: `A hypetrain started on ${(data.levelProgress * 100).toFixed(0)}% towards level ${data.level}`,
          title: 'Hypetrain started',
        })
        break

      case 'progress':
        this.addEvent({
          ...evt,
          isMeta: true,
          title: 'Hypetrain progressed',
        })
        break

      case 'end':
        this.addEvent({
          ...evt,
          text: `A hypetrain ended on ${(this.hypetrainProgress * 100).toFixed(0)}% towards level ${data.level}`,
          title: 'Hypetrain ended',
        })
        break
      }
    },

    handleKoFiDonation(eventId, data, time) {
      let text
      if (data.isSubscription && data.isFirstSubPayment) {
        text = `${data.from} just started a monthly subscription of ${Number(data.amount).toFixed(2)}€`
      } else if (data.isSubscription && !data.isFirstSubPayment) {
        text = `${data.from} continued their monthly subscription of ${Number(data.amount).toFixed(2)}€`
      } else {
        text = `${data.from} just donated ${Number(data.amount).toFixed(2)}€`
      }

      this.addEvent({
        eventId,
        extraData: { amount: Number(data.amount) },
        filterKey: 'donation',
        icon: 'fas fa-circle-dollar-to-slot',
        subtext: data.message ? data.message : undefined,
        text,
        time: new Date(time),
        title: 'Ko-fi Donation received',
      })
    },

    handlePollEnd(eventId, data, time) {
      if (data.poll.status === 'archived') {
        return
      }

      this.addEvent({
        eventId,
        filterKey: 'pollEnd',
        icon: 'fas fa-square-poll-vertical',
        subtext: data.poll.choices.map(choice => `${choice.title} (${choice.votes})`).join(' | '),
        text: data.poll.title,
        time: new Date(time),
        title: `Poll Ended (${data.poll.status})`,
      })
    },

    handleRaid(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'raid',
        hasReplay: true,
        icon: 'fas fa-parachute-box',
        soundUrl: '/public/fanfare.webm',
        text: `${data.from} just raided with ${data.viewercount} raiders`,
        time: new Date(time),
        title: 'Incoming raid',
      })
    },

    handleShoutoutCreated(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'shoutout',
        icon: 'fas fa-bullhorn',
        text: `We gave a shoutout for ${data.to} to ${data.viewers} viewers`,
        time: new Date(time),
        title: 'Shoutout created',
      })
    },

    handleShoutoutReceived(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'shoutout',
        icon: 'fas fa-bullhorn',
        text: `${data.from} just gave us a shoutout to ${data.viewers} viewers`,
        time: new Date(time),
        title: 'Shoutout received',
      })
    },

    handleStreamOffline(eventId, time) {
      this.addEvent({
        eventId,
        filterKey: 'streamOffline',
        icon: 'fas fa-clapperboard text-danger',
        time: new Date(time),
        title: 'Stream Offline',
      })

      this.streamOfflineTime = new Date(time)
    },

    handleSub(evt, eventId, data, time) {
      const text = evt === 'resub' ? `resubscribed for the ${data.subscribed_months}. time` : 'subscribed'
      const tier = data.plan === 'Prime' ? 'P' : `T${Number(data.plan) / 1000}`
      const title = evt === 'resub' ? `Resub (${tier})` : `New Sub (${tier})`
      this.addEvent({
        eventId,
        extraData: { count: 1 },
        filterKey: 'subs',
        hasReplay: true,
        icon: 'fas fa-star',
        subtext: data.message,
        text: `${data.user} just ${text} (${tier})`,
        time: new Date(time),
        title,
      })
    },

    handleSubgift(evt, eventId, data, time) {
      const from = data.user === userAnonSubgifter ? 'ANON' : data.from

      const tier = data.plan === 'Prime' ? 'Prime' : `Tier ${Number(data.plan) / 1000}`

      if (evt === 'submysterygift') {
        this.addEvent({
          eventId,
          extraData: { count: data.number },
          filterKey: 'subs',
          hasReplay: true,
          icon: 'fas fa-gift',
          subtext: () => this.subgiftRecipients[data.origin_id] ? `To: ${this.subgiftRecipients[data.origin_id].join(', ')}` : undefined,
          text: `${from} just gifted ${data.number} subs`,
          time: time ? new Date(time) : null,
          title: `Subs gifted (${tier})`,
          variant: 'warning',
        })

        this.knownMultiGiftIDs.push(data.origin_id)
        return
      }

      if (data.origin_id) {
        this.subgiftRecipients[data.origin_id] = [
          ...this.subgiftRecipients[data.origin_id] || [],
          data.to,
        ].sort((a, b) => a.localeCompare(b))
      }

      this.addEvent({
        eventId,
        extraData: { count: 1 },
        filterKey: 'subs',
        hasReplay: true,
        icon: 'fas fa-gift',
        originId: data.origin_id,
        text: `${from} just gifted ${data.to} a sub`,
        time: time ? new Date(time) : null,
        title: `Sub gifted (${tier})`,
        variant: 'warning',
      })
    },

    handleTimeout(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'ban',
        icon: 'fas fa-ban',
        time: new Date(time),
        title: `${data.target_name} has been timed out for ${data.seconds}s`,
      })
    },

    handleTitleUpdate(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'streamUpdate',
        icon: 'fas fa-heading',
        text: data.title,
        time: new Date(time),
        title: 'Title updated',
      })
    },

    handleWatchStreak(eventId, data, time) {
      this.addEvent({
        eventId,
        filterKey: 'watchStreak',
        icon: 'fas fa-circle-info',
        subtext: data.message,
        text: `${data.user} watched ${data.streak} consecutive streams`,
        time: new Date(time),
        title: 'Watch-Streak shared',
      })
    },

    markRead() {
      this.storedData.readDate = new Date().getTime()
      this.storageSave()
    },

    repeatEvent(eventId) {
      return this.eventClient.replayEvent(eventId)
    },

    resolveSubtext(subtext) {
      if (typeof subtext === 'function') {
        return subtext()
      }

      return subtext
    },

    storageKey() {
      const channel = this.eventClient.paramOptionFallback('channel').replace(/^#*/, '')
      return [STORAGE_KEY, channel].join('.')
    },

    storageLoad() {
      this.storedData = {
        // Default values
        filters: {},
        readDate: 0,

        // Stored data
        ...JSON.parse(window.localStorage.getItem(this.storageKey()) || '{}'),
      }
    },

    storageSave() {
      window.localStorage.setItem(this.storageKey(), JSON.stringify(this.storedData))
    },

    timeDisplay(time) {
      return dayjs(time).format('llll')
    },

    timeSince(time) {
      return dayjs(time).from(this.now)
    },

    toggleFilterVisibility(filter) {
      if (!this.storedData.filters[filter]) {
        this.storedData.filters[filter] = this.filters[filter]
      }

      this.storedData.filters[filter].visible = !this.storedData.filters[filter].visible
      this.storageSave()
    },
  },

  name: 'EventFeed',
})

dayjs.extend(dayjsLocalizedFormat)
dayjs.extend(dayjsRelativeTime)

app.mount('#app')
