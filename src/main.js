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
Vue.config.silent = process.env.NODE_ENV !== 'dev'

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
    commonToastOpts: {
      appendToast: true,
      autoHideDelay: 3000,
      bodyClass: 'd-none',
      solid: true,
      toaster: 'b-toaster-bottom-right',
    },

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

    toastError(message, options = {}) {
      this.$bvToast.toast('...', {
        ...this.commonToastOpts,
        ...options,
        noAutoHide: true,
        title: message,
        variant: 'danger',
      })
    },

    toastInfo(message, options = {}) {
      this.$bvToast.toast('...', {
        ...this.commonToastOpts,
        ...options,
        title: message,
        variant: 'info',
      })
    },

    toastSuccess(message, options = {}) {
      this.$bvToast.toast('...', {
        ...this.commonToastOpts,
        ...options,
        title: message,
        variant: 'success',
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
