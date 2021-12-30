<template>
  <div>
    <b-row>
      <b-col>
        <b-card no-body>
          <b-card-header>
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'hashtag']"
            />
            Channels
          </b-card-header>
          <b-list-group flush>
            <b-list-group-item
              v-for="channel in sortedChannels"
              :key="channel"
              class="d-flex align-items-center align-middle"
            >
              <span class="mr-auto">
                <font-awesome-icon
                  fixed-width
                  class="mr-1"
                  :icon="['fas', 'hashtag']"
                />
                {{ channel }}
              </span>
              <b-button
                size="sm"
                variant="danger"
                @click="removeChannel(channel)"
              >
                <font-awesome-icon
                  fixed-width
                  class="mr-1"
                  :icon="['fas', 'minus']"
                />
              </b-button>
            </b-list-group-item>

            <b-list-group-item>
              <b-input-group>
                <b-form-input
                  v-model="models.addChannel"
                  @keyup.enter="addChannel"
                />
                <b-input-group-append>
                  <b-button
                    variant="success"
                    @click="addChannel"
                  >
                    <font-awesome-icon
                      fixed-width
                      class="mr-1"
                      :icon="['fas', 'plus']"
                    />
                    Add
                  </b-button>
                </b-input-group-append>
              </b-input-group>
            </b-list-group-item>
          </b-list-group>
        </b-card>
      </b-col>
      <b-col>
        <b-card
          no-body
          class="mb-3"
        >
          <b-card-header>
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'users']"
            />
            Bot-Editors
          </b-card-header>
          <b-list-group flush>
            <b-list-group-item
              v-for="editor in sortedEditors"
              :key="editor"
              class="d-flex align-items-center align-middle"
            >
              <b-avatar
                class="mr-3"
                :src="userProfiles[editor] ? userProfiles[editor].profile_image_url : ''"
              />
              <span class="mr-auto">{{ userProfiles[editor] ? userProfiles[editor].display_name : editor }}</span>
              <b-button
                size="sm"
                variant="danger"
                @click="removeEditor(editor)"
              >
                <font-awesome-icon
                  fixed-width
                  class="mr-1"
                  :icon="['fas', 'minus']"
                />
              </b-button>
            </b-list-group-item>

            <b-list-group-item>
              <b-input-group>
                <b-form-input
                  v-model="models.addEditor"
                  @keyup.enter="addEditor"
                />
                <b-input-group-append>
                  <b-button
                    variant="success"
                    @click="addEditor"
                  >
                    <font-awesome-icon
                      fixed-width
                      class="mr-1"
                      :icon="['fas', 'plus']"
                    />
                    Add
                  </b-button>
                </b-input-group-append>
              </b-input-group>
            </b-list-group-item>
          </b-list-group>
        </b-card>

        <b-card
          no-body
        >
          <b-card-header
            class="d-flex align-items-center align-middle"
          >
            <span class="mr-auto">
              <font-awesome-icon
                fixed-width
                class="mr-1"
                :icon="['fas', 'ticket-alt']"
              />
              Auth-Tokens
            </span>
            <b-button-group size="sm">
              <b-button
                variant="success"
                @click="newAPIToken"
              >
                <font-awesome-icon
                  fixed-width
                  class="mr-1"
                  :icon="['fas', 'plus']"
                />
              </b-button>
            </b-button-group>
          </b-card-header>
          <b-list-group flush>
            <b-list-group-item
              v-if="createdAPIToken"
              variant="success"
            >
              Token was created, copy it within 30s as you will not see it again:<br>
              <code>{{ createdAPIToken.token }}</code>
            </b-list-group-item>

            <b-list-group-item
              v-for="(token, uuid) in apiTokens"
              :key="uuid"
              class="d-flex align-items-center align-middle"
            >
              <span class="mr-auto">
                {{ token.name }}<br>
                <b-badge
                  v-for="module in token.modules"
                  :key="module"
                >{{ module === '*' ? 'ANY' : module }}</b-badge>
              </span>
              <b-button
                size="sm"
                variant="danger"
                @click="removeAPIToken(uuid)"
              >
                <font-awesome-icon
                  fixed-width
                  class="mr-1"
                  :icon="['fas', 'minus']"
                />
              </b-button>
            </b-list-group-item>
          </b-list-group>
        </b-card>
      </b-col>
      <b-col>
        <b-card
          no-body
          class="mb-3"
        >
          <b-card-header>
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'sign-in-alt']"
            />
            Bot Connection
          </b-card-header>

          <b-card-body>
            <p>
              Here you can manage your bots auth-token: it is using to communicate with chat and Twitch APIs. This will override the token you might have provided when starting the bot and will be automatically renewed as long as you don't change your password or revoke the apps permission on your bot account.
            </p>
            <ul>
              <li>Copy the URL provided below</li>
              <li>Open an inkognito tab or different browser you are not logged into Twitch or are logged in with your bot account</li>
              <li>Open the copied URL, sign in with the bot account and accept the permissions</li>
              <li>The bot will display a message containing the authorized account. If this account is wrong, just start over, the token will be overwritten.</li>
            </ul>
            <b-input-group>
              <b-form-input
                placeholder="Loading..."
                readonly
                :value="authURLs.update_bot_token"
              />
              <b-input-group-append>
                <b-button
                  :variant="copyButtonVariant.botConnection"
                  @click="copyAuthURL('botConnection')"
                >
                  <font-awesome-icon
                    fixed-width
                    class="mr-1"
                    :icon="['fas', 'clipboard']"
                  />
                  Copy
                </b-button>
              </b-input-group-append>
            </b-input-group>
          </b-card-body>
        </b-card>

        <b-card
          no-body
          class="mb-3"
        >
          <b-card-header>
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'sign-in-alt']"
            />
            Channel Permissions
          </b-card-header>

          <b-card-body>
            <p>
              In order to access non-public information as channelpoint redemptions the bot needs additional permissions. The <strong>owner</strong> of the channel needs to grant those!
            </p>
            <ul>
              <li>Copy the URL provided below</li>
              <li>Pass the URL to the channel owner and tell them to open it with their personal account logged in</li>
              <li>The bot will display a message containing the updated account</li>
            </ul>
            <b-input-group>
              <b-form-input
                placeholder="Loading..."
                readonly
                :value="authURLs.update_channel_scopes"
              />
              <b-input-group-append>
                <b-button
                  :variant="copyButtonVariant.channelPermission"
                  @click="copyAuthURL('channelPermission')"
                >
                  <font-awesome-icon
                    fixed-width
                    class="mr-1"
                    :icon="['fas', 'clipboard']"
                  />
                  Copy
                </b-button>
              </b-input-group-append>
            </b-input-group>
          </b-card-body>
        </b-card>
      </b-col>
    </b-row>

    <!-- API-Token Editor -->
    <b-modal
      v-if="showAPITokenEditModal"
      hide-header-close
      :ok-disabled="!validateAPIToken"
      ok-title="Save"
      size="md"
      :visible="showAPITokenEditModal"
      title="New API-Token"
      @hidden="showAPITokenEditModal=false"
      @ok="saveAPIToken"
    >
      <b-form-group
        label="Name"
        label-for="formAPITokenName"
      >
        <b-form-input
          id="formAPITokenName"
          v-model="models.apiToken.name"
          :state="Boolean(models.apiToken.name)"
          type="text"
        />
      </b-form-group>

      <b-form-group
        label="Enabled for Modules"
      >
        <b-form-checkbox-group
          v-model="models.apiToken.modules"
          class="mb-3"
          :options="availableModules"
          text-field="text"
          value-field="value"
        />
      </b-form-group>
    </b-modal>
  </div>
</template>

<script>
import * as constants from './const.js'

import axios from 'axios'
import Vue from 'vue'

export default {
  computed: {
    availableModules() {
      return [
        { text: 'ANY', value: '*' },
        ...[...this.modules || []].sort()
          .filter(m => m !== 'config-editor')
          .map(m => ({ text: m, value: m })),
      ]
    },

    sortedChannels() {
      return [...this.generalConfig?.channels || []].sort((a, b) => a.toLocaleLowerCase().localeCompare(b.toLocaleLowerCase()))
    },

    sortedEditors() {
      return [...this.generalConfig?.bot_editors || []].sort((a, b) => {
        const an = this.userProfiles[a]?.login || a
        const bn = this.userProfiles[b]?.login || b

        return an.localeCompare(bn)
      })
    },

    validateAPIToken() {
      return this.models.apiToken.modules.length > 0 && Boolean(this.models.apiToken.name)
    },
  },

  data() {
    return {
      apiTokens: {},
      authURLs: {},
      copyButtonVariant: {
        botConnection: 'primary',
        channelPermission: 'primary',
      },

      createdAPIToken: null,
      generalConfig: {},
      models: {
        addChannel: '',
        addEditor: '',
        apiToken: {},
      },

      modules: [],

      showAPITokenEditModal: false,
      userProfiles: {},
    }
  },

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

    copyAuthURL(type) {
      let prom = null
      let btnField = null

      switch (type) {
      case 'botConnection':
        prom = navigator.clipboard.writeText(this.authURLs.update_bot_token)
        btnField = 'botConnection'
        break
      case 'channelPermission':
        prom = navigator.clipboard.writeText(this.authURLs.update_channel_scopes)
        btnField = 'channelPermission'
        break
      }

      return prom
        .then(() => {
          this.copyButtonVariant[btnField] = 'success'
        })
        .catch(() => {
          this.copyButtonVariant[btnField] = 'danger'
        })
        .finally(() => {
          window.setTimeout(() => {
            this.copyButtonVariant[btnField] = 'primary'
          }, 2000)
        })
    },

    fetchAPITokens() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      return axios.get('config-editor/auth-tokens', this.$root.axiosOptions)
        .then(resp => {
          this.apiTokens = resp.data
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    fetchAuthURLs() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      return axios.get('config-editor/auth-urls', this.$root.axiosOptions)
        .then(resp => {
          this.authURLs = resp.data
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    fetchGeneralConfig() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      return axios.get('config-editor/general', this.$root.axiosOptions)
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        .then(resp => {
          this.generalConfig = resp.data

          const promises = []
          for (const editor of this.generalConfig.bot_editors) {
            promises.push(this.fetchProfile(editor))
          }

          return Promise.all(promises)
        })
    },

    fetchModules() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      return axios.get('config-editor/modules')
        .then(resp => {
          this.modules = resp.data
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    fetchProfile(user) {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      return axios.get(`config-editor/user?user=${user}`, this.$root.axiosOptions)
        .then(resp => {
          Vue.set(this.userProfiles, user, resp.data)
          this.$bus.$emit(constants.NOTIFY_LOADING_DATA, false)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    newAPIToken() {
      Vue.set(this.models, 'apiToken', {
        modules: [],
        name: '',
      })
      this.showAPITokenEditModal = true
    },

    removeAPIToken(uuid) {
      axios.delete(`config-editor/auth-tokens/${uuid}`, this.$root.axiosOptions)
        .then(() => {
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
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

    saveAPIToken(evt) {
      if (!this.validateAPIToken) {
        evt.preventDefault()
        return
      }

      axios.post(`config-editor/auth-tokens`, this.models.apiToken, this.$root.axiosOptions)
        .then(resp => {
          this.createdAPIToken = resp.data
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
          window.setTimeout(() => {
            this.createdAPIToken = null
          }, 30000)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    updateGeneralConfig() {
      axios.put('config-editor/general', this.generalConfig, this.$root.axiosOptions)
        .then(() => {
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },
  },

  mounted() {
    this.$bus.$on(constants.NOTIFY_CONFIG_RELOAD, () => {
      Promise.all([
        this.fetchGeneralConfig(),
        this.fetchAPITokens(),
        this.fetchAuthURLs(),
      ]).then(() => {
        this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, false)
        this.$bus.$emit(constants.NOTIFY_LOADING_DATA, false)
      })
    })

    Promise.all([
      this.fetchGeneralConfig(),
      this.fetchAPITokens(),
      this.fetchAuthURLs(),
      this.fetchModules(),
    ]).then(() => this.$bus.$emit(constants.NOTIFY_LOADING_DATA, false))
  },

  name: 'TwitchBotEditorAppGeneralConfig',
}
</script>
