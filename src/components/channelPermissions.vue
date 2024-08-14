<template>
  <div class="container my-3">
    <div class="row justify-content-center mb-3">
      <div class="col-8">
        <p v-html="$t('channel.permissionStart', { channel })" />
        <div
          v-for="perm in permissions"
          :key="perm.scope"
          class="form-check form-switch"
        >
          <input
            :id="`switch${perm.scope}`"
            v-model="granted[perm.scope]"
            class="form-check-input"
            type="checkbox"
            role="switch"
          >
          <label
            class="form-check-label"
            :for="`switch${perm.scope}`"
          >{{ perm.description }}</label>
        </div>
        <div class="form-check form-switch mt-2">
          <input
            id="switch_all"
            v-model="allPermissions"
            class="form-check-input"
            type="checkbox"
            role="switch"
          >
          <label
            class="form-check-label"
            for="switch_all"
          >{{ $t('channel.permissionsAll') }}</label>
        </div>
        <div class="input-group mt-4">
          <input
            type="text"
            class="form-control"
            :value="permissionsURL || ''"
            :disabled="!permissionsURL"
            readonly
          >
          <button
            ref="copyBtn"
            class="btn btn-primary"
            :disabled="!authURLs?.update_bot_token"
            @click="copyAuthURL"
          >
            <i class="fas fa-clipboard fa-fw" />
          </button>
        </div>
      </div>
      <div class="col-4">
        <div class="card">
          <div class="card-header">
            <i class="fas fa-circle-info fa-fw me-1" />
            {{ $t('channel.permissionInfoHeader') }}
          </div>
          <div class="card-body">
            <p v-html="$t('channel.permissionIntro')" />
            <ul>
              <li
                v-for="(bpt, idx) in $tm('channel.permissionIntroBullets')"
                :key="`idx${idx}`"
              >
                {{ bpt }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import BusEventTypes from '../helpers/busevents'
import { defineComponent } from 'vue'

export default defineComponent({
  computed: {
    allPermissions: {
      get(): boolean {
        return this.extendedScopeNames
          .filter((scope: string) => !this.granted[scope])
          .length === 0
      },

      set(all: boolean): void {
        this.granted = Object.fromEntries(this.extendedScopeNames
          .map((scope: string) => [scope, all]))
      },
    },

    extendedScopeNames(): Array<string> {
      return Object.entries(this.authURLs.available_extended_scopes || {})
        .map(e => e[0])
    },

    permissions(): Array<any> {
      return Object.entries(this.authURLs.available_extended_scopes || {}).map(e => ({
        description: e[1],
        scope: e[0],
      }))
    },

    permissionsURL(): string {
      if (!this.authURLs.update_channel_scopes) {
        return ''
      }

      const scopes = Object.entries(this.granted).filter(e => e[1])
        .map(e => e[0])

      const u = new URL(this.authURLs.update_channel_scopes)
      u.searchParams.set('scope', scopes.join(' '))
      return u.toString()
    },
  },

  data() {
    return {
      authURLs: {} as any,
      generalConfig: {} as any,
      granted: {} as any,
    }
  },

  methods: {
    /**
     * Copies auth-url for the bot into clipboard and gives user feedback
     * by colorizing copy-button for a short moment
     */
    copyAuthURL(): void {
      navigator.clipboard.writeText(this.permissionsURL)
        .then(() => {
          const btn = this.$refs.copyBtn as Element
          btn.classList.replace('btn-primary', 'btn-success')
          window.setTimeout(() => btn.classList.replace('btn-success', 'btn-primary'), 2500)
        })
    },

    /**
     * Fetches auth-URLs from the backend
     */
    fetchAuthURLs(): Promise<void> | undefined {
      return this.$root?.fetchJSON('config-editor/auth-urls')
        .then((data: any) => {
          this.authURLs = data
        })
    },

    /**
     * Fetches the general config object from the backend
     */
    fetchGeneralConfig(): Promise<void> | undefined {
      return this.$root?.fetchJSON('config-editor/general')
        .then((data: any) => {
          this.generalConfig = data
        })
    },

    /**
     * Loads the granted scopes into the object for easier display
     * of the permission switches
     */
    loadScopes(): void {
      this.granted = Object.fromEntries((this.generalConfig.channel_scopes[this.channel] || []).map((scope: string) => [scope, true]))
    },
  },

  mounted() {
    // Reload config after it changed
    this.bus.on(BusEventTypes.ConfigReload, () => this.fetchGeneralConfig()?.then(() => this.loadScopes()))

    // Socket-reconnect could mean we need new auth-urls as the state
    // may have changed due to bot-restart
    this.bus.on(BusEventTypes.NotifySocketConnected, () => this.fetchAuthURLs())

    // Do initial fetches
    this.fetchAuthURLs()
    this.fetchGeneralConfig()?.then(() => this.loadScopes())
  },

  name: 'TwitchBotEditorChannelPermissions',

  props: {
    channel: {
      required: true,
      type: String,
    },
  },

  watch: {
    channel() {
      this.loadScopes()
    },
  },
})
</script>
