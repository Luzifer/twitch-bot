const CRON_VALIDATION = /^(?:(?:@every (?:\d+(?:s|m|h))+)|(?:(?:(?:(?:\d+,)+\d+|(?:\d+(?:\/|-)\d+)|\d+|\*|\*\/\d+)(?: |$)){5}))$/

Vue.config.devtools = true
new Vue({
  computed: {
    authURL() {
      const scopes = []

      const params = new URLSearchParams()
      params.set('client_id', this.vars.TwitchClientID)
      params.set('redirect_uri', window.location.href.split('#')[0])
      params.set('response_type', 'token')
      params.set('scope', scopes.join(' '))

      return `https://id.twitch.tv/oauth2/authorize?${params.toString()}`
    },

    axiosOptions() {
      return {
        headers: {
          authorization: this.authToken,
        },
      }
    },

    sortedChannels() {
      return this.generalConfig?.channels
        ?.sort((a, b) => a.toLocaleLowerCase().localeCompare(b.toLocaleLowerCase()))
    },

    sortedEditors() {
      return this.generalConfig?.bot_editors
        ?.sort((a, b) => {
          const an = this.userProfiles[a]?.login || a
          const bn = this.userProfiles[b]?.login || b

          return an.localeCompare(bn)
        })
    },

    validateAutoMessage() {
      if (!this.models.autoMessage.sendMode) {
        return false
      }

      if (this.models.autoMessage.sendMode === 'cron' && !this.validateAutoMessageCron) {
        return false
      }

      if (this.models.autoMessage.sendMode === 'lines' && (!this.models.autoMessage.message_interval || Number(this.models.autoMessage.message_interval) <= 0)) {
        return false
      }

      if (this.validateAutoMessageMessageLength < this.models.autoMessage.message?.length) {
        return false
      }

      if (!this.validateAutoMessageChannel) {
        return false
      }

      return true
    },

    validateAutoMessageChannel() {
      return Boolean(this.models.autoMessage.channel?.match(/^[a-zA-Z0-9_]{4,25}$/))
    },

    validateAutoMessageCron() {
      if (this.models.autoMessage.sendMode !== 'cron' && !this.models.autoMessage.cron) {
        return true
      }

      return Boolean(this.models.autoMessage.cron?.match(CRON_VALIDATION))
    },

    validateAutoMessageMessageLength() {
      return this.models.autoMessage.use_action ? 496 : 500
    },
  },

  data: {
    authToken: null,
    autoMessageFields: [
      {
        key: 'channel',
        sortable: true,
        class: 'col-1 text-nowrap',
        thClass: 'align-middle',
      },
      {
        key: 'message',
        sortable: true,
        class: 'col-9',
        thClass: 'align-middle',
      },
      {
        key: 'cron',
        class: 'col-1 text-nowrap',
        thClass: 'align-middle',
      },
      {
        key: 'actions',
        label: '',
        class: 'col-1 text-right',
        thClass: 'align-middle',
      },
    ],

    autoMessages: [],
    autoMessageSendModes: [
      { text: 'Cron', value: 'cron' },
      { text: 'Number of lines', value: 'lines' },
    ],

    editMode: 'automessages', // 'general', FIXME
    error: null,
    generalConfig: {},
    models: {
      addChannel: '',
      addEditor: '',
      autoMessage: {},
    },

    rules: [],
    showAutoMessageEditModal: false,
    userProfiles: {},
    vars: {},
  },

  el: '#app',

  methods: {
    addChannel() {
      this.generalConfig.channels.push(this.models.addChannel.replace(/^#*/, ''))
      this.models.addChannel = ''

      this.updateGeneralConfig()
    },

    addEditor() {
      this.fetchProfile(this.models.addEditor)
      this.generalConfig.bot_editors.push(this.models.addEditor)
      this.models.addEditor = ''

      this.updateGeneralConfig()
    },

    deleteAutoMessage(uuid) {
      axios.delete(`config-editor/auto-messages/${uuid}`, this.axiosOptions)
        .then(() => window.setTimeout(() => this.reload(), 1000))
        .catch(err => this.handleFetchError(err))
    },

    editAutoMessage(msg) {
      Vue.set(this.models, 'autoMessage', {
        ...msg,
        sendMode: msg.cron ? 'cron' : 'lines',
      })
      this.showAutoMessageEditModal = true
    },

    fetchAutoMessages() {
      axios.get('config-editor/auto-messages', this.axiosOptions)
        .then(resp => {
          this.autoMessages = resp.data
        })
        .catch(err => this.handleFetchError(err))
    },

    fetchGeneralConfig() {
      axios.get('config-editor/general', this.axiosOptions)
        .then(resp => {
          this.generalConfig = resp.data
        })
        .catch(err => this.handleFetchError(err))
        .then(() => {
          for (const editor of this.generalConfig.bot_editors) {
            this.fetchProfile(editor)
          }
        })
    },

    fetchProfile(user) {
      axios.get(`config-editor/user?user=${user}`, this.axiosOptions)
        .then(resp => Vue.set(this.userProfiles, user, resp.data))
        .catch(err => this.handleFetchError(err))
    },

    fetchRules() {
      axios.get('config-editor/rules', this.axiosOptions)
        .then(resp => {
          this.rules = resp.data
        })
        .catch(err => this.handleFetchError(err))
    },

    fetchVars() {
      axios.get('editor/vars.json')
        .then(resp => {
          this.vars = resp.data
        })
    },

    handleFetchError(err) {
      switch (err.response.status) {
      case 403:
        this.authToken = null
        this.error = 'This user is not authorized for the config editor'
        break
      case 502:
        this.error = 'Looks like the bot is currently not reachable. Please check it is running and refresh the interface.'
        break
      default:
        this.error = `Something went wrong: ${err.response.data} (${err.response.status})`
      }
    },

    newAutoMessage() {
      Vue.set(this.models, 'autoMessage', {})
      this.showAutoMessageEditModal = true
    },

    reload() {
      this.fetchAutoMessages()
      this.fetchGeneralConfig()
      this.fetchRules()
    },

    removeChannel(channel) {
      this.generalConfig.channels = this.generalConfig.channels
        .filter(ch => ch !== channel)

      this.updateGeneralConfig()
    },

    removeEditor(editor) {
      this.generalConfig.bot_editors = this.generalConfig.bot_editors
        .filter(ed => ed !== editor)

      this.updateGeneralConfig()
    },

    saveAutoMessage(evt) {
      if (!this.validateAutoMessage) {
        evt.preventDefault()
      }

      const obj = { ...this.models.autoMessage }

      if (this.models.autoMessage.sendMode === 'cron') {
        delete obj.message_interval
      } else if (this.models.autoMessage.sendMode === 'lines') {
        delete obj.cron
      }

      if (obj.uuid) {
        axios.put(`config-editor/auto-messages/${obj.uuid}`, obj, this.axiosOptions)
          .catch(err => this.handleFetchError(err))
          .then(() => window.setTimeout(() => this.reload(), 1000))
      } else {
        axios.post(`config-editor/auto-messages`, obj, this.axiosOptions)
          .catch(err => this.handleFetchError(err))
          .then(() => window.setTimeout(() => this.reload(), 1000))
      }
    },

    updateGeneralConfig() {
      axios.put('config-editor/general', this.generalConfig, this.axiosOptions)
        .catch(err => this.handleFetchError(err))
        .then(() => window.setTimeout(() => this.reload(), 1000))
    },
  },

  mounted() {
    this.fetchVars()

    const params = new URLSearchParams(window.location.hash.substring(1))
    this.authToken = params.get('access_token') || null

    if (this.authToken) {
      window.history.replaceState(null, '', window.location.href.split('#')[0])
      this.reload()
    }
  },

  watch: {},
})
