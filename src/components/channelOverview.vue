<template>
  <div class="container my-3">
    <div class="row justify-content-center mb-3">
      <div class="col-6">
        <div class="input-group">
          <span class="input-group-text">
            <i class="fas fa-hashtag fa-fw me-1" />
          </span>
          <input
            v-model="inputAddChannel.text"
            type="text"
            :class="inputAddChannelClasses"
            @keypress.enter="addChannel"
          >
          <button
            class="btn btn-success"
            :disabled="!inputAddChannel.valid"
            @click="addChannel"
          >
            <i class="fas fa-plus fa-fw me-1" />
            {{ $t('channel.btnAdd') }}
          </button>
        </div>
      </div>
    </div>
    <div class="row justify-content-center">
      <div class="col">
        <table class="table">
          <thead>
            <tr>
              <th>{{ $t("channel.table.colChannel") }}</th>
              <th>{{ $t("channel.table.colPermissions") }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="channel in channels"
              :key="channel.name"
            >
              <td class="align-content-center">
                <i class="fas fa-hashtag fa-fw me-1" />
                {{ channel.name }}
              </td>
              <td class="align-content-center">
                <i
                  v-if="channel.numScopesGranted === 0"
                  class="fas fa-triangle-exclamation fa-fw me-1 text-danger"
                  :title="$t('channel.table.titleNoPermissions')"
                />
                <i
                  v-else-if="channel.numScopesGranted < numExtendedScopes"
                  class="fas fa-triangle-exclamation fa-fw me-1 text-warning"
                  :title="$t('channel.table.titlePartialPermissions')"
                />
                <i
                  v-else
                  class="fas fa-circle-check fa-fw me-1 text-success"
                  :title="$t('channel.table.titleAllPermissions')"
                />
                {{ $t('channel.table.textPermissions', {
                  avail: numExtendedScopes,
                  granted: channel.numScopesGranted
                }) }}
              </td>
              <td class="text-end">
                <div class="btn-group btn-group-sm">
                  <RouterLink
                    :to="{ name:'channelPermissions', params: { channel: channel.name } }"
                    class="btn btn-secondary"
                  >
                    <i class="fas fa-pencil-alt fa-fw" />
                  </RouterLink>
                  <button
                    class="btn btn-danger"
                    @click="removeChannel(channel.name)"
                  >
                    <i class="fas fa-minus fa-fw" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import BusEventTypes from '../helpers/busevents'
import { defineComponent } from 'vue'
import { successToast } from '../helpers/toasts'

export default defineComponent({
  computed: {
    channels(): Array<any> {
      return this.generalConfig.channels?.map((name: string) => ({
        name,
        numScopesGranted: (this.generalConfig.channel_scopes[name] || [])
          .filter((scope: string) => Object.keys(this.authURLs.available_extended_scopes).includes(scope))
          .length,
      }))
        .sort((a: any, b: any) => a.name.localeCompare(b.name))
    },

    inputAddChannelClasses(): string {
      const classes = ['form-control']

      if (this.inputAddChannel.valid) {
        classes.push('is-valid')
      } else if (this.inputAddChannel.text) {
        classes.push('is-invalid')
      }

      return classes.join(' ')
    },

    numExtendedScopes(): number {
      return Object.keys(this.authURLs.available_extended_scopes || {}).length
    },
  },

  data() {
    return {
      authURLs: {} as any,
      generalConfig: {} as any,
      inputAddChannel: {
        text: '',
        valid: false,
      },
    }
  },

  methods: {
    /**
     * Adds the channel entered into the input field to the list
     */
    addChannel(): Promise<void> | undefined {
      if (!this.inputAddChannel.valid) {
        return
      }

      const channel = this.inputAddChannel.text

      return this.updateGeneralConfig({
        ...this.generalConfig,
        channels: [
          ...this.generalConfig.channels.filter((chan: string) => chan !== channel),
          channel,
        ],
      })
        ?.then(() => {
          this.inputAddChannel.text = ''
          this.bus.emit(BusEventTypes.Toast, successToast(this.$t('channel.toastChannelAdded')))
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
     * Tells backend to remove a channel
     */
    removeChannel(channel: string): Promise<void> | undefined {
      return this.updateGeneralConfig({
        ...this.generalConfig,
        channels: this.generalConfig.channels.filter((chan: string) => chan !== channel),
      })
        ?.then(() => this.bus.emit(BusEventTypes.Toast, successToast(this.$t('channel.toastChannelRemoved'))))
    },

    /**
     * Writes general config back to backend
     *
     * @param config Configuration object to write (MUST contain all config)
     */
    updateGeneralConfig(config: any): Promise<void> | undefined {
      return this.$root?.fetchJSON('config-editor/general', {
        body: JSON.stringify(config),
        method: 'PUT',
      })
    },
  },

  mounted() {
    // Reload config after it changed
    this.bus.on(BusEventTypes.ConfigReload, () => this.fetchGeneralConfig())

    // Do initial fetches
    this.fetchAuthURLs()
    this.fetchGeneralConfig()
  },

  name: 'TwitchBotEditorChannelOverview',

  watch: {
    'inputAddChannel.text'(to) {
      this.inputAddChannel.valid = to.match(/^[a-zA-Z0-9_]{4,25}$/)
    },
  },
})
</script>
