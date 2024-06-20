<template>
  <div class="card user-select-none">
    <div class="card-header">
      {{ $t('dashboard.changelog.heading') }}
    </div>
    <div
      class="card-body"
      v-html="changelog"
    />
  </div>
</template>

<script lang="ts">
// @ts-ignore - Has an esbuild loader to be loaded as text
import ChangeLog from '../../../History.md'

import { defineComponent } from 'vue'
import { parse as marked } from 'marked'

export default defineComponent({
  computed: {
    changelog(): string {
      const latestVersions = (ChangeLog as string)
        .split('\n')
        .filter((line: string) => line) // Remove empty lines to fix broken output
        .join('\n')
        .split('#')
        .slice(0, 3) // Last 2 versions (first element is empty)
        .join('###')

      const parts = [
        latestVersions,
        '---',
        this.$t('dashboard.changelog.fullLink'),
      ]

      return marked(parts.join('\n'), {
        async: false,
        breaks: false,
        extensions: null,
        gfm: true,
        hooks: null,
        pedantic: false,
        silent: true,
        tokenizer: null,
        walkTokens: null,
      }) as string
    },
  },

  name: 'DashboardChangelog',
})
</script>

<style scoped>
.card-body {
  user-select: text !important;
}
</style>
