<template>
  <StatusPanel
    :header="$t('dashboard.activeRaffles.header')"
    :loading="loading"
    :value="value"
    :click-route="{name: 'rafflesList'}"
    :caption="$t('dashboard.activeRaffles.caption')"
  />
</template>

<script lang="ts">
import BusEventTypes from '../../helpers/busevents'
import { defineComponent } from 'vue'
import StatusPanel from './_statuspanel.vue'

export default defineComponent({
  components: { StatusPanel },

  computed: {
    value(): string {
      return `${this.activeRaffles}`
    },
  },

  data() {
    return {
      activeRaffles: 0,
      loading: true,
    }
  },

  methods: {
    fetchRaffleCount(): void {
      fetch('raffle/', this.$root?.fetchOpts)
        .then((resp: Response) => this.$root?.parseResponseFromJSON(resp))
        .then((data: any) => {
          this.activeRaffles = data.filter((raffle: any) => raffle.status === 'active').length
          this.loading = false
        })
    },
  },

  mounted() {
    this.bus.on(BusEventTypes.RaffleChanged, () => {
      this.fetchRaffleCount()
    })
    this.fetchRaffleCount()
  },

  name: 'DashboardBotRaffles',
})
</script>
