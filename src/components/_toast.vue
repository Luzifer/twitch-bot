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
import { defineComponent, type PropType } from 'vue'
import { Toast } from 'bootstrap'

export type ToastContent = {
  id: string
  autoHide?: boolean
  color?: string
  delay?: number
  text: string
}

export default defineComponent({
  data() {
    return {
      hdl: null as Toast | null,
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
    const t: Element = this.$refs.toast as Element

    t.addEventListener('hidden.bs.toast', () => this.$emit('hidden'))

    this.hdl = new Toast(t, {
      autohide: this.toast.autoHide !== false,
      delay: this.toast.delay || 5000,
    })

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
