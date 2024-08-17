<template>
  <div class="container my-3">
    <div class="row justify-content-center mb-3">
      <div class="col-6">
        <div class="input-group">
          <span class="input-group-text">
            <i class="fas fa-user fa-fw me-1" />
          </span>
          <input
            v-model="inputAddEditor.text"
            type="text"
            :class="inputAddEditorClasses"
            @keypress.enter="addEditor"
          >
          <button
            class="btn btn-success"
            :disabled="!inputAddEditor.valid"
            @click="addEditor"
          >
            <i class="fas fa-plus fa-fw me-1" />
            {{ $t('editors.btnAdd') }}
          </button>
        </div>
      </div>
    </div>
    <div class="row justify-content-center">
      <div
        v-for="editor in editors"
        :key="editor.id"
        class="col-2"
      >
        <div class="card relative">
          <div class="card-body text-center">
            <p>
              <img
                :src="editor.profile_image_url"
                class="img rounded-circle w-50"
              >
            </p>
            <p class="mb-0">
              <code>{{ editor.display_name }}</code>
            </p>
          </div>
          <button
            class="btn btn-danger btn-sm editor-delete"
            :disabled="!editor.canDelete"
            @click="removeEditor(editor)"
          >
            <i class="fas fa-trash fa-fw" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import BusEventTypes from '../helpers/busevents'
import { defineComponent } from 'vue'
import { successToast } from '../helpers/toasts'

export default defineComponent({
  computed: {
    editors(): Array<any> {
      return (this.generalConfig.bot_editors || [])
        .filter((user: string) => this.profiles[user])
        .map((user: string) => ({
          ...this.profiles[user],
          canDelete: this.$root?.userInfo?.id !== this.profiles[user].id,
        }))
        .sort((a: any, b: any) => a.login.localeCompare(b.login))
    },

    inputAddEditorClasses(): string {
      const classes = ['form-control']

      if (this.inputAddEditor.valid) {
        classes.push('is-valid')
      } else if (this.inputAddEditor.text) {
        classes.push('is-invalid')
      }

      return classes.join(' ')
    },
  },

  data() {
    return {
      generalConfig: {} as any,
      inputAddEditor: {
        text: '',
        valid: false,
      },

      profiles: {} as any,
    }
  },

  methods: {
    /**
     * Adds the editor entered into the input field to the list
     */
    addEditor(): Promise<void> | undefined {
      if (!this.inputAddEditor.valid) {
        return
      }

      const editor = this.inputAddEditor.text

      return this.updateGeneralConfig({
        ...this.generalConfig,
        bot_editors: [
          ...this.generalConfig.bot_editors.filter((user: string) => user !== editor),
          editor,
        ],
      })
        ?.then(() => {
          this.inputAddEditor.text = ''
          this.bus.emit(BusEventTypes.Toast, successToast(this.$t('editors.toastEditorAdded')))
        })
    },

    /**
     * Fetches the general config object from the backend
     */
    fetchGeneralConfig(): Promise<void> | undefined {
      return this.$root?.fetchJSON('config-editor/general')
        .then((data: any) => {
          this.generalConfig = data
          data.bot_editors.forEach((editor: string) => this.fetchProfile(editor))
        })
    },

    /**
     * Fetches a profile for the given user(-id)
     */
    fetchProfile(user: string): Promise<void> | undefined {
      return this.$root?.fetchJSON(`config-editor/user?user=${user}`)
        .then((data: any) => {
          this.profiles[user] = data
        })
    },

    removeEditor(editor: any): Promise<void> | undefined {
      return this.updateGeneralConfig({
        ...this.generalConfig,
        bot_editors: this.generalConfig.bot_editors
          .filter((user: string) => ![editor.login, editor.id, editor.display_name].includes(user)),
      })
        ?.then(() => {
          this.bus.emit(BusEventTypes.Toast, successToast(this.$t('editors.toastEditorRemoved')))
        })
    },

    /**
     * Writes general config back to backend
     *
     * @param config Configuration object to write (MUST contain all config)
     */
    updateGeneralConfig(config: any): Promise<void> | undefined {
      return this.$root?.fetchJSON('config-editor/general', {
        body: JSON.stringify(config),
        method: 'PUT',
      })
    },
  },

  mounted() {
  // Reload config after it changed
    this.bus.on(BusEventTypes.ConfigReload, () => this.fetchGeneralConfig())

    // Do initial fetches
    this.fetchGeneralConfig()
  },

  name: 'TwitchBotEditorBotEditors',

  watch: {
    'inputAddEditor.text'(to) {
      this.inputAddEditor.valid = to.match(/^[a-zA-Z0-9_]{4,25}$/)
    },
  },
})
</script>

<style scoped>
.editor-delete {
  position: absolute;
  right: 5px;
  top: 5px;
}
</style>
