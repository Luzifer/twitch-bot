import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'

import BotAuth from './components/botauth.vue'
import BotEditors from './components/editors.vue'
import ChannelOverview from './components/channelOverview.vue'
import ChannelPermissions from './components/channelPermissions.vue'
import Dashboard from './components/dashboard.vue'

const routes = [
  {
    component: Dashboard,
    name: 'dashboard',
    path: '/',
  },

  // General settings
  {
    component: BotAuth,
    name: 'botAuth',
    path: '/bot-auth',
  },
  {
    children: [
      {
        component: ChannelOverview,
        name: 'channels',
        path: '',
      },
      {
        component: ChannelPermissions,
        name: 'channelPermissions',
        path: ':channel',
        props: true,
      },
    ],
    path: '/channels',
  },
  {
    component: BotEditors,
    name: 'editors',
    path: '/editors',
  },
  {
    component: {},
    name: 'tokens',
    path: '/tokens',
  },

  // Auto-Messages
  {
    children: [
      {
        component: {},
        name: 'autoMessagesList',
        path: '',
      },
      {
        component: {},
        name: 'autoMessageEdit',
        path: ':id',
      },
      {
        component: {},
        name: 'autoMessageNew',
        path: 'new',
      },
    ],
    path: '/auto-messages',
  },

  // Rules
  {
    children: [
      {
        component: {},
        name: 'rulesList',
        path: '',
      },
      {
        component: {},
        name: 'rulesEdit',
        path: ':id',
      },
      {
        component: {},
        name: 'rulesNew',
        path: 'new',
      },
    ],
    path: '/rules',
  },

  // Raffles
  {
    children: [
      {
        component: {},
        name: 'rafflesList',
        path: '',
      },
      {
        children: [
          {
            component: {},
            name: 'rafflesEdit',
            path: '',
          },
          {
            component: {},
            name: 'raffleEntries',
            path: 'entries',
          },
        ],
        path: ':id',
      },
      {
        component: {},
        name: 'rafflesNew',
        path: 'new',
      },
    ],
    path: '/raffles',
  },
] as RouteRecordRaw[]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
