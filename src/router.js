/* eslint-disable sort-imports */

import VueRouter from 'vue-router'

import Automessages from './automessages.vue'
import GeneralConfig from './generalConfig.vue'
import Rules from './rules.vue'

const routes = [
  {
    component: GeneralConfig,
    name: 'general-config',
    path: '/',
  },
  {
    component: Automessages,
    name: 'edit-automessages',
    path: '/automessages',
  },
  {
    component: Rules,
    name: 'edit-rules',
    path: '/rules',
  },
]

export default new VueRouter({
  routes,
})
