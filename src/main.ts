/* eslint-disable sort-imports */

import './style.scss'
import 'bootstrap/dist/css/bootstrap.css'
import '@fortawesome/fontawesome-free/css/all.css'

import 'bootstrap/dist/js/bootstrap.bundle'

import { createApp, h } from 'vue'

import router from './router'
import App from './components/app.vue'

const app = createApp({
  name: 'TwitchBotEditor',
  render() {
    return h(App)
  },

  router,
})

app.use(router)
app.mount('#app')
