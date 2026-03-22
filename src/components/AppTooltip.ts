import Tooltip from 'bootstrap/js/dist/tooltip'
import { defineComponent, nextTick } from 'vue'

function vnodeToText(node: any): string {
  if (node == null) {
    return ''
  }

  if (typeof node === 'string') {
    return node
  }

  if (Array.isArray(node)) {
    return node.map(vnodeToText).join('')
  }

  if (typeof node.children === 'string') {
    return node.children
  }

  if (Array.isArray(node.children)) {
    return node.children.map(vnodeToText).join('')
  }

  if (node.children?.default) {
    return vnodeToText(node.children.default())
  }

  return ''
}

export default defineComponent({
  computed: {
    content(): string {
      return vnodeToText(this.$slots.default?.()).trim()
    },
  },

  data() {
    return {
      tooltip: null as Tooltip | null,
    }
  },

  methods: {
    updateTooltip() {
      const target = document.getElementById(this.target)
      if (!target) {
        return
      }

      const title = this.content
      target.removeAttribute('title')
      target.removeAttribute('data-bs-original-title')
      target.setAttribute('data-bs-custom-class', 'app-tooltip app-tooltip-dark')
      target.setAttribute('data-bs-title', title)

      this.tooltip?.dispose()
      this.tooltip = new Tooltip(target, {
        title,
        trigger: this.triggers as 'hover focus' | 'click' | 'hover' | 'focus' | 'manual' | 'click hover' | 'click focus' | 'click hover focus' | undefined,
      })
    },
  },

  mounted() {
    this.updateTooltip()
  },

  name: 'TwitchBotAppTooltip',

  props: {
    target: {
      required: true,
      type: String,
    },

    triggers: {
      default: 'hover focus',
      type: String,
    },
  },

  render() {
    return null
  },

  unmounted() {
    this.tooltip?.dispose()
  },

  watch: {
    async content() {
      await nextTick()
      this.updateTooltip()
    },

    async target() {
      await nextTick()
      this.updateTooltip()
    },
  },
})
