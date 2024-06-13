<template>
  <div class="h-100">
    <head-nav :is-logged-in="false" />
    <div class="content d-flex align-items-center justify-content-center">
      <button
        class="btn btn-twitch"
        :disabled="loading"
        @click="openAuthURL"
      >
        <i :class="{'fa-fw me-1': true, 'fab fa-twitch': !loading, 'fas fa-circle-notch fa-spin': loading }" />
        Login with Twitch
      </button>
    </div>
    <toaster />
  </div>
</template>

<script lang="ts">
import BusEventTypes from '../helpers/busevents'
import { defineComponent } from 'vue'

import HeadNav from './_headNav.vue'
import Toaster from './_toaster.vue'

export default defineComponent({
  components: { HeadNav, Toaster },

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

  data() {
    return {
      loading: false,
    }
  },

  methods: {
    openAuthURL(): void {
      window.location.href = this.authURL
    },
  },

  mounted() {
    this.bus.on(BusEventTypes.LoginProcessing, (loading: boolean) => {
      this.loading = loading
    })
  },

  name: 'TwitchBotEditorLogin',
})
</script>

<style scoped>
.btn-twitch {
  background-color: #6441a5;
}

.content {
  height: calc(100vh - 56px);
}
</style>
