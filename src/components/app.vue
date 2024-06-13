<template>
  <div class="h-100">
    <nav class="navbar navbar-expand-lg bg-body-tertiary">
      <div class="container-fluid">
        <a class="navbar-brand" href="#" @click.prevent>
          <i class="fas fa-robot fa-fw mr-1"></i>
          Twitch-Bot
        </a>

        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent"
          aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>

        <div class="collapse navbar-collapse" id="navbarSupportedContent">
          <ul class="navbar-nav me-auto mb-2 mb-lg-0"></ul>
          <ul class="navbar-nav ms-auto mb-2 mb-lg-0">
            <li class="nav-item dropdown">
              <a class="nav-link d-flex align-items-center" href="#" role="button" data-bs-toggle="dropdown"
                aria-expanded="false">
                <img class="rounded-circle nav-profile-image" src="https://placehold.co/400">
              </a>
              <ul class="dropdown-menu dropdown-menu-end">
                <li><a class="dropdown-item" href="#">Sign-out</a></li>
              </ul>
            </li>
          </ul>
        </div>
      </div>
    </nav>

    <div class="layout">
      <div class="layoutNav bg-body-tertiary d-flex">
        <div class="nav flex-grow-1">

          <div class="layoutNavHeading">Core</div>
          <a href="#" class="nav-link">
            <i class="fas fa-cog fa-fw me-1"></i>
            General Settings
          </a>

          <div class="layoutNavHeading">Chat Interaction</div>
          <a href="#" class="nav-link">
            <i class="fas fa-envelope-open-text fa-fw me-1"></i>
            Auto-Messages
          </a>
          <a href="#" class="nav-link">
            <i class="fas fa-inbox fa-fw me-1"></i>
            Rules
          </a>

          <div class="layoutNavHeading">Modules</div>
          <a href="#" class="nav-link">
            <i class="fas fa-dice fa-fw me-1"></i>
            Raffle
          </a>

        </div>
      </div>
      <div class="layoutContent">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
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

.layoutNavHeading {
  color: color-mix(in srgb, var(--bs-body-color) 50%, transparent);
  font-size: 0.75rem;
  font-weight: bold;
  padding: 1.75rem 1rem 0.75rem;
  text-transform: uppercase;
}

.nav-profile-image {
  max-width: 24px;
}

.navbar {
  z-index: 1000;
}
</style>
