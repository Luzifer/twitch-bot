<template>
  <div class="h-100">
    <head-nav :is-logged-in="true" />

    <div class="layout">
      <div class="layoutNav bg-body-tertiary d-flex">
        <div class="nav flex-grow-1">
          <div class="layoutNavHeading">
            Core
          </div>
          <a
            href="#"
            class="nav-link"
          >
            <i class="fas fa-cog fa-fw me-1" />
            General Settings
          </a>

          <div class="layoutNavHeading">
            Chat Interaction
          </div>
          <a
            href="#"
            class="nav-link"
          >
            <i class="fas fa-envelope-open-text fa-fw me-1" />
            Auto-Messages
          </a>
          <a
            href="#"
            class="nav-link"
          >
            <i class="fas fa-inbox fa-fw me-1" />
            Rules
          </a>

          <div class="layoutNavHeading">
            Modules
          </div>
          <a
            href="#"
            class="nav-link"
          >
            <i class="fas fa-dice fa-fw me-1" />
            Raffle
          </a>
        </div>
      </div>
      <div class="layoutContent">
        <router-view />

        <toaster />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

import HeadNav from './_headNav.vue'
import Toaster from './_toaster.vue'

export default defineComponent({
  components: { HeadNav, Toaster },

  name: 'TwitchBotEditorApp',
})
</script>

<style scoped>
.layout {
  display: flex;
}

.layoutContent {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  min-height: calc(100vh - 56px);
  min-width: 1;
  padding-left: 225px;
  position: relative;
}

.layoutNav {
  height: 100vh;
  left: 0;
  padding-top: 56px;
  position: fixed;
  right: 0;
  top: 0;
  width: 225px;
  z-index: 900;
}

.layoutNav>.nav {
  flex-direction: column;
  flex-wrap: nowrap;
  overflow-y: auto;
}

.layoutNav>.nav>.nav-link {
  align-items: center;
  color: inherit;
  display: flex;
  padding-bottom: 0.75rem;
  padding-left: 1.5rem;
  padding-top: 0.75rem;
  position: relative;
}

.layoutNav>.nav>.nav-link.disabled {
  color: var(--bs-nav-link-disabled-color);
}

.layoutNavHeading {
  color: color-mix(in srgb, var(--bs-body-color) 50%, transparent);
  font-size: 0.75rem;
  font-weight: bold;
  padding: 1.75rem 1rem 0.75rem;
  text-transform: uppercase;
}
</style>
