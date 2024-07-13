<template>
  <div class="container my-3">
    <div class="row justify-content-center">
      <div class="col col-9">
        <div class="card">
          <div class="card-header">
            {{ $t('botauth.heading') }}
          </div>
          <div class="card-body">
            <p>{{ $t('botauth.description') }}</p>
            <ol>
              <li
                v-for="msg in $tm('botauth.directives')"
                :key="msg"
              >
                {{ msg }}
              </li>
            </ol>
            <div class="input-group">
              <input
                type="text"
                class="form-control"
                :value="authURLs?.update_bot_token || ''"
                :disabled="!authURLs?.update_bot_token"
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
        </div>
      </div>
      <div class="col col-3">
        <div
          v-if="botProfile.profile_image_url"
          class="card"
        >
          <div class="card-body text-center">
            <p>
              <img
                :src="botProfile.profile_image_url"
                class="img rounded-circle w-50"
              >
            </p>
            <p class="mb-0">
              <code>{{ botProfile.display_name }}</code>
            </p>
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
  data() {
    return {
      authURLs: {} as any,
      botProfile: {} as any,
      generalConfig: {} as any,
    }
  },

  methods: {
    /**
     * Copies auth-url for the bot into clipboard and gives user feedback
     * by colorizing copy-button for a short moment
     */
    copyAuthURL(): void {
      navigator.clipboard.writeText(this.authURLs.update_bot_token)
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
     * Fetches the bot profile (including display-name and profile
     * image) and stores it locally
     *
     * @param user Login-name of the user to fetch the profile for
     */
    fetchBotProfile(user: string): Promise<void> | undefined {
      return this.$root?.fetchJSON(`config-editor/user?user=${user}`)
        .then((data: any) => {
          this.botProfile = data
        })
    },

    /**
     * Fetches the general config object from the backend including the
     * authorized bot-name
     */
    fetchGeneralConfig(): Promise<void> | undefined {
      return this.$root?.fetchJSON('config-editor/general')
        .then((data: any) => {
          this.generalConfig = data
        })
        .then(() => this.fetchBotProfile(this.generalConfig.bot_name))
    },
  },

  mounted() {
    // Reload config after it changed
    this.bus.on(BusEventTypes.ConfigReload, () => this.fetchGeneralConfig())

    // Socket-reconnect could mean we need new auth-urls as the state
    // may have changed due to bot-restart
    this.bus.on(BusEventTypes.NotifySocketConnected, () => this.fetchAuthURLs())

    // Do initial fetches
    this.fetchAuthURLs()
    this.fetchGeneralConfig()
  },

  name: 'TwitchBotEditorBotAuth',
})
</script>
