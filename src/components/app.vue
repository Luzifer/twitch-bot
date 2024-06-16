<template>
  <div class="h-100">
    <head-nav :is-logged-in="true" />

    <div class="layout">
      <div class="layoutNav bg-body-tertiary d-flex">
        <SideNav />
      </div>
      <div class="layoutContent">
        <router-view />

        <toaster />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import HeadNav from './_headNav.vue'
import SideNav from './_sideNav.vue'
import Toaster from './_toaster.vue'
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
  padding-top: 56px;
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
</style>
