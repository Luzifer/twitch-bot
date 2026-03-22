/* eslint-disable sort-imports */

import 'bootstrap/dist/css/bootstrap.min.css'
import 'codejar-linenumbers/js/codejar-linenumbers.css'

import { fab } from '@fortawesome/free-brands-svg-icons'
import { fas } from '@fortawesome/free-solid-svg-icons'
import { library } from '@fortawesome/fontawesome-svg-core'
import { createPinia } from 'pinia'
import { createApp, defineComponent, h } from 'vue'

import App from './app.vue'
import { bus } from './lib/eventBus'
import FaIcon from './components/FaIcon.vue'
import router from './router'
import { useAppStore } from './stores/app'

library.add(fab, fas)
document.documentElement.setAttribute('data-bs-theme', 'dark')

const pinia = createPinia()

const root = createApp(defineComponent({
  async mounted() {
    const appStore = useAppStore()
    await appStore.fetchVars()

    const params = new URLSearchParams(window.location.hash.replace(/^[#/]+/, ''))
    if (params.has('access_token')) {
      appStore.setAuthToken(params.get('access_token') || null)
      this.$router.replace({ name: 'general-config' })
    }
  },

  render() {
    return h(App)
  },
}))

root.use(pinia)
root.use(router)
root.config.globalProperties.$bus = bus
root.component('FaIcon', FaIcon)
root.component('FontAwesomeIcon', FaIcon)
root.mount('#app')
