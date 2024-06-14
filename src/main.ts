/* eslint-disable sort-imports */
/* global RequestInit, TimerHandler */

import './style.scss' // Internal global styles
import 'bootstrap/dist/css/bootstrap.css' // Bootstrap 5 Styles
import '@fortawesome/fontawesome-free/css/all.css' // All FA free icons

import { createApp, h } from 'vue'
import mitt from 'mitt'

import BusEventTypes from './helpers/busevents'
import ConfigNotifyListener from './helpers/configNotify'
import { errorToast } from './helpers/toasts'

import router from './router'
import App from './components/app.vue'
import Login from './components/login.vue'

const app = createApp({
  computed: {
    fetchOpts(): RequestInit {
      return {
        credentials: 'same-origin',
        headers: {
          'Accept': 'application/json',
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json',
        },
      }
    },

    tokenRenewAt(): Date | null {
      if (this.tokenExpiresAt === null || this.tokenExpiresAt.getTime() < this.now.getTime()) {
        // We don't know when it expires or it's expired, we can't renew
        return null
      }

      // We renew 720sec before expiration (0.8 * 1h)
      return new Date(this.tokenExpiresAt.getTime() - 720000)
    },
  },

  data(): Object {
    return {
      now: new Date(),
      tickers: {},
      token: '',
      tokenExpiresAt: null as Date | null,
      tokenUser: '',
      userInfo: null as null | {},
      vars: {},
    }
  },

  methods: {
    /**
     * Checks whether the API returned an 403 and in case it did triggers
     * a logout and throws the user back into the login screen
     *
     * @param resp The response to the fetch request
     * @returns The Response object from the resp parameter
     */
    check403(resp: Response): Response {
      if (resp.status === 403) {
        /*
         * User token is not valid and therefore should be removed
         * which essentially triggers a logout
         */
        this.logout()
        throw new Error('user has been logged out')
      }

      return resp
    },

    loadVars(): Promise<void | Response> {
      return fetch('editor/vars.json')
        .then((resp: Response) => resp.json())
        .then((data: any) => {
          this.vars = data
        })
    },

    login(token: string, expiresAt: Date, username: string): void {
      this.token = token
      this.tokenExpiresAt = expiresAt
      this.tokenUser = username
      window.localStorage.setItem('twitch-bot-token', JSON.stringify({ expiresAt, token, username }))
      // Nuke the Twitch auth-response from the browser history
      window.history.replaceState(null, '', window.location.href.split('#')[0])

      fetch(`config-editor/user?user=${this.tokenUser}`, this.$root.fetchOpts)
        .then((resp: Response) => this.$root.check403(resp))
        .then((resp: Response) => resp.json())
        .then((data: any) => {
          this.userInfo = data
        })
    },

    logout(): void {
      window.localStorage.removeItem('twitch-bot-token')
      this.token = ''
      this.tokenExpiresAt = null
      this.tokenUser = ''
    },

    registerTicker(id: string, func: TimerHandler, intervalMs: number): void {
      this.unregisterTicker(id)
      this.tickers[id] = window.setInterval(func, intervalMs)
    },

    renewToken(): void {
      if (!this.tokenRenewAt || this.tokenRenewAt.getTime() > this.now.getTime()) {
        return
      }

      fetch('config-editor/refreshToken', this.$root.fetchOpts)
        .then((resp: Response) => this.$root.check403(resp))
        .then((resp: Response) => resp.json())
        .then((data: any) => this.login(data.token, new Date(data.expiresAt), data.user))
    },

    unregisterTicker(id: string): void {
      if (this.tickers[id]) {
        window.clearInterval(this.tickers[id])
      }
    },
  },

  mounted(): void {
    this.bus.on('logout', this.logout)

    this.$root.registerTicker('updateRootNow', () => {
      this.now = new Date()
    }, 30000)
    this.$root.registerTicker('renewToken', () => this.renewToken(), 60000)

    // Start background-listen for config updates
    new ConfigNotifyListener((msgType: string) => {
      this.bus.emit(msgType)
    })

    this.loadVars()

    const params = new URLSearchParams(window.location.hash.replace(/^[#/]+/, ''))
    const authToken = params.get('access_token')
    if (authToken) {
      this.bus.emit(BusEventTypes.LoginProcessing, true)
      fetch('config-editor/login', {
        body: JSON.stringify({ token: authToken }),
        headers: { 'Content-Type': 'application/json' },
        method: 'POST',
      })
        .then((resp: Response): any => {
          if (resp.status !== 200) {
            let errorText = 'Login failed unexpectedly'
            if (resp.status === 403) {
              errorText = 'Access denied to this bot instance'
            }

            this.bus.emit(BusEventTypes.LoginProcessing, false)
            this.bus.emit(BusEventTypes.Toast, errorToast(errorText))
            throw new Error(`login failed, status=${resp.status}`)
          }

          return resp.json()
        })
        .then((data: any) => this.login(data.token, new Date(data.expiresAt), data.user))
    } else {
      const tokenData = window.localStorage.getItem('twitch-bot-token')
      if (tokenData !== null) {
        const data = JSON.parse(tokenData)
        this.login(data.token, new Date(data.expiresAt), data.username)
      }
    }
  },

  name: 'TwitchBotEditor',
  render() {
    if (this.token) {
      return h(App)
    }

    return h(Login)
  },

  router,
})

app.config.globalProperties.bus = mitt()
app.use(router)
app.mount('#app')
