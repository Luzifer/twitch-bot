<template>
  <div>
    <b-navbar
      toggleable="lg"
      type="dark"
      variant="primary"
      class="mb-3"
    >
      <b-navbar-brand :to="{ name: 'general-config' }">
        <font-awesome-icon
          fixed-width
          class="mr-1"
          :icon="['fas', 'robot']"
        />
        Twitch-Bot
      </b-navbar-brand>

      <b-navbar-toggle target="nav-collapse" />

      <b-collapse
        id="nav-collapse"
        is-nav
      >
        <b-navbar-nav v-if="isAuthenticated">
          <b-nav-item
            :to="{ name: 'general-config' }"
          >
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'cog']"
            />
            General
          </b-nav-item>
          <b-nav-item
            :to="{ name: 'edit-automessages' }"
          >
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'envelope-open-text']"
            />
            Auto-Messages
          </b-nav-item>
          <b-nav-item
            :to="{ name: 'edit-rules' }"
          >
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'inbox']"
            />
            Rules
          </b-nav-item>
          <b-nav-item
            :to="{ name: 'raffle' }"
          >
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'dice']"
            />
            Raffle
          </b-nav-item>
        </b-navbar-nav>

        <b-navbar-nav class="ml-auto">
          <b-nav-text
            v-if="loadingData"
          >
            <font-awesome-icon
              fixed-width
              class="text-warning"
              :icon="['fas', 'spinner']"
              pulse
            />
          </b-nav-text>

          <b-nav-text
            class="ml-2"
          >
            <template
              v-for="check in status.checks"
            >
              <font-awesome-icon
                :id="`statusCheck${check.name}`"
                :key="check.key"
                fixed-width
                :class="{ 'text-danger': !check.success, 'text-success': check.success }"
                :icon="['fas', 'question-circle']"
              />
              <b-tooltip
                :key="check.key"
                :target="`statusCheck${check.name}`"
                triggers="hover"
              >
                {{ check.description }}
              </b-tooltip>
            </template>
          </b-nav-text>

          <b-nav-text class="ml-2">
            <font-awesome-icon
              v-if="configNotifySocketConnected"
              id="socketConnectionStatus"
              fixed-width
              class="mr-1 text-success"
              :icon="['fas', 'ethernet']"
            />
            <font-awesome-icon
              v-else
              id="socketConnectionStatus"
              fixed-width
              class="mr-1 text-danger"
              :icon="['fas', 'ethernet']"
            />
            <b-tooltip
              target="socketConnectionStatus"
              triggers="hover"
            >
              <span v-if="configNotifySocketConnected">Connected to Bot</span>
              <span v-else>Disconnected from Bot</span>
            </b-tooltip>
          </b-nav-text>

          <b-nav-text class="ml-2">
            <font-awesome-icon
              id="botInfo"
              fixed-width
              class="mr-1"
              :icon="['fas', 'info-circle']"
            />
            <b-tooltip
              target="botInfo"
              triggers="hover"
            >
              Version: <code>{{ $root.vars.Version }}</code>
            </b-tooltip>
          </b-nav-text>
        </b-navbar-nav>
      </b-collapse>
    </b-navbar>

    <b-container>
      <!-- Error display -->
      <b-row
        v-if="error"
        class="sticky-row"
      >
        <b-col>
          <b-alert
            dismissible
            show
            variant="danger"
            @dismissed="error = null"
          >
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'exclamation-circle']"
            />
            {{ error }}
          </b-alert>
        </b-col>
      </b-row>

      <!-- Working display -->
      <b-row
        v-if="changePending"
        class="sticky-row"
      >
        <b-col>
          <b-alert
            show
            variant="info"
          >
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'spinner']"
              pulse
            />
            Your change was submitted and is pending, please wait for config to be updated!
          </b-alert>
        </b-col>
      </b-row>

      <!-- Logged-out state -->
      <b-row
        v-if="!isAuthenticated"
      >
        <b-col
          class="text-center"
        >
          <b-button
            :disabled="!$root.vars.TwitchClientID"
            :href="authURL"
            variant="twitch"
          >
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fab', 'twitch']"
            />
            Login with Twitch
          </b-button>
        </b-col>
      </b-row>

      <!-- Logged-in state -->
      <router-view v-else />
    </b-container>
  </div>
</template>

<script>
import * as constants from './const.js'

import axios from 'axios'

export default {
  computed: {
    authURL() {
      const scopes = []

      const params = new URLSearchParams()
      params.set('client_id', this.$root.vars.TwitchClientID)
      params.set('redirect_uri', window.location.href.split('#')[0].split('?')[0])
      params.set('response_type', 'token')
      params.set('scope', scopes.join(' '))

      return `https://id.twitch.tv/oauth2/authorize?${params.toString()}`
    },
  },

  created() {
    this.$bus.$on(constants.NOTIFY_CHANGE_PENDING, p => {
      this.changePending = Boolean(p)
    })
    this.$bus.$on(constants.NOTIFY_ERROR, err => {
      this.error = err
    })
    this.$bus.$on(constants.NOTIFY_FETCH_ERROR, err => {
      this.handleFetchError(err)
    })
    this.$bus.$on(constants.NOTIFY_LOADING_DATA, l => {
      this.loadingData = Boolean(l)
    })
  },

  data() {
    return {
      changePending: false,
      configNotifyBackoff: 100,
      configNotifySocket: null,
      configNotifySocketConnected: false,
      error: null,
      loadingData: false,
      status: {},
    }
  },

  methods: {
    fetchStatus() {
      return axios.get('status/status.json?fail-status=200')
        .then(resp => {
          this.status = resp.data
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    handleFetchError(err) {
      switch (err.response.status) {
      case 403:
        this.$root.authToken = null
        this.error = 'This user is not authorized for the config editor'
        break
      case 502:
        this.error = 'Looks like the bot is currently not reachable. Please check it is running and refresh the interface.'
        break
      default:
        this.error = `Something went wrong: ${err.response.data} (${err.response.status})`
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
        console.debug('[notify] Socket connected')
        this.configNotifySocketConnected = true
      }
      this.configNotifySocket.onmessage = evt => {
        const msg = JSON.parse(evt.data)

        console.debug(`[notify] Socket message received type=${msg.msg_type}`)
        this.configNotifyBackoff = 100 // We've received a message, reset backoff

        if (msg.msg_type === constants.NOTIFY_CONFIG_RELOAD) {
          this.$bus.$emit(constants.NOTIFY_CONFIG_RELOAD)
        }
      }
      this.configNotifySocket.onclose = evt => {
        console.debug(`[notify] Socket was closed wasClean=${evt.wasClean}`)
        this.configNotifySocketConnected = false
        updateBackoffAndReconnect()
      }
    },
  },

  mounted() {
    if (this.isAuthenticated) {
      this.openConfigNotifySocket()
    }

    window.setInterval(() => this.fetchStatus(), 10000)
    this.fetchStatus()
  },

  name: 'TwitchBotEditorApp',

  props: {
    isAuthenticated: {
      required: true,
      type: Boolean,
    },
  },

  watch: {
    isAuthenticated(to) {
      if (to && !this.configNotifySocketConnected) {
        this.openConfigNotifySocket()
      }
    },
  },
}
</script>

<style>
.btn-twitch {
  background-color: #6441a5;
}
.sticky-row {
  position: sticky;
  top: 0;
}
</style>
