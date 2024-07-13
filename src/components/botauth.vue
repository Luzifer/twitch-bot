<template>
  <div class="container my-3 user-select-none">
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
                class="form-control user-select-all"
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
    copyAuthURL(): void {
      navigator.clipboard.writeText(this.authURLs.update_bot_token)
        .then(() => {
          const btn = this.$refs.copyBtn as Element
          btn.classList.replace('btn-primary', 'btn-success')
          window.setTimeout(() => btn.classList.replace('btn-success', 'btn-primary'), 2500)
        })
    },

    fetchAuthURLs(): Promise<void> {
      return fetch('config-editor/auth-urls', this.$root?.fetchOpts)
        .then((resp: Response) => this.$root?.parseResponseFromJSON(resp))
        .then((data: any) => {
          this.authURLs = data
        })
    },

    fetchBotProfile(user: string): Promise<void> {
      return fetch(`config-editor/user?user=${user}`, this.$root?.fetchOpts)
        .then((resp: Response) => this.$root?.parseResponseFromJSON(resp))
        .then((data: any) => {
          this.botProfile = data
        })
    },

    fetchGeneralConfig(): Promise<void> {
      return fetch('config-editor/general', this.$root?.fetchOpts)
        .then((resp: Response) => this.$root?.parseResponseFromJSON(resp))
        .then((data: any) => {
          this.generalConfig = data
        })
    },
  },

  mounted() {
    this.fetchAuthURLs()
    this.fetchGeneralConfig()
      .then(() => this.fetchBotProfile(this.generalConfig.bot_name))
  },

  name: 'TwitchBotEditorBotAuth',
})
</script>
