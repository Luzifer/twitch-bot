<template>
  <div
    :class="cardClass"
    @click="navigate"
  >
    <div class="card-body">
      <div class="fs-6 text-center">
        {{ header }}
      </div>
      <template v-if="loading">
        <div class="fs-1 text-center">
          <i class="fa-solid fa-circle-notch fa-spin" />
        </div>
      </template>
      <template v-else>
        <div :class="valueClass">
          {{ value }}
        </div>
      </template>
      <div
        v-if="caption"
        class="text-muted text-center"
      >
        <small>{{ caption }}</small>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { type RouteLocationRaw } from 'vue-router'

export default defineComponent({
  computed: {
    cardClass(): string {
      const classList = ['card']

      if (this.clickRoute) {
        classList.push('pointer-click')
      }

      return classList.join(' ')
    },

    valueClass(): string {
      const classList = ['fs-1 text-center']

      if (this.valueExtraClass) {
        classList.push(this.valueExtraClass)
      }

      return classList.join(' ')
    },
  },

  methods: {
    navigate(): void {
      if (!this.clickRoute) {
        return
      }

      this.$router.push(this.clickRoute)
    },
  },

  name: 'DashboardStatusPanel',

  props: {
    caption: {
      default: null,
      type: String,
    },

    clickRoute: {
      default: null,
      type: {} as PropType<RouteLocationRaw>,
    },

    header: {
      required: true,
      type: String,
    },

    loading: {
      default: false,
      type: Boolean,
    },

    value: {
      required: true,
      type: String,
    },

    valueExtraClass: {
      default: null,
      type: String,
    },
  },
})
</script>

<style scoped>
.pointer-click {
  cursor: pointer;
}
</style>
