<template>
  <StatusPanel
    :header="$t('dashboard.healthCheck.header')"
    :loading="!status.checks"
    :value="value"
    :value-extra-class="valueClass"
    :caption="$t('dashboard.healthCheck.caption')"
  />
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import StatusPanel from './_statuspanel.vue'

export default defineComponent({
  components: { StatusPanel },

  computed: {
    nChecks(): Number {
      return this.status?.checks?.length || 0
    },

    nSuccess(): Number {
      return this.status?.checks?.filter((check: any) => check?.success).length || 0
    },

    value(): string {
      return `${this.nSuccess} / ${this.nChecks}`
    },

    valueClass(): string {
      return this.nSuccess === this.nChecks ? 'text-success' : 'text-danger'
    },
  },

  data() {
    return {
      status: {} as any,
    }
  },

  methods: {
    fetchStatus(): void {
      this.$root?.fetchJSON('status/status.json?fail-status=200')
        .then((data: any) => {
          this.status = data
        })
    },
  },

  mounted() {
    this.$root?.registerTicker('dashboardHealthCheck', () => this.fetchStatus(), 30000)
    this.fetchStatus()
  },

  name: 'DashboardHealthCheck',

  unmounted() {
    this.$root?.unregisterTicker('dashboardHealthCheck')
  },
})
</script>
