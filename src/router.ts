import { createRouter, createWebHashHistory } from 'vue-router'

import Automessages from './views/automessages.vue'
import GeneralConfig from './views/generalConfig.vue'
import Raffle from './views/raffle.vue'
import Rules from './views/rules.vue'

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
    component: Raffle,
    name: 'raffle',
    path: '/raffle',
  },
  {
    component: Rules,
    name: 'edit-rules',
    path: '/rules',
  },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
