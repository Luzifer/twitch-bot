<template>
  <nav class="navbar navbar-expand-lg bg-body-tertiary">
    <div class="container-fluid">
      <span class="navbar-brand">
        <i class="fas fa-robot fa-fw me-1 text-info" />
        Twitch-Bot
      </span>

      <button
        class="navbar-toggler"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#navbarSupportedContent"
        aria-controls="navbarSupportedContent"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon" />
      </button>

      <div
        id="navbarSupportedContent"
        class="collapse navbar-collapse"
      >
        <ul class="navbar-nav me-auto mb-2 mb-lg-0" />
        <ul class="navbar-nav ms-auto mb-2 mb-lg-0">
          <li
            v-if="isLoggedIn"
            class="nav-item dropdown"
          >
            <a
              ref="userMenuToggle"
              class="nav-link d-flex align-items-center"
              href="#"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              <img
                class="rounded-circle nav-profile-image"
                :src="profileImage"
              >
            </a>
            <ul class="dropdown-menu dropdown-menu-end">
              <li>
                <a
                  class="dropdown-item"
                  href="#"
                  @click.prevent="logout"
                >{{ $t('nav.signOut') }}</a>
              </li>
            </ul>
          </li>
        </ul>
      </div>
    </div>
  </nav>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { Dropdown } from 'bootstrap'

export default defineComponent({
  computed: {
    profileImage(): string {
      return this.$root?.userInfo?.profile_image_url || ''
    },
  },

  methods: {
    logout() {
      this.bus.emit('logout')
    },
  },

  mounted() {
    if (this.isLoggedIn) {
      new Dropdown(this.$refs.userMenuToggle as Element)
    }
  },

  name: 'TwitchBotEditorHeadNav',

  props: {
    isLoggedIn: {
      required: true,
      type: Boolean,
    },
  },
})
</script>

<style>
.nav-profile-image {
  max-width: 24px;
}

.navbar {
  z-index: 1000;
}
</style>
