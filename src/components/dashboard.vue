<template>
  <div class="container my-3">
    <div class="row justify-content-center">
      <!-- Here Number Scheduled Events Panel -->
      <div
        v-for="component in statusComponents"
        :key="component"
        class="col col-2"
      >
        <component
          :is="component"
        />
      </div>
    </div>

    <div class="row mt-3">
      <div class="col">
        <DashboardEventlog />
      </div>
      <div class="col">
        <DashboardChangelog />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import DashboardActiveRaffles from './dashboard/activeRaffles.vue'
import DashboardBotScopes from './dashboard/scopes.vue'
import DashboardChangelog from './dashboard/changelog.vue'
import DashboardEventlog from './dashboard/eventlog.vue'
import DashboardHealthCheck from './dashboard/healthcheck.vue'
import { defineComponent } from 'vue'

export default defineComponent({
  components: {
    DashboardActiveRaffles,
    DashboardBotScopes,
    DashboardChangelog,
    DashboardEventlog,
    DashboardHealthCheck,
  },

  data() {
    return {
      statusComponents: [
        'DashboardHealthCheck',
        'DashboardBotScopes',
        'DashboardActiveRaffles',
      ],
    }
  },

  name: 'TwitchBotEditorDashboard',
})
</script>
