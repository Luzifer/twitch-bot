<template>
  <span
    class="fa-icon"
    v-bind="forwardedAttrs"
    :class="['fa-icon', $attrs.class]"
    v-html="svgHtml"
  />
</template>

<script lang="ts">
import { findIconDefinition, icon, type IconName, type IconPrefix } from '@fortawesome/fontawesome-svg-core'
import { defineComponent } from 'vue'

export default defineComponent({
  computed: {
    forwardedAttrs() {
      const attrs = { ...this.$attrs }
      delete attrs.class
      return attrs
    },

    svgHtml() {
      const [prefix, iconName] = this.icon as [IconPrefix, IconName]
      const definition = findIconDefinition({
        iconName,
        prefix,
      })

      return icon(definition, {
        classes: [
          this.fixedWidth ? 'fa-fw' : '',
          this.pulse ? 'fa-pulse' : '',
          this.spin ? 'fa-spin' : '',
          this.spinPulse ? 'fa-spin-pulse' : '',
        ].filter(Boolean),
      }).html.join('')
    },
  },

  inheritAttrs: false,

  name: 'TwitchBotFontAwesomeIcon',

  props: {
    fixedWidth: {
      default: false,
      type: Boolean,
    },

    icon: {
      required: true,
      type: Array,
    },

    pulse: {
      default: false,
      type: Boolean,
    },

    spin: {
      default: false,
      type: Boolean,
    },

    spinPulse: {
      default: false,
      type: Boolean,
    },
  },
})
</script>

<style scoped>
.fa-icon :deep(svg) {
  vertical-align: -0.125em;
}
</style>
