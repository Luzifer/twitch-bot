<template>
  <div>
    <b-row>
      <b-col>
        <b-card-group columns>
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

          <b-card no-body>
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

          <b-card no-body>
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
        </b-card-group>
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

    fetchAPITokens() {
      return axios.get('config-editor/auth-tokens', this.$root.axiosOptions)
        .then(resp => {
          this.apiTokens = resp.data
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    fetchGeneralConfig() {
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
      return axios.get('config-editor/modules')
        .then(resp => {
          this.modules = resp.data
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    fetchProfile(user) {
      return axios.get(`config-editor/user?user=${user}`, this.$root.axiosOptions)
        .then(resp => Vue.set(this.userProfiles, user, resp.data))
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
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING)
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
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING)
          window.setTimeout(() => {
            this.createdAPIToken = null
          }, 30000)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    updateGeneralConfig() {
      axios.put('config-editor/general', this.generalConfig, this.$root.axiosOptions)
        .then(() => {
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },
  },

  mounted() {
    this.$bus.$on(constants.NOTIFY_CONFIG_RELOAD, () => {
      this.fetchGeneralConfig()
      this.fetchAPITokens()
    })

    this.fetchGeneralConfig()
    this.fetchAPITokens()
    this.fetchModules()
  },

  name: 'TwitchBotEditorAppGeneralConfig',
}
</script>
