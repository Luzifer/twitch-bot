/* global axios, Vue */

/* eslint-disable camelcase --- We are working with data from a Go JSON API */

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

    availableActionsForAdd() {
      return this.actions.map(a => ({ text: a.name, value: a.type }))
    },

    availableEvents() {
      return [
        { text: 'Clear Event-Matching', value: null },
        ...this.vars.KnownEvents,
      ]
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


    validateRule() {
      if (!this.models.rule.match_message__validation) {
        this.validateReason = 'rule.match_message__validation'
        return false
      }

      if (!this.validateDuration(this.models.rule.cooldown, false)) {
        this.validateReason = 'rule.cooldown'
        return false
      }

      if (!this.validateDuration(this.models.rule.user_cooldown, false)) {
        this.validateReason = 'rule.user_cooldown'
        return false
      }

      if (!this.validateDuration(this.models.rule.channel_cooldown, false)) {
        this.validateReason = 'rule.channel_cooldown'
        return false
      }

      for (const action of this.models.rule.actions || []) {
        const def = this.getActionDefinitionByType(action.type)
        if (!def) {
          this.validateReason = `nodef: ${action.type}`
          return false
        }

        if (!def.fields) {
          // No fields to check
          continue
        }

        for (const field of def.fields) {
          if (!field.optional && !action.attributes[field.key]) {
            this.validateReason = `${action.type} -> ${field.key} -> opt`
            return false
          }
        }
      }

      return true
    },
  },

  data: {
    actions: [],
    authToken: null,
    autoMessageFields: [
      {
        class: 'col-1 text-nowrap',
        key: 'channel',
        sortable: true,
        thClass: 'align-middle',
      },
      {
        class: 'col-9',
        key: 'message',
        sortable: true,
        thClass: 'align-middle',
      },
      {
        class: 'col-1 text-nowrap',
        key: 'cron',
        thClass: 'align-middle',
      },
      {
        class: 'col-1 text-right',
        key: 'actions',
        label: '',
        thClass: 'align-middle',
      },
    ],

    autoMessageSendModes: [
      { text: 'Cron', value: 'cron' },
      { text: 'Number of lines', value: 'lines' },
    ],

    autoMessages: [],

    editMode: 'general',
    error: null,
    generalConfig: {},
    models: {
      addAction: '',
      addChannel: '',
      addEditor: '',
      autoMessage: {},
      rule: {},
    },

    rules: [],
    rulesFields: [
      {
        class: 'col-3',
        key: '_match',
        label: 'Match',
        thClass: 'align-middle',
      },
      {
        class: 'col-8',
        key: '_description',
        label: 'Description',
        thClass: 'align-middle',
      },
      {
        class: 'col-1 text-right',
        key: '_actions',
        label: '',
        thClass: 'align-middle',
      },
    ],

    showAutoMessageEditModal: false,
    showRuleEditModal: false,
    userProfiles: {},
    validateReason: null,
    vars: {},
  },

  el: '#app',

  methods: {
    addAction() {
      const attributes = {}

      for (const field of this.getActionDefinitionByType(this.models.addAction).fields || []) {
        let defaultValue = null

        switch (field.type) {
        case 'bool':
          defaultValue = field.default === 'true'
          break
        case 'int64':
          defaultValue = field.default ? Number(field.default) : 0
          break
        case 'string':
          defaultValue = field.default
          break
        case 'stringslice':
          defaultValue = []
          break
        }

        attributes[field.key] = defaultValue
      }

      if (!this.models.rule.actions) {
        Vue.set(this.models.rule, 'actions', [])
      }

      this.models.rule.actions.push({ attributes, type: this.models.addAction })
    },

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

    delayedReload() {
      window.setTimeout(() => this.reload(), 1000)
    },

    deleteAutoMessage(uuid) {
      axios.delete(`config-editor/auto-messages/${uuid}`, this.axiosOptions)
        .then(() => this.delayedReload())
        .catch(err => this.handleFetchError(err))
    },

    deleteRule(uuid) {
      axios.delete(`config-editor/rules/${uuid}`, this.axiosOptions)
        .then(() => this.delayedReload())
        .catch(err => this.handleFetchError(err))
    },

    editAutoMessage(msg) {
      Vue.set(this.models, 'autoMessage', {
        ...msg,
        sendMode: msg.cron ? 'cron' : 'lines',
      })
      this.showAutoMessageEditModal = true
    },

    editRule(msg) {
      Vue.set(this.models, 'rule', {
        ...msg,
      })
      this.showRuleEditModal = true
      this.validateMatcherRegex()
    },

    fetchActions() {
      axios.get('config-editor/actions')
        .then(resp => {
          this.actions = resp.data
        })
        .catch(err => this.handleFetchError(err))
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

    formatRuleActions(rule) {
      const badges = []

      for (const action of rule.actions || []) {
        for (const actionDefinition of this.actions) {
          if (actionDefinition.type !== action.type) {
            continue
          }

          badges.push(actionDefinition.name)
        }
      }

      return badges
    },

    formatRuleMatch(rule) {
      const badges = []

      if (rule.match_channels) {
        badges.push({ key: 'Channels', value: rule.match_channels.join(', ') })
      }

      if (rule.match_event) {
        badges.push({ key: 'Event', value: rule.match_event })
      }

      if (rule.match_message) {
        badges.push({ key: 'Message', value: rule.match_message })
      }

      if (rule.match_users) {
        badges.push({ key: 'Users', value: rule.match_users.join(', ') })
      }

      return badges
    },

    getActionDefinitionByType(type) {
      for (const ad of this.actions) {
        if (ad.type === type) {
          return ad
        }
      }

      return null
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

    moveAction(idx, direction) {
      const tmp = [...this.models.rule.actions]

      const eltmp = tmp[idx]
      tmp[idx] = tmp[idx + direction]
      tmp[idx + direction] = eltmp

      Vue.set(this.models.rule, 'actions', tmp)
    },

    newAutoMessage() {
      Vue.set(this.models, 'autoMessage', {})
      this.showAutoMessageEditModal = true
    },

    newRule() {
      Vue.set(this.models, 'rule', {})
      this.showRuleEditModal = true
    },

    reload() {
      this.fetchAutoMessages()
      this.fetchGeneralConfig()
      this.fetchRules()
    },

    removeAction(idx) {
      this.models.rule.actions = this.models.rule.actions.filter((_, i) => i !== idx)
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
          .then(() => this.delayedReload())
      } else {
        axios.post(`config-editor/auto-messages`, obj, this.axiosOptions)
          .catch(err => this.handleFetchError(err))
          .then(() => this.delayedReload())
      }
    },

    saveRule(evt) {
      if (!this.validateRule) {
        evt.preventDefault()
      }

      const obj = { ...this.models.rule }

      if (obj.uuid) {
        axios.put(`config-editor/rules/${obj.uuid}`, obj, this.axiosOptions)
          .catch(err => this.handleFetchError(err))
          .then(() => this.delayedReload())
      } else {
        axios.post(`config-editor/rules`, obj, this.axiosOptions)
          .catch(err => this.handleFetchError(err))
          .then(() => this.delayedReload())
      }
    },

    updateGeneralConfig() {
      axios.put('config-editor/general', this.generalConfig, this.axiosOptions)
        .catch(err => this.handleFetchError(err))
        .then(() => this.delayedReload())
    },

    validateDuration(duration, required) {
      if (!duration && !required) {
        return true
      }

      return Boolean(duration.match(/(?:\d+(?:s|m|h))+/))
    },

    validateMatcherRegex() {
      if (this.models.rule.match_message === '') {
        Vue.set(this.models.rule, 'match_message__validation', true)
        return
      }

      return axios.put(`config-editor/validate-regex?regexp=${encodeURIComponent(this.models.rule.match_message)}`)
        .then(() => {
          Vue.set(this.models.rule, 'match_message__validation', true)
        })
        .catch(() => {
          Vue.set(this.models.rule, 'match_message__validation', false)
        })
    },

    validateTwitchBadge(tag) {
      return this.vars.IRCBadges.includes(tag)
    },
  },

  mounted() {
    this.fetchVars()
    this.fetchActions()

    const params = new URLSearchParams(window.location.hash.substring(1))
    this.authToken = params.get('access_token') || null

    if (this.authToken) {
      window.history.replaceState(null, '', window.location.href.split('#')[0])
      this.reload()
    }
  },

  name: 'ConfigEditor',

  watch: {
    'models.rule.match_message'(to, from) {
      if (to === from) {
        return
      }

      this.validateMatcherRegex()
    },
  },
})
