/* eslint-disable sort-imports */

// Darkly design
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'
import 'bootswatch/dist/darkly/bootstrap.css'

import axios from 'axios'

// Vue & BootstrapVue
import Vue from 'vue'
import { BootstrapVue } from 'bootstrap-vue'
import VueRouter from 'vue-router'

Vue.use(BootstrapVue)
Vue.use(VueRouter)

// FontAwesome
import { library } from '@fortawesome/fontawesome-svg-core'
import { fab } from '@fortawesome/free-brands-svg-icons'
import { fas } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon, FontAwesomeLayers } from '@fortawesome/vue-fontawesome'

library.add(fab, fas)
Vue.component('FontAwesomeIcon', FontAwesomeIcon)
Vue.component('FontAwesomeLayers', FontAwesomeLayers)

// App
import App from './app.vue'
import Router from './router.js'

Vue.config.devtools = process.env.NODE_ENV === 'dev'

Vue.prototype.$bus = new Vue()

new Vue({
  components: { App },
  computed: {
    axiosOptions() {
      return {
        headers: {
          authorization: this.authToken,
        },
      }
    },
  },

  data: {
    authToken: null,
    vars: {},
  },

  el: '#app',

  methods: {
    fetchVars() {
      return axios.get('editor/vars.json')
        .then(resp => {
          this.vars = resp.data
        })
    },
  },

  mounted() {
    this.fetchVars()

    const params = new URLSearchParams(window.location.hash.replace(/^[#/]+/, ''))
    if (params.has('access_token')) {
      this.authToken = params.get('access_token') || null
      this.$router.replace({ name: 'general-config' })
    }
  },

  name: 'TwitchBotEditor',

  render(h) {
    return h(App, { props: { isAuthenticated: Boolean(this.authToken) } })
  },

  router: Router,
})
