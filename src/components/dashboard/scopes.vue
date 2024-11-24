<template>
  <StatusPanel
    :header="$t('dashboard.botScopes.header')"
    :loading="loading"
    :value="value"
    :value-extra-class="valueClass"
    :caption="$t('dashboard.botScopes.caption')"
    :click-route="{name:'botAuth'}"
  />
</template>

<script lang="ts">
import BusEventTypes from '../../helpers/busevents'
import { defineComponent } from 'vue'
import StatusPanel from './_statuspanel.vue'

export default defineComponent({
  components: { StatusPanel },

  computed: {
    nMissing(): number {
      return this.$root?.vars?.DefaultBotScopes
        ?.filter((scope: string) => !this.botScopes.includes(scope))
        .length || 0
    },

    value(): string {
      return `${this.nMissing}`
    },

    valueClass(): string {
      return this.nMissing === 0 ? 'text-success' : 'text-warning'
    },
  },

  data() {
    return {
      botScopes: [] as string[],
      loading: true,
    }
  },

  methods: {
    fetchGeneralConfig(): void {
      this.$root?.fetchJSON('config-editor/general')
        .then((data: any) => {
          this.botScopes = data.channel_scopes[data.bot_name] || []
          this.loading = false
        })
    },
  },

  mounted() {
    // Scopes might have changed due to authorization change
    this.bus.on(BusEventTypes.ConfigReload, () => this.fetchGeneralConfig())

    this.fetchGeneralConfig()
  },

  name: 'DashboardBotScopes',
})
</script>
