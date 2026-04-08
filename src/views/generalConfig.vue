<template>
  <div>
    <div class="row">
      <div class="col">
        <div class="card">
          <div class="card-header">
            <font-awesome-icon
              fixed-width
              class="me-1"
              :icon="['fas', 'hashtag']"
            />
            Channels
          </div>
          <div class="list-group list-group-flush">
            <div
              v-for="channel in sortedChannels"
              :key="channel"
              class="list-group-item d-flex align-items-center align-middle"
            >
              <font-awesome-icon
                fixed-width
                class="me-1"
                :icon="['fas', 'hashtag']"
              />
              {{ channel }}
              <span class="ms-auto me-2">
                <font-awesome-icon
                  v-if="!generalConfig.channel_has_token[channel]"
                  :id="`channelPublicWarn${channel}`"
                  fixed-width
                  class="ms-1 text-danger"
                  :icon="['fas', 'exclamation-triangle']"
                />
                <font-awesome-icon
                  v-else-if="!hasAllExtendedScopes(channel)"
                  :id="`channelPublicWarn${channel}`"
                  fixed-width
                  class="ms-1 text-warning"
                  :icon="['fas', 'exclamation-triangle']"
                />
                <AppTooltip
                  :target="`channelPublicWarn${channel}`"
                  triggers="hover"
                >
                  <template v-if="!generalConfig.channel_has_token[channel]">
                    Bot is not authorized to access Twitch on behalf of this channels owner (tokens are missing).
                    Click pencil to grant permissions.
                  </template>
                  <template v-else>
                    Channel is missing {{ missingExtendedScopes(channel).length }} extended permissions.
                    Click pencil to change granted permissions.
                  </template>
                </AppTooltip>
              </span>
              <div class="btn-group btn-group-sm">
                <button
                  type="button"
                  class="btn btn-primary"
                  @click="editChannelPermissions(channel)"
                >
                  <font-awesome-icon
                    fixed-width
                    :icon="['fas', 'pencil-alt']"
                  />
                </button>
                <button
                  type="button"
                  class="btn btn-danger"
                  @click="removeChannel(channel)"
                >
                  <font-awesome-icon
                    fixed-width
                    :icon="['fas', 'minus']"
                  />
                </button>
              </div>
            </div>

            <div class="list-group-item">
              <div class="input-group input-group-sm">
                <span class="input-group-text">
                  <font-awesome-icon
                    class="me-1"
                    :icon="['fas', 'hashtag']"
                  />
                </span>
                <input
                  v-model="models.addChannel"
                  class="form-control"
                  :class="{
                    'is-invalid': models.addChannel !== '' && !validateUserName(models.addChannel),
                    'is-valid': !!validateUserName(models.addChannel),
                  }"
                  type="text"
                  @keyup.enter="addChannel"
                >
                <button
                  type="button"
                  class="btn btn-success"
                  :disabled="!validateUserName(models.addChannel)"
                  @click="addChannel"
                >
                  <font-awesome-icon
                    fixed-width
                    class="me-1"
                    :icon="['fas', 'plus']"
                  />
                  Add
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="col">
        <div class="card mb-3">
          <div class="card-header">
            <font-awesome-icon
              fixed-width
              class="me-1"
              :icon="['fas', 'users']"
            />
            Bot-Editors
          </div>
          <div class="list-group list-group-flush">
            <div
              v-for="editor in sortedEditors"
              :key="editor"
              class="list-group-item d-flex align-items-center align-middle"
            >
              <img
                :src="userProfiles[editor] ? userProfiles[editor].profile_image_url : ''"
                style="height: 2rem; width: 2rem; object-fit: cover;"
                class="me-3 rounded-circle bg-secondary-subtle"
              >
              <span class="me-auto">{{ userProfiles[editor] ? userProfiles[editor].display_name : editor }}</span>
              <button
                type="button"
                class="btn btn-danger btn-sm"
                @click="removeEditor(editor)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'minus']"
                />
              </button>
            </div>

            <div class="list-group-item">
              <div class="input-group input-group-sm">
                <input
                  v-model="models.addEditor"
                  class="form-control"
                  :class="{
                    'is-invalid': models.addEditor !== '' && !validateUserName(models.addEditor),
                    'is-valid': !!validateUserName(models.addEditor),
                  }"
                  type="text"
                  @keyup.enter="addEditor"
                >
                <button
                  type="button"
                  class="btn btn-success"
                  :disabled="!validateUserName(models.addEditor)"
                  @click="addEditor"
                >
                  <font-awesome-icon
                    fixed-width
                    class="me-1"
                    :icon="['fas', 'plus']"
                  />
                  Add
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header d-flex align-items-center align-middle">
            <span class="me-auto">
              <font-awesome-icon
                fixed-width
                class="me-1"
                :icon="['fas', 'ticket-alt']"
              />
              Auth-Tokens
            </span>
            <div class="btn-group btn-group-sm">
              <button
                type="button"
                class="btn btn-success"
                @click="newAPIToken"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'plus']"
                />
              </button>
            </div>
          </div>
          <div class="list-group list-group-flush">
            <div
              v-if="createdAPIToken"
              class="list-group-item list-group-item-success"
            >
              Token was created, copy it within 30s as you will not see it again:<br>
              <code>{{ createdAPIToken.token }}</code>
            </div>

            <div
              v-for="(token, uuid) in apiTokens"
              :key="uuid"
              class="list-group-item d-flex align-items-center align-middle"
            >
              <span class="me-auto">
                {{ token.name }}<br>
                <span
                  v-for="module in token.modules"
                  :key="module"
                  class="badge bg-secondary-subtle text-secondary-emphasis me-1"
                >{{ module === '*' ? 'ANY' : module }}</span>
              </span>
              <button
                type="button"
                class="btn btn-danger btn-sm"
                @click="removeAPIToken(uuid)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'minus']"
                />
              </button>
            </div>
          </div>
        </div>
      </div>
      <div class="col">
        <div
          class="card mb-3"
          :class="botConnectionCardVariant ? `border-${botConnectionCardVariant}` : null"
        >
          <div
            class="card-header d-flex align-items-center align-middle"
            :class="botConnectionCardVariant ? `bg-${botConnectionCardVariant}` : null"
          >
            <span class="me-auto">
              <font-awesome-icon
                fixed-width
                class="me-1"
                :icon="['fas', 'sign-in-alt']"
              />
              Bot Connection
            </span>
            <template v-if="generalConfig.bot_name">
              <code
                id="botUserName"
              >
                {{ generalConfig.bot_name }}
                <AppTooltip
                  target="botUserName"
                  triggers="hover"
                >
                  Twitch Login-Name of the bot user currently authorized
                </AppTooltip>
              </code>
            </template>
            <template v-else>
              <font-awesome-icon
                id="botUserNameDC"
                fixed-width
                class="me-1 text-danger"
                :icon="['fas', 'unlink']"
              />
              <AppTooltip
                target="botUserNameDC"
                triggers="hover"
              >
                Bot is not currently authorized!
              </AppTooltip>
            </template>
          </div>

          <div class="card-body">
            <p>
              Here you can manage your bots auth-token: it's required to communicate with Twitch Chat and APIs. The access will be valid as long as you don't change the password or revoke the apps permission in your bot account.
            </p>
            <ul>
              <li>Copy the URL provided below</li>
              <li><strong>Open an incognito tab or different browser you are not logged into Twitch or are logged in with your bot account</strong></li>
              <li>Open the copied URL, sign in with the bot account and accept the permissions</li>
              <li>You will see a message containing the authorized account. If this account is wrong, just start over, the token will be overwritten.</li>
            </ul>
            <p
              v-if="botMissingScopes > 0"
              class="alert alert-warning"
            >
              <font-awesome-icon
                fixed-width
                class="me-1"
                :icon="['fas', 'exclamation-triangle']"
              />
              Bot is missing {{ botMissingScopes }} of its required scopes which will cause features not to work properly. Please re-authorize the bot using the URL below.
            </p>
            <div class="input-group input-group-sm">
              <input
                placeholder="Loading..."
                class="form-control"
                readonly
                :value="botAuthTokenURL"
                @focus="($event.currentTarget as HTMLInputElement).select()"
              >
              <button
                type="button"
                class="btn"
                :class="`btn-${copyButtonVariant.botConnection}`"
                @click="copyAuthURL('botConnection')"
              >
                <font-awesome-icon
                  fixed-width
                  class="me-1"
                  :icon="['fas', 'clipboard']"
                />
                Copy
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- API-Token Editor -->
    <AppModal
      v-if="showAPITokenEditModal"
      :ok-disabled="!validateAPIToken"
      ok-title="Save"
      size="md"
      :model-value="showAPITokenEditModal"
      title="New API-Token"
      @hidden="showAPITokenEditModal=false"
      @update:model-value="showAPITokenEditModal = $event"
      @ok="saveAPIToken"
    >
      <div class="mb-3">
        <label
          class="form-label"
          for="formAPITokenName"
        >Name</label>
        <input
          id="formAPITokenName"
          v-model="models.apiToken.name"
          class="form-control"
          :class="{
            'is-invalid': models.apiToken.name === '',
            'is-valid': Boolean(models.apiToken.name),
          }"
          type="text"
        >
      </div>

      <div class="mb-3">
        <label class="form-label">Enabled for Modules</label>
        <div class="mb-3">
          <div
            v-for="module in availableModules"
            :key="module.value"
            class="form-check mb-2"
          >
            <input
              :id="`api-token-module-${module.value}`"
              v-model="models.apiToken.modules"
              class="form-check-input"
              :value="module.value"
              type="checkbox"
            >
            <label
              class="form-check-label"
              :for="`api-token-module-${module.value}`"
            >{{ module.text }}</label>
          </div>
        </div>
      </div>
    </AppModal>

    <!-- Channel Permission Editor -->
    <AppModal
      v-if="showPermissionEditModal"
      hide-footer
      size="lg"
      title="Edit Permissions for Channel"
      :model-value="showPermissionEditModal"
      @hidden="showPermissionEditModal=false"
      @update:model-value="showPermissionEditModal = $event"
    >
      <div class="row">
        <div class="col">
          <p>The bot should be able to&hellip;</p>
          <div id="channelPermissions">
            <div
              v-for="permission in extendedPermissions"
              :key="permission.value"
              class="form-check form-switch mb-2"
            >
              <input
                :id="`channelPermissions-${permission.value}`"
                v-model="models.channelPermissions"
                class="form-check-input"
                :value="permission.value"
                type="checkbox"
              >
              <label
                class="form-check-label"
                :for="`channelPermissions-${permission.value}`"
              >{{ permission.text }}</label>
            </div>
          </div>
          <p class="mt-3">
            &hellip;on this channel.
          </p>
        </div>
        <div class="col">
          <p>
            In order to access non-public information as channel-point redemptions or take actions limited to the channel owner the bot needs additional permissions. The <strong>owner</strong> of the channel needs to grant those!
          </p>
          <ul>
            <li>Select permissions on the left side</li>
            <li>Copy the URL provided below</li>
            <li>Pass the URL to the channel owner and tell them to open it with their personal account logged in</li>
            <li>The bot will display a message containing the updated account</li>
          </ul>
        </div>
      </div>
      <div class="row">
        <div class="col">
          <div class="input-group">
            <input
              placeholder="Loading..."
              class="form-control"
              readonly
              :value="extendedPermissionsURL"
              @focus="($event.currentTarget as HTMLInputElement).select()"
            >
            <button
              type="button"
              class="btn"
              :class="`btn-${copyButtonVariant.channelPermission}`"
              @click="copyAuthURL('channelPermission')"
            >
              <font-awesome-icon
                fixed-width
                class="me-1"
                :icon="['fas', 'clipboard']"
              />
              Copy
            </button>
          </div>
        </div>
      </div>
    </AppModal>
  </div>
</template>

<script lang="ts">
import * as constants from '../lib/const'
import type { AuthTokensResponse, AuthURLsResponse, ConfigAuthToken, GeneralConfig, TwitchUser } from '../types'
import { api } from '../api'
import AppModal from '../components/AppModal.vue'
import AppTooltip from '../components/AppTooltip'
import { defineComponent } from 'vue'
import { useAppStore } from '../stores/app'

export default defineComponent({
  components: {
    AppModal,
    AppTooltip,
  },

  computed: {
    availableModules() {
      return [
        { text: 'ANY', value: '*' },
        ...[...this.modules || []].sort()
          .filter(m => m !== 'config-editor')
          .map(m => ({ text: m, value: m })),
      ]
    },

    botAuthTokenURL() {
      if (!this.authURLs || !this.authURLs.update_bot_token) {
        return ''
      }

      let scopes = [...this.appStore.vars.DefaultBotScopes || []]

      if (this.generalConfig && this.generalConfig.channel_scopes && this.generalConfig.bot_name && this.generalConfig.channel_scopes[this.generalConfig.bot_name]) {
        scopes = [
          ...new Set([
            ...scopes,
            ...this.generalConfig.channel_scopes[this.generalConfig.bot_name],
          ]),
        ]
      }

      const u = new URL(this.authURLs.update_bot_token)
      u.searchParams.set('scope', scopes.join(' '))
      return u.toString()
    },

    botConnectionCardVariant(): string {
      return this.appStore.status?.overall_status_success ? '' : 'warning'
    },

    botMissingScopes(): number {
      let missing = 0

      if (!this.generalConfig || !this.generalConfig.channel_scopes || !this.generalConfig.bot_name) {
        return -1
      }

      const grantedScopes = [...this.generalConfig.channel_scopes[this.generalConfig.bot_name] || []]

      for (const scope of this.appStore.vars.DefaultBotScopes || []) {
        if (!grantedScopes.includes(scope)) {
          missing++
        }
      }

      return missing
    },

    extendedPermissions() {
      return Object.keys(this.authURLs.available_extended_scopes || {})
        .map(v => ({ text: this.authURLs.available_extended_scopes[v], value: v }))
        .sort((a, b) => a.value.localeCompare(b.value))
    },

    extendedPermissionsURL() {
      if (!this.authURLs?.update_channel_scopes) {
        return null
      }

      const u = new URL(this.authURLs.update_channel_scopes)
      u.searchParams.set('scope', this.models.channelPermissions.join(' '))
      return u.toString()
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
      apiTokens: {} as AuthTokensResponse,
      appStore: useAppStore(),
      authURLs: {} as AuthURLsResponse,
      copyButtonVariant: {
        botConnection: 'primary',
        channelPermission: 'primary',
      },

      createdAPIToken: null as ConfigAuthToken | null,
      generalConfig: {
        bot_editors: [],
        channel_has_token: {},
        channel_scopes: {},
        channels: [],
      } as GeneralConfig,

      models: {
        addChannel: '',
        addEditor: '',
        apiToken: {} as Pick<ConfigAuthToken, 'modules' | 'name'>,
        channelPermissions: [] as string[],
      },

      modules: [] as string[],

      showAPITokenEditModal: false,
      showPermissionEditModal: false,
      userProfiles: {} as Record<string, TwitchUser>,
    }
  },

  methods: {
    addChannel() {
      if (!this.validateUserName(this.models.addChannel)) {
        return
      }

      this.generalConfig.channels.push(this.models.addChannel.replace(/^#*/, ''))
      this.models.addChannel = ''

      this.updateGeneralConfig()
    },

    addEditor() {
      if (!this.validateUserName(this.models.addEditor)) {
        return
      }

      this.fetchProfile(this.models.addEditor)
      this.generalConfig.bot_editors.push(this.models.addEditor)
      this.models.addEditor = ''

      this.updateGeneralConfig()
    },

    async copyAuthURL(type: 'botConnection' | 'channelPermission') {
      let prom: Promise<void> | null
      let btnField: 'botConnection' | 'channelPermission' | null = null

      switch (type) {
      case 'botConnection':
        prom = navigator.clipboard.writeText(this.botAuthTokenURL)
        btnField = 'botConnection'
        break
      case 'channelPermission':
        prom = navigator.clipboard.writeText(this.extendedPermissionsURL!)
        btnField = 'channelPermission'
        break
      default:
        throw new Error(`invalid auth-url type ${type}`)
      }

      try {
        try {
          await prom
          this.copyButtonVariant[btnField] = 'success'
        } catch {
          this.copyButtonVariant[btnField] = 'danger'
        }
      } finally {
        window.setTimeout(() => {
          this.copyButtonVariant[btnField] = 'primary'
        }, 2000)
      }
    },

    editChannelPermissions(channel: string) {
      let permissionSet = [...this.generalConfig.channel_scopes[channel] || []]

      if (channel === this.generalConfig.bot_name) {
        permissionSet = [
          ...permissionSet,
          ...this.appStore.vars.DefaultBotScopes,
        ]
      }

      this.models.channelPermissions = [...new Set(permissionSet)]
      this.showPermissionEditModal = true
    },

    async fetchAPITokens() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      try {
        const resp = await api.get<AuthTokensResponse>('config-editor/auth-tokens')
        this.apiTokens = resp || {}
      } catch (err) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    async fetchAuthURLs() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      try {
        const resp = await api.get<AuthURLsResponse>('config-editor/auth-urls')
        this.authURLs = resp || {} as AuthURLsResponse
      } catch (err) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    async fetchGeneralConfig() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)

      let resp: GeneralConfig | undefined
      try {
        resp = await api.get<GeneralConfig>('config-editor/general')
      } catch (err) {
        this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }

      if (!resp) {
        return []
      }

      this.generalConfig = resp
      const promises = []
      for (const editor of this.generalConfig.bot_editors) {
        promises.push(this.fetchProfile(editor))
      }
      return Promise.all(promises)
    },

    async fetchModules() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      try {
        const resp = await api.get<string[]>('config-editor/modules', false)
        this.modules = resp || []
      } catch (err) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    async fetchProfile(user: string) {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      try {
        const resp = await api.get<TwitchUser>(`config-editor/user?user=${user}`)
        if (resp) {
          this.userProfiles[user] = resp
        }
        this.$bus.$emit(constants.NOTIFY_LOADING_DATA, false)
      } catch (err) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    hasAllExtendedScopes(channel: string) {
      if (!this.generalConfig.channel_scopes[channel]) {
        return false
      }

      for (const scope in this.authURLs.available_extended_scopes) {
        if (!this.generalConfig.channel_scopes[channel].includes(scope)) {
          return false
        }
      }

      return true
    },

    missingExtendedScopes(channel: string) {
      if (!this.generalConfig.channel_scopes[channel]) {
        return Object.keys(this.authURLs.available_extended_scopes || {})
      }

      const missing = []

      for (const scope in this.authURLs.available_extended_scopes) {
        if (!this.generalConfig.channel_scopes[channel].includes(scope)) {
          missing.push(scope)
        }
      }

      return missing
    },

    newAPIToken() {
      this.models.apiToken = {
        modules: [],
        name: '',
      }
      this.showAPITokenEditModal = true
    },

    removeAPIToken(uuid: string) {
      api.delete(`config-editor/auth-tokens/${uuid}`)
        .then(() => {
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    removeChannel(channel: string) {
      this.generalConfig.channels = this.generalConfig.channels
        .filter(ch => ch !== channel)

      this.updateGeneralConfig()
    },

    removeEditor(editor: string) {
      this.generalConfig.bot_editors = this.generalConfig.bot_editors
        .filter(ed => ed !== editor)

      this.updateGeneralConfig()
    },

    saveAPIToken(evt: Event) {
      if (!this.validateAPIToken) {
        evt.preventDefault()
        return
      }

      api.post<ConfigAuthToken>(`config-editor/auth-tokens`, this.models.apiToken)
        .then(resp => {
          this.createdAPIToken = resp || null
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
          window.setTimeout(() => {
            this.createdAPIToken = null
          }, 30000)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    updateGeneralConfig() {
      api.put('config-editor/general', this.generalConfig)
        .then(() => {
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    validateUserName(user: string) {
      return user.match(constants.REGEXP_USER)
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

  name: 'TwitchBotGeneralConfigView',
})
</script>
