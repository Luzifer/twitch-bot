<overlay-title>Event-Feed</overlay-title>

<template>
  <div class="container-fluid py-3">
    <div class="row">
      <div class="col">
        <!-- Stream-Summary -->
        <div class="card mb-3">
          <div class="card-body">
            <div class="d-flex align-items-center justify-content-between">
              <span
                v-for="item in sortedStats"
                :key="item.key"
                class="me-2 d-inline-flex align-items-center"
              >
                <i :class="`fa-fw ${item.icon}`" />
                <span class="badge rounded-pill text-bg-primary ms-1">
                  {{ item.value }}
                </span>
              </span>
            </div>
          </div>
        </div>

        <!-- Event-List -->
        <div class="card">
          <div class="card-header d-flex justify-content-between align-items-center">
            Recent events
            <div class="btn-group btn-group-sm">
              <div class="btn-group btn-group-sm">
                <button
                  type="button"
                  class="btn btn-secondary dropdown-toggle"
                  data-bs-toggle="dropdown"
                  aria-expanded="false"
                >
                  <i class="fas fa-filter fa-fw me-1" />
                  Filters ({{ filterCount }})
                </button>
                <ul class="dropdown-menu dropdown-menu-end">
                  <li
                    v-for="(filter, filterKey) in filters"
                    :key="filterKey"
                  >
                    <a
                      :class="{'dropdown-item': true, 'active': filter.visible}"
                      href="#"
                      @click.prevent="toggleFilterVisibility(filterKey)"
                    >
                      {{ filter.name }}
                    </a>
                  </li>
                </ul>
              </div>

              <button
                class="btn btn-secondary"
                @click="markRead"
              >
                <i class="fas fa-eye fa-fw me-1" />
                Mark read
              </button>
            </div>
          </div>

          <div class="list-group list-group-flush">
            <!-- Active Hypetrain pin -->
            <div
              v-if="hypetrain.active"
              class="list-group-item"
            >
              <div class="d-flex w-100 align-items-center">
                <h5 class="mb-0">
                  <i :class="`fas fa-train fa-fw me-2`" />
                  Hypetrain in progress towards Level {{ hypetrain.level }}…
                </h5>
              </div>

              <div class="progress my-3">
                <div
                  class="progress-bar progress-bar-striped"
                  :style="`width: ${(hypetrain.progress * 100).toFixed(2)}%`"
                />
              </div>
            </div>

            <!-- Event-Item -->
            <div
              v-for="event in recentEvents"
              :key="event.time.getTime()"
              :class="eventClass(event)"
            >
              <div class="d-flex w-100 align-items-center">
                <h5 class="mb-0 me-auto">
                  <i :class="`${event.icon} fa-fw me-2`" /> {{ event.title }}
                </h5>
                <button
                  v-if="event.hasReplay"
                  class="btn btn-sm me-1"
                  title="Re-Play Event"
                  @click="repeatEvent(event.eventId)"
                >
                  <i class="fas fa-share fa-fw" />
                </button>
                <small :title="timeDisplay(event.time)">
                  {{ timeSince(event.time) }}
                </small>
              </div>

              <div
                v-if="event.text"
                class="d-flex my-1 w-100 justify-content-between align-items-start premono"
              >
                {{ event.text }}
              </div>
              <p
                v-if="resolveSubtext(event.subtext)"
                class="mb-1"
              >
                <small>
                  <span class="premono">{{ resolveSubtext(event.subtext) }}</span>
                </small>
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.min.css'

import type { AdbreakBeginFields, BanFields, BitsFields, CategoryUpdateFields, ChannelpointRedeemFields, CustomSocketMessage, FollowFields, HypetrainBeginFields, HypetrainEndFields, HypetrainProgressFields, KofiDonationFields, ModeratorAddFields, ModeratorRemoveFields, PollEndFields, PrimepaidupgradeFields, RaidFields, ResubFields, ShoutoutCreatedFields, ShoutoutReceivedFields, SubFields, SubgiftFields, SubmysterygiftFields, TimeoutFields, TitleUpdateFields, VipAddFields, VipRemoveFields, WatchStreakFields } from './eventTypes.js'
import { createApp, defineComponent } from 'vue'
import { customFilters, customHandler } from './eventfeed.custom.js'
import { dom, library } from '@fortawesome/fontawesome-svg-core'
import dayjs from 'dayjs'
import dayjsLocalizedFormat from 'dayjs/plugin/localizedFormat.js'
import dayjsRelativeTime from 'dayjs/plugin/relativeTime.js'
import EventClient from './eventclient.js'
import { fas } from '@fortawesome/free-solid-svg-icons'

export type Event = {
  eventId: string
  extraData?: Record<string, any>
  filterKey: string
  hasReplay?: boolean
  icon: string
  isMeta?: boolean
  originId?: string
  subtext?: string | (() => string | undefined)
  text?: string
  time: Date
  title: string
}

export type Filter = {
  name: string
  visible: boolean
}

type StoredData = {
  filters: Record<string, Filter>
  readDate: number
}

const STORAGE_KEY = 'io.luzifer.eventfeed'

library.add(fas)
dom.watch()

const defaultFilters: Record<string, Filter> = {
  adbreak: { name: 'Adbreaks', visible: true },
  ban: { name: 'Bans / Timeouts', visible: true },
  bits: { name: 'Bits', visible: true },
  channelpoint: { name: 'Channel-Points', visible: true },
  donation: { name: 'Donations', visible: true },
  follow: { name: 'Follows', visible: true },
  hypetrain: { name: 'Hypetrains', visible: true },
  pollEnd: { name: 'Poll-Summary', visible: true },
  raid: { name: 'Raids', visible: true },
  roleChange: { name: 'Role Change', visible: true },
  shoutout: { name: 'Shoutouts', visible: true },
  streamOffline: { name: 'Stream-Offline', visible: true },
  streamUpdate: { name: 'Stream-Update', visible: true },
  subs: { name: 'Subs', visible: true },
  watchStreak: { name: 'Watchstreaks', visible: true },
}

const userAnonSubgifter = 'ananonymousgifter'
const userAnonCheerer = 'ananonymouscheerer'

const component = defineComponent({
  computed: {
    filterCount() {
      const filters = Object.values(this.filters as Record<string, Filter>)
      return `${filters.filter(f => f.visible).length} / ${filters.length}`
    },

    filters(): Record<string, Filter> {
      return Object.fromEntries(Object.entries({
        ...defaultFilters,
        ...customFilters(),
        ...this.storedData.filters || {},
      })
        .filter(e => Object.keys(defaultFilters).includes(e[0]) || Object.keys(customFilters()).includes(e[0]))
        .sort((a, b) => (a[1] as Filter).name.localeCompare((b[1] as Filter).name))) as Record<string, Filter>
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

      return evts[0].extraData!
    },

    recentEvents() {
      return [...this.events]
        .filter(evt => !evt.isMeta)
        .filter(evt => this.filters[evt.filterKey]?.visible !== false)
        .filter(evt => !evt.originId || !this.knownMultiGiftIDs.includes(evt.originId))
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
            .reduce((sum, evt) => sum + evt.extraData!.bits, 0),
        },
        {
          icon: 'fas fa-circle-dollar-to-slot',
          key: 'donation',
          value: evts
            .filter(evt => evt.filterKey === 'donation')
            .reduce((sum, evt) => sum + evt.extraData!.amount, 0)
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
            .filter(evt => !evt.originId || !this.knownMultiGiftIDs.includes(evt.originId))
            .reduce((sum, evt) => sum + evt.extraData!.count, 0),
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
        moderator_add: ({ event_id, fields, time }) => this.handleRoleChange(event_id, fields, time, 'Moderator', true),
        moderator_remove: ({ event_id, fields, time }) => this.handleRoleChange(event_id, fields, time, 'Moderator', false),
        poll_end: ({ event_id, fields, time }) => this.handlePollEnd(event_id, fields, time),
        primepaidupgrade: ({ event_id, fields, time }) => this.handlePrimePaidUpgrade(event_id, fields, time),
        raid: ({ event_id, fields, time }) => this.handleRaid(event_id, fields, time),
        resub: ({ event_id, fields, time, type }) => this.handleSub(type, event_id, fields, time),
        shoutout_created: ({ event_id, fields, time }) => this.handleShoutoutCreated(event_id, fields, time),
        shoutout_received: ({ event_id, fields, time }) => this.handleShoutoutReceived(event_id, fields, time),
        stream_offline: ({ event_id, time }) => this.handleStreamOffline(event_id, time),
        sub: ({ event_id, fields, time, type }) => this.handleSub(type, event_id, fields, time),
        subgift: ({ event_id, fields, time, type }) => this.handleSubgift(type, event_id, fields, time),
        submysterygift: ({ event_id, fields, time, type }) => this.handleSubgift(type, event_id, fields, time),
        timeout: ({ event_id, fields, time }) => this.handleTimeout(event_id, fields, time),
        title_update: ({ event_id, fields, time }) => this.handleTitleUpdate(event_id, fields, time),
        vip_add: ({ event_id, fields, time }) => this.handleRoleChange(event_id, fields, time, 'VIP', true),
        vip_remove: ({ event_id, fields, time }) => this.handleRoleChange(event_id, fields, time, 'VIP', false),
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
      eventClient: null as null | EventClient,
      events: [] as Event[],
      now: new Date(),
      storedData: {
        filters: {},
        readDate: 0,
      } as StoredData,

      // Workaround for Twitch not sending hypetrain progress in end-event
      // eslint-disable-next-line sort-keys
      hypetrainProgress: 0,
      knownMultiGiftIDs: [] as string[],
      streamOfflineTime: new Date(0),
      subgiftRecipients: {} as Record<string, string[]>,
    }
  },

  methods: {
    /**
     * @param {Event} event
     */
    addEvent(event: Event) {
      if (!event.eventId || !event.filterKey || !event.time || !event.title) {
        throw new Error(`Event missing fields: ${event}`)
      }

      this.events = [
        ...this.events.filter(evt => evt.eventId !== event.eventId),
        event,
      ]
    },

    eventClass(event: Event) {
      const classes = ['border-event', 'list-group-item']

      if (this.storedData.readDate && this.storedData.readDate > event.time.getTime()) {
        classes.push('disabled')
      }

      if (event.filterKey) {
        classes.push(`event-${event.filterKey}`)
      }

      return classes.join(' ')
    },

    handleAdBreak(eventId: string, data: AdbreakBeginFields, time: Date) {
      this.addEvent({
        eventId,
        filterKey: 'adbreak',
        icon: 'fas fa-rectangle-ad text-warning',
        text: `${data.duration}s ad-break is now running`,
        time: new Date(time),
        title: 'Ad-Break started',
      })
    },

    handleBan(eventId: string, data: BanFields, time: Date) {
      this.addEvent({
        eventId,
        filterKey: 'ban',
        icon: 'fas fa-ban',
        time: new Date(time),
        title: `${data.target_name} has been banned`,
      })
    },

    handleBits(eventId: string, data: BitsFields, time: Date) {
      const from = data.user === userAnonCheerer ? 'Someone' : data.user

      this.addEvent({
        eventId,
        extraData: { bits: data.bits },
        filterKey: 'bits',
        hasReplay: true,
        icon: 'fas fa-gem',
        subtext: data.message,
        text: `${from} just spent ${data.bits} Bits`,
        time: new Date(time),
        title: 'Bits donated',
      })
    },

    handleCategoryUpdate(eventId: string, data: CategoryUpdateFields, time: Date) {
      this.addEvent({
        eventId,
        filterKey: 'streamUpdate',
        icon: 'fas fa-gamepad',
        text: data.category,
        time: new Date(time),
        title: 'Category updated',
      })
    },

    handleChannelPoints(eventId: string, data: ChannelpointRedeemFields, time: Date) {
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

    handleCustom(eventObj: CustomSocketMessage) {
      const evt = customHandler(eventObj)
      if (evt !== null) {
        this.addEvent(evt)
      }
    },

    handleFollow(eventId: string, data: FollowFields, time: Date) {
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

    handleHypetrain(
      eventId: string,
      data: HypetrainBeginFields | HypetrainEndFields | HypetrainProgressFields,
      time: Date,
      phase: 'start' | 'end' | 'progress',
    ) {
      const progress = 'levelProgress' in data
        ? data.levelProgress
        : this.hypetrainProgress

      const evt: Event = {
        eventId,
        extraData: {
          active: phase !== 'end',
          level: data.level,
          progress,
        },

        filterKey: 'hypetrain',
        icon: 'fas fa-train',
        time: new Date(time),
        title: '',
      }

      this.hypetrainProgress = evt.extraData!.progress

      switch (phase) {
      case 'start':
        this.addEvent({
          ...evt,
          text: `A hypetrain started on ${(progress * 100).toFixed(0)}% towards level ${data.level}`,
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

    handleKoFiDonation(eventId: string, data: KofiDonationFields, time: Date) {
      let text
      if (data.isSubscription && data.isFirstSubPayment) {
        text = `${data.from} just started a monthly subscription of ${Number(data.amount).toFixed(2)} ${data.currency}`
      } else if (data.isSubscription && !data.isFirstSubPayment) {
        text = `${data.from} continued their monthly subscription of ${Number(data.amount).toFixed(2)} ${data.currency}`
      } else {
        text = `${data.from} just donated ${Number(data.amount).toFixed(2)} ${data.currency}`
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

    handlePollEnd(eventId: string, data: PollEndFields, time: Date) {
      if (data.status === 'archived') {
        return
      }

      // Map into stub-type
      const poll = data.poll as {
        choices: {
          title: string
          votes: number
        }[]
      }

      this.addEvent({
        eventId,
        filterKey: 'pollEnd',
        icon: 'fas fa-square-poll-vertical',
        subtext: poll.choices.map(choice => `${choice.title} (${choice.votes})`).join(' | '),
        text: data.title,
        time: new Date(time),
        title: `Poll Ended (${data.status})`,
      })
    },

    handlePrimePaidUpgrade(eventId: string, data: PrimepaidupgradeFields, time: Date) {
      this.addEvent({
        eventId,
        extraData: { count: 0 }, // this is not a sub itself, just a change for the future but related to subs
        filterKey: 'subs',
        icon: 'fas fa-circle-up',
        text: `${data.user} upgraded from Prime to paid sub`,
        time: new Date(time),
        title: 'Prime to Paid Upgrade',
      })
    },

    handleRaid(eventId: string, data: RaidFields, time: Date) {
      this.addEvent({
        eventId,
        filterKey: 'raid',
        hasReplay: true,
        icon: 'fas fa-parachute-box',
        text: `${data.from} just raided with ${data.viewercount} raiders`,
        time: new Date(time),
        title: 'Incoming raid',
      })
    },

    handleRoleChange(eventId: string, data: VipAddFields | VipRemoveFields | ModeratorAddFields | ModeratorRemoveFields, time: Date, role: string, added: boolean = true) {
      const action = added ? 'added to' : 'removed from'

      this.addEvent({
        eventId,
        filterKey: 'roleChange',
        icon: 'fas fa-user-tag',
        text: `${data.user} got ${action} role ${role}`,
        time: new Date(time),
        title: 'User-role changed',
      })
    },

    handleShoutoutCreated(eventId: string, data: ShoutoutCreatedFields, time: Date) {
      this.addEvent({
        eventId,
        filterKey: 'shoutout',
        icon: 'fas fa-bullhorn',
        text: `We gave a shoutout for ${data.to} to ${data.viewers} viewers`,
        time: new Date(time),
        title: 'Shoutout created',
      })
    },

    handleShoutoutReceived(eventId: string, data: ShoutoutReceivedFields, time: Date) {
      this.addEvent({
        eventId,
        filterKey: 'shoutout',
        icon: 'fas fa-bullhorn',
        text: `${data.from} just gave us a shoutout to ${data.viewers} viewers`,
        time: new Date(time),
        title: 'Shoutout received',
      })
    },

    handleStreamOffline(eventId: string, time: Date) {
      this.addEvent({
        eventId,
        filterKey: 'streamOffline',
        icon: 'fas fa-clapperboard text-danger',
        time: new Date(time),
        title: 'Stream Offline',
      })

      this.streamOfflineTime = new Date(time)
    },

    handleSub(evt: string, eventId: string, data: ResubFields | SubFields, time: Date) {
      const text = evt === 'resub'
        ? `resubscribed for the ${(data as ResubFields).subscribed_months}. time`
        : 'subscribed'
      const tier = data.plan === 'Prime' ? 'P' : `T${Number(data.plan) / 1000}`
      const title = evt === 'resub' ? `Resub (${tier})` : `New Sub (${tier})`
      this.addEvent({
        eventId,
        extraData: { count: 1 },
        filterKey: 'subs',
        hasReplay: true,
        icon: 'fas fa-star',
        subtext: evt === 'resub' ? (data as ResubFields).message : undefined,
        text: `${data.user} just ${text} (${tier})`,
        time: new Date(time),
        title,
      })
    },

    handleSubgift(
      evt: 'subgift' | 'submysterygift',
      eventId: string,
      data: SubgiftFields | SubmysterygiftFields,
      time: Date,
    ) {
      const from = data.user === userAnonSubgifter ? 'ANON' : data.from

      const tier = data.plan === 'Prime' ? 'Prime' : `Tier ${Number(data.plan) / 1000}`

      if (evt === 'submysterygift') {
        this.addEvent({
          eventId,
          extraData: { count: (data as SubmysterygiftFields).number },
          filterKey: 'subs',
          hasReplay: true,
          icon: 'fas fa-gift',
          subtext: () => this.subgiftRecipients[data.origin_id] ? `To: ${this.subgiftRecipients[data.origin_id].join(', ')}` : undefined,
          text: `${from} just gifted ${(data as SubmysterygiftFields).number} subs`,
          time: new Date(time),
          title: `Subs gifted (${tier})`,
        })

        this.knownMultiGiftIDs.push(data.origin_id)
        return
      }

      if (data.origin_id) {
        this.subgiftRecipients[data.origin_id] = [
          ...this.subgiftRecipients[data.origin_id] || [],
          (data as SubgiftFields).to,
        ].sort((a, b) => a.localeCompare(b))
      }

      this.addEvent({
        eventId,
        extraData: { count: 1 },
        filterKey: 'subs',
        hasReplay: true,
        icon: 'fas fa-gift',
        originId: data.origin_id,
        text: `${from} just gifted ${(data as SubgiftFields).to} a sub`,
        time: new Date(time),
        title: `Sub gifted (${tier})`,
      })
    },

    handleTimeout(eventId: string, data: TimeoutFields, time: Date) {
      this.addEvent({
        eventId,
        filterKey: 'ban',
        icon: 'fas fa-ban',
        time: new Date(time),
        title: `${data.target_name} has been timed out for ${data.seconds}s`,
      })
    },

    handleTitleUpdate(eventId: string, data: TitleUpdateFields, time: Date) {
      this.addEvent({
        eventId,
        filterKey: 'streamUpdate',
        icon: 'fas fa-heading',
        text: data.title,
        time: new Date(time),
        title: 'Title updated',
      })
    },

    handleWatchStreak(eventId: string, data: WatchStreakFields, time: Date) {
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

    repeatEvent(eventId: string) {
      return this.eventClient!.replayEvent(eventId)
    },

    resolveSubtext(subtext: string | (() => string | undefined) | undefined): string | undefined {
      if (typeof subtext === 'function') {
        return subtext()
      }

      return subtext
    },

    storageKey(): string {
      if (!this.eventClient || typeof this.eventClient.paramOptionFallback('channel') !== 'string') {
        throw new Error('channel parameter not present')
      }

      const channel = this.eventClient.paramOptionFallback('channel')!.replace(/^#*/, '')
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

    timeDisplay(time: Date) {
      return dayjs(time).format('llll')
    },

    timeSince(time: Date) {
      return dayjs(time).from(this.now)
    },

    toggleFilterVisibility(filter: string) {
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

queueMicrotask(() => createApp(component).mount('#app'))

export default component
</script>

<style>
[v-cloak] { display: none; }
.border-event {
  border-left-width: 5px !important;
  border-left-style: solid !important;
  border-left-color: #9147ff;
}
.border-event.event-bits { border-left-color: #5cffbe !important; }
.border-event.event-channelpoint { border-left-color: #ffd37a !important; }
.border-event.event-follow { border-left-color: #ff38db !important; }
.border-event.event-raid { border-left-color: #ebeb00 !important; }
.border-event.event-streamOffline { border-left-color: rgb(var(--bs-danger-rgb)) !important; }
.border-event.event-subs { border-left-color: #1f69ff !important; }
.m50 {
  max-height: 40vh;
  overflow-y: auto;
}
.premono {
  font-family: monospace;
  font-size: 0.9em;
  white-space: pre-wrap;
}
</style>
