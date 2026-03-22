<template>
  <div>
    <nav class="navbar navbar-expand-lg navbar-dark bg-body-tertiary mb-3">
      <div class="container-fluid">
        <RouterLink
          class="navbar-brand"
          :to="{ name: 'general-config' }"
        >
          <fa-icon
            fixed-width
            class="me-1"
            :icon="['fas', 'robot']"
          />
          Twitch-Bot
        </RouterLink>

        <button
          class="navbar-toggler"
          type="button"
          @click="navbarOpen = !navbarOpen"
        >
          <span class="navbar-toggler-icon" />
        </button>

        <div
          class="collapse navbar-collapse"
          :class="{ show: navbarOpen }"
        >
          <ul
            v-if="appStore.isAuthenticated"
            class="navbar-nav"
          >
            <li class="nav-item">
              <RouterLink
                class="nav-link"
                :to="{ name: 'general-config' }"
              >
                <fa-icon
                  fixed-width
                  class="me-1"
                  :icon="['fas', 'cog']"
                />
                General
              </RouterLink>
            </li>
            <li class="nav-item">
              <RouterLink
                class="nav-link"
                :to="{ name: 'edit-automessages' }"
              >
                <fa-icon
                  fixed-width
                  class="me-1"
                  :icon="['fas', 'envelope-open-text']"
                />
                Auto-Messages
              </RouterLink>
            </li>
            <li class="nav-item">
              <RouterLink
                class="nav-link"
                :to="{ name: 'edit-rules' }"
              >
                <fa-icon
                  fixed-width
                  class="me-1"
                  :icon="['fas', 'inbox']"
                />
                Rules
              </RouterLink>
            </li>
            <li class="nav-item">
              <RouterLink
                class="nav-link"
                :to="{ name: 'raffle' }"
              >
                <fa-icon
                  fixed-width
                  class="me-1"
                  :icon="['fas', 'dice']"
                />
                Raffle
              </RouterLink>
            </li>
          </ul>

          <div class="navbar-nav ms-auto align-items-lg-center gap-2">
            <span
              v-if="appStore.loadingData"
              class="navbar-text"
            >
              <fa-icon
                fixed-width
                class="text-warning"
                :icon="['fas', 'spinner']"
                spin-pulse
              />
            </span>

            <span class="navbar-text">
              <template
                v-for="check in statusChecks"
                :key="check.name"
              >
                <fa-icon
                  :id="`statusCheck${check.name}`"
                  fixed-width
                  class="me-2"
                  :class="{ 'text-danger': !check.success, 'text-success': check.success }"
                  :icon="['fas', 'question-circle']"
                />
                <AppTooltip :target="`statusCheck${check.name}`">
                  {{ check.description }}
                </AppTooltip>
              </template>
            </span>

            <span class="navbar-text">
              <fa-icon
                id="socketConnectionStatus"
                fixed-width
                :class="configNotifySocketConnected ? 'text-success' : 'text-danger'"
                :icon="['fas', 'ethernet']"
              />
              <AppTooltip target="socketConnectionStatus">
                {{ configNotifySocketConnected ? 'Connected to Bot' : 'Disconnected from Bot' }}
              </AppTooltip>
            </span>

            <span class="navbar-text">
              <fa-icon
                id="botInfo"
                fixed-width
                :icon="['fas', 'info-circle']"
              />
              <AppTooltip target="botInfo">
                Version: {{ appStore.vars.Version || '-' }}
              </AppTooltip>
            </span>
          </div>
        </div>
      </div>
    </nav>

    <div class="toast-container position-fixed top-0 end-0 p-3">
      <div
        v-for="toast in appStore.toasts"
        :key="toast.id"
        class="toast show border-0"
        :class="toastClass(toast.variant)"
      >
        <div class="toast-body">
          {{ toast.message }}
        </div>
      </div>
    </div>

    <div class="container">
      <div
        v-if="appStore.error"
        class="row sticky-row"
      >
        <div class="col">
          <div class="alert alert-danger alert-dismissible">
            <fa-icon
              fixed-width
              class="me-1"
              :icon="['fas', 'exclamation-circle']"
            />
            {{ appStore.error }}
            <button
              type="button"
              class="btn-close"
              aria-label="Close"
              @click="appStore.setError(null)"
            />
          </div>
        </div>
      </div>

      <div
        v-if="appStore.changePending"
        class="row sticky-row"
      >
        <div class="col">
          <div class="alert alert-info">
            <fa-icon
              fixed-width
              class="me-1"
              :icon="['fas', 'spinner']"
              spin-pulse
            />
            Your change was submitted and is pending, please wait for config to be updated!
          </div>
        </div>
      </div>

      <div
        v-if="!appStore.isAuthenticated"
        class="row"
      >
        <div class="col text-center">
          <a
            class="btn btn-twitch"
            :class="{ disabled: !appStore.vars.TwitchClientID }"
            :href="authURL"
          >
            <fa-icon
              fixed-width
              class="me-1"
              :icon="['fab', 'twitch']"
            />
            Login with Twitch
          </a>
        </div>
      </div>

      <RouterView v-else />
    </div>

    <ConfirmModalHost />
  </div>
</template>

<script lang="ts">
import * as constants from './lib/const'
import { api, HttpError } from './api'
import type { ConfigNotifyMessage, StatusResponse } from './types'
import { RouterLink, RouterView } from 'vue-router'
import AppTooltip from './components/AppTooltip'
import ConfirmModalHost from './components/ConfirmModalHost'
import { defineComponent } from 'vue'
import { useAppStore } from './stores/app'

export default defineComponent({
  components: {
    AppTooltip,
    ConfirmModalHost,
    RouterLink,
    RouterView,
  },

  computed: {
    authURL(): string {
      const params = new URLSearchParams()
      params.set('client_id', this.appStore.vars.TwitchClientID)
      params.set('redirect_uri', window.location.href.split('#')[0].split('?')[0])
      params.set('response_type', 'token')
      params.set('scope', '')

      return `https://id.twitch.tv/oauth2/authorize?${params.toString()}`
    },

    statusChecks() {
      return this.appStore.status?.checks || []
    },
  },

  data() {
    return {
      appStore: useAppStore(),
      configNotifyBackoff: 100,
      configNotifySocket: null as WebSocket | null,
      configNotifySocketConnected: false,
      navbarOpen: false,
    }
  },

  methods: {
    async fetchStatus() {
      try {
        this.appStore.setStatus(await api.get<StatusResponse>('status/status.json?fail-status=200', false) as StatusResponse)
      } catch (err) {
        this.$bus.emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    handleFetchError(err: unknown) {
      const httpErr = err as HttpError

      switch (httpErr.status) {
      case 403:
        this.appStore.setAuthToken(null)
        this.appStore.setError('This user is not authorized for the config editor')
        break
      case 502:
        this.appStore.setError('Looks like the bot is currently not reachable. Please check it is running and refresh the interface.')
        break
      default:
        this.appStore.setError(`Something went wrong: ${String(httpErr.data)} (${httpErr.status})`)
      }
    },

    openConfigNotifySocket() {
      if (this.configNotifySocket) {
        this.configNotifySocket.close()
        this.configNotifySocket = null
      }

      const updateBackoffAndReconnect = () => {
        this.configNotifyBackoff = Math.min(this.configNotifyBackoff * 1.5, 10000)
        window.setTimeout(() => this.openConfigNotifySocket(), this.configNotifyBackoff)
      }

      this.configNotifySocket = new WebSocket(`${window.location.href.split('#')[0].replace(/^http/, 'ws')}config-editor/notify-config`)
      this.configNotifySocket.onopen = () => {
        this.configNotifySocketConnected = true
      }
      this.configNotifySocket.onmessage = evt => {
        const msg = JSON.parse(evt.data) as ConfigNotifyMessage
        this.configNotifyBackoff = 100

        if (msg.msg_type !== 'ping') {
          this.$bus.emit(msg.msg_type)
        }
      }
      this.configNotifySocket.onclose = () => {
        this.configNotifySocketConnected = false
        updateBackoffAndReconnect()
      }
    },

    toastClass(variant: string) {
      return {
        'bg-danger text-white': variant === 'danger',
        'bg-info text-dark': variant === 'info',
        'bg-success text-white': variant === 'success',
      }
    },
  },

  mounted() {
    this.$bus.on(constants.NOTIFY_CHANGE_PENDING, payload => {
      this.appStore.setChangePending(Boolean(payload))
    })
    this.$bus.on(constants.NOTIFY_ERROR, payload => {
      this.appStore.setError(payload as string)
    })
    this.$bus.on(constants.NOTIFY_FETCH_ERROR, payload => {
      this.handleFetchError(payload)
    })
    this.$bus.on(constants.NOTIFY_LOADING_DATA, payload => {
      this.appStore.setLoadingData(Boolean(payload))
    })

    if (this.appStore.isAuthenticated) {
      this.openConfigNotifySocket()
    }

    window.setInterval(() => this.fetchStatus(), 10000)
    this.fetchStatus()
  },

  name: 'TwitchBotApp',

  watch: {
    'appStore.isAuthenticated'(to: boolean) {
      if (to && !this.configNotifySocketConnected) {
        this.openConfigNotifySocket()
      }
    },
  },
})
</script>

<style>
:root {
  --bs-body-font-size: 0.9rem;
}

.btn-twitch {
  background-color: #6441a5;
  color: #fff;
}

.btn-twitch:hover {
  background-color: #7d5bbe;
  color: #fff;
}

.sticky-row {
  position: sticky;
  top: 0;
  z-index: 1020;
}

.app-tooltip .tooltip-inner {
  max-width: 320px;
  text-align: left;
  white-space: pre-line;
}

.app-tooltip-dark {
  --bs-tooltip-bg: #111827;
  --bs-tooltip-color: #f9fafb;
}

.badge {
  align-items: center;
  display: inline-flex;
  line-height: 1.2;
  vertical-align: middle;
}
</style>
