<template>
  <div class="nav flex-grow-1">
    <template
      v-for="section in navigation"
      :key="section.header"
    >
      <div class="navHeading user-select-none">
        {{ section.header }}
      </div>
      <RouterLink
        v-for="link in section.links"
        :key="link.target"
        :to="{name: link.target}"
        class="nav-link user-select-none"
      >
        <i :class="`${link.icon} fa-fw me-1`" />
        {{ link.name }}
      </RouterLink>
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'

export default defineComponent({
  components: { RouterLink },

  data() {
    return {
      navigation: [
        {
          header: this.$t('menu.headers.core'),
          links: [
            { icon: 'fas fa-chart-area', name: this.$t('menu.dashboard'), target: 'dashboard' },
            { icon: 'fas fa-cog', name: this.$t('menu.generalSettings'), target: 'generalSettings' },
          ],
        },
        {
          header: this.$t('menu.headers.chatInteraction'),
          links: [
            { icon: 'fas fa-envelope-open-text', name: this.$t('menu.autoMessages'), target: 'autoMessagesList' },
            { icon: 'fas fa-inbox', name: this.$t('menu.rules'), target: 'rulesList' },
          ],
        },
        {
          header: this.$t('menu.headers.modules'),
          links: [{ icon: 'fas fa-dice', name: this.$t('menu.raffles'), target: 'rafflesList' }],
        },
      ],
    }
  },

  name: 'TwitchBotEditorSideNav',
})
</script>

<style scoped>
.nav {
  flex-direction: column;
  flex-wrap: nowrap;
  overflow-y: auto;
}

.nav>.nav-link {
  align-items: center;
  color: inherit;
  display: flex;
  padding-bottom: 0.75rem;
  padding-left: 1.5rem;
  padding-top: 0.75rem;
  position: relative;
}

.nav>.nav-link.disabled {
  color: var(--bs-nav-link-disabled-color);
}

.navHeading {
  color: color-mix(in srgb, var(--bs-body-color) 50%, transparent);
  font-size: 0.75rem;
  font-weight: bold;
  padding: 1.75rem 1rem 0.75rem;
  text-transform: uppercase;
}
</style>
