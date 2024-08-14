import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'

import BothAuth from './components/botauth.vue'
import ChannelOverview from './components/channelOverview.vue'
import ChannelPermissions from './components/channelPermissions.vue'
import Dashboard from './components/dashboard.vue'

const routes = [
  { component: Dashboard, name: 'dashboard', path: '/' },

  // General settings
  { component: BothAuth, name: 'botAuth', path: '/bot-auth' },
  { component: ChannelOverview, name: 'channels', path: '/channels' },
  { component: ChannelPermissions, name: 'channelPermissions', path: '/channels/:channel/permissions', props: true },
  { component: {}, name: 'editors', path: '/editors' },
  { component: {}, name: 'tokens', path: '/tokens' },

  // Auto-Messages
  { component: {}, name: 'autoMessagesList', path: '/auto-messages' },
  { component: {}, name: 'autoMessageEdit', path: '/auto-messages/edit/:id' },
  { component: {}, name: 'autoMessageNew', path: '/auto-messages/new' },

  // Rules
  { component: {}, name: 'rulesList', path: '/rules' },
  { component: {}, name: 'rulesEdit', path: '/rules/edit/:id' },
  { component: {}, name: 'rulesNew', path: '/rules/new' },

  // Raffles
  { component: {}, name: 'rafflesList', path: '/raffles' },
  { component: {}, name: 'rafflesEdit', path: '/raffles/edit/:id' },
  { component: {}, name: 'rafflesNew', path: '/raffles/new' },
] as RouteRecordRaw[]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
