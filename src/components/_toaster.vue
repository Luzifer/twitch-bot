<template>
  <div class="toast-container bottom-0 end-0 p-3">
    <toast
      v-for="toast in toasts"
      :key="toast.id"
      :toast="toast"
      @hidden="removeToast(toast.id)"
    />
  </div>
</template>

<script lang="ts">
import Toast, { ToastContent } from './_toast.vue'
import BusEventTypes from '../helpers/busevents'
import { defineComponent } from 'vue'

export default defineComponent({
  components: { Toast },

  data() {
    return {
      toasts: [] as ToastContent[],
    }
  },

  methods: {
    removeToast(id: string) {
      this.toasts = this.toasts.filter((t: ToastContent) => t.id !== id)
    },
  },

  mounted() {
    this.bus.on(BusEventTypes.Toast, (toast: ToastContent) => this.toasts.push({
      ...toast,
      id: toast.id || crypto.randomUUID(),
    }))
  },

  name: 'TwitchBotEditorToaster',
})
</script>
