/* eslint-disable sort-imports */

import VueRouter from 'vue-router'

import GeneralConfig from './generalConfig.vue'

const routes = [
  {
    component: GeneralConfig,
    name: 'general-config',
    path: '/',
  },
  {
    name: 'edit-automessages',
    path: '/automessages',
  },
  {
    name: 'edit-rules',
    path: '/rules',
  },
]

export default new VueRouter({
  routes,
})
