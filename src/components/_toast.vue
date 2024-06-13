<template>
  <div
    ref="toast"
    :class="classForToast(toast)"
    role="alert"
    aria-live="assertive"
    aria-atomic="true"
  >
    <div class="d-flex">
      <div class="toast-body">
        {{ toast.text }}
      </div>
      <button
        type="button"
        :class="classForCloseButton(toast)"
        data-bs-dismiss="toast"
        aria-label="Close"
      />
    </div>
  </div>
</template>


<script lang="ts">
import { defineComponent, PropType } from 'vue'
import { Toast } from 'bootstrap'
import { ToastContent } from './_toaster.vue'

export default defineComponent({
  data() {
    return {
      hdl: null,
    }
  },

  emits: ['hidden'],

  methods: {
    classForCloseButton(toast: ToastContent): string {
      const classes = [
        'btn-close',
        'me-2',
        'm-auto',
      ]

      if (toast.color) {
        classes.push('btn-close-white')
      }

      return classes.join(' ')
    },

    classForToast(toast: ToastContent): string {
      const classes = [
        'toast',
        'align-items-center',
      ]

      if (toast.color) {
        classes.push('border-0', `text-bg-${toast.color}`)
      }

      return classes.join(' ')
    },
  },

  mounted() {
    this.$refs.toast.addEventListener('hidden.bs.toast', () => this.$emit('hidden'))
    this.hdl = new Toast(this.$refs.toast)
    this.hdl.show()
  },

  name: 'TwitchBotEditorToast',

  props: {
    toast: {
      required: true,
      type: Object as PropType<ToastContent>,
    },
  },
})
</script>
