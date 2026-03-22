<template>
  <div>
    <div class="row">
      <div class="col">
        <div class="table-responsive">
          <table class="table table-striped table-hover align-middle">
            <thead>
              <tr>
                <th>Channel</th>
                <th>Message</th>
                <th>Cron</th>
                <th class="text-end">
                  <button
                    class="btn btn-success btn-sm"
                    @click="newAutoMessage"
                  >
                    <fa-icon
                      fixed-width
                      :icon="['fas', 'plus']"
                    />
                  </button>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!autoMessages.length">
                <td
                  colspan="4"
                  class="text-center text-muted"
                >
                  No auto-messages configured.
                </td>
              </tr>
              <tr
                v-for="item in autoMessages"
                :key="item.uuid"
              >
                <td>
                  <fa-icon
                    fixed-width
                    class="me-1"
                    :icon="['fas', 'hashtag']"
                  />
                  {{ item.channel }}
                </td>
                <td>
                  {{ item.message }}<br>
                  <span
                    v-if="item.disable"
                    class="badge text-bg-danger mt-1 me-1"
                  >
                    Disabled
                  </span>
                  <span
                    v-if="item.disable_on_template"
                    class="badge bg-secondary-subtle text-secondary-emphasis mt-1 me-1"
                  >
                    Disable on Template
                  </span>
                  <span
                    v-if="item.only_on_live"
                    class="badge bg-secondary-subtle text-secondary-emphasis mt-1 me-1"
                  >
                    Only during Stream
                  </span>
                </td>
                <td><code>{{ item.cron }}</code></td>
                <td class="text-end text-nowrap">
                  <div class="btn-group btn-group-sm">
                    <button
                      class="btn btn-outline-secondary"
                      @click="editAutoMessage(item)"
                    >
                      <fa-icon
                        fixed-width
                        :icon="['fas', 'pen']"
                      />
                    </button>
                    <button
                      class="btn btn-danger"
                      @click="deleteAutoMessage(item.uuid!)"
                    >
                      <fa-icon
                        fixed-width
                        :icon="['fas', 'minus']"
                      />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <AppModal
      v-model="showAutoMessageEditModal"
      :ok-disabled="!validateAutoMessage"
      ok-title="Save"
      size="lg"
      title="Edit Auto-Message"
      @ok="saveAutoMessage"
    >
      <div class="row">
        <div class="col-lg-8">
          <div class="mb-3">
            <label
              class="form-label"
              for="formAutoMessageChannel"
            >
              Channel
            </label>
            <div class="input-group">
              <span class="input-group-text">#</span>
              <input
                id="formAutoMessageChannel"
                v-model="models.autoMessage.channel"
                class="form-control"
                :class="fieldStateClass(validateAutoMessageChannel)"
                type="text"
              >
            </div>
          </div>

          <hr>

          <div class="mb-3">
            <label
              class="form-label"
              for="formAutoMessageMessage"
            >
              Message
            </label>
            <template-editor
              id="formAutoMessageMessage"
              v-model="models.autoMessage.message"
              :state="Boolean(models.autoMessage.message)"
              @valid-template="(valid: boolean) => updateTemplateValid('autoMessage.message', valid)"
            />
            <div class="form-text">
              <fa-icon
                fixed-width
                class="me-1 text-success"
                :icon="['fas', 'code']"
              />
              Ensure the template result has a length of less than {{ validateAutoMessageMessageLength }} characters.
            </div>
          </div>

          <div class="form-check form-switch mb-3">
            <input
              id="formAutoMessageAction"
              v-model="models.autoMessage.use_action"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="formAutoMessageAction"
            >
              Send message as action (<code>/me</code>)
            </label>
          </div>

          <hr>

          <div class="mb-3">
            <label
              class="form-label"
              for="formAutoMessageSendMode"
            >
              Sending Mode
            </label>
            <select
              id="formAutoMessageSendMode"
              v-model="models.autoMessage.sendMode"
              class="form-select"
            >
              <option
                v-for="mode in autoMessageSendModes"
                :key="mode.value"
                :value="mode.value"
              >
                {{ mode.text }}
              </option>
            </select>
          </div>

          <div
            v-if="models.autoMessage.sendMode === 'cron'"
            class="mb-3"
          >
            <label
              class="form-label"
              for="formAutoMessageCron"
            >
              Send at
            </label>
            <input
              id="formAutoMessageCron"
              v-model="models.autoMessage.cron"
              class="form-control"
              :class="fieldStateClass(validateAutoMessageCron)"
              type="text"
            >
            <div class="form-text">
              <code>@every [time]</code> or cron syntax
            </div>
          </div>

          <div
            v-if="models.autoMessage.sendMode === 'lines'"
            class="mb-3"
          >
            <label
              class="form-label"
              for="formAutoMessageNLines"
            >
              Send every
            </label>
            <div class="input-group">
              <input
                id="formAutoMessageNLines"
                v-model="models.autoMessage.message_interval"
                class="form-control"
                type="number"
              >
              <span class="input-group-text">Lines</span>
            </div>
          </div>

          <hr>

          <div class="form-check form-switch mb-3">
            <input
              id="formOnlyOnLive"
              v-model="models.autoMessage.only_on_live"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="formOnlyOnLive"
            >
              Send only when channel is live
            </label>
          </div>

          <div class="form-check form-switch mb-3">
            <input
              id="formDisableAutoMessage"
              v-model="models.autoMessage.disable"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="formDisableAutoMessage"
            >
              Disable Auto-Message entirely
            </label>
          </div>

          <div class="mb-3">
            <label
              class="form-label"
              for="formAutoMessageDisableOnTemplate"
            >
              Disable on Template
            </label>
            <div class="form-text mb-2">
              <fa-icon
                fixed-width
                class="me-1 text-success"
                :icon="['fas', 'code']"
              />
              Template expression resulting in <code>true</code> to disable the rule or <code>false</code> to enable it.
            </div>
            <template-editor
              id="formAutoMessageDisableOnTemplate"
              v-model="models.autoMessage.disable_on_template"
              @valid-template="(valid: boolean) => updateTemplateValid('autoMessage.disable_on_template', valid)"
            />
          </div>
        </div>

        <div class="col-lg-4">
          <h6>Getting Help</h6>
          <p>
            For information about available template functions and variables to use in the <strong>Message</strong> see the
            <a
              href="https://github.com/Luzifer/twitch-bot/wiki#templating"
              rel="noopener noreferrer"
              target="_blank"
            >Templating</a> section of the Wiki.
          </p>
          <p>
            For information about the <strong>Cron</strong> syntax have a look at
            <a
              href="https://cron.help/"
              rel="noopener noreferrer"
              target="_blank"
            >cron.help</a>.
            You can also use <code>@every [time]</code> syntax.
          </p>
        </div>
      </div>
    </AppModal>
  </div>
</template>

<script lang="ts">
import * as constants from '../lib/const'
import { api } from '../api'
import AppModal from '../components/AppModal.vue'
import type { AutoMessage } from '../types'
import { confirmDialog } from '../lib/confirmModal'
import { defineComponent } from 'vue'
import TemplateEditor from '../components/tplEditor.vue'

type AutoMessageForm = AutoMessage & {
  sendMode?: 'cron' | 'lines'
}

export default defineComponent({
  components: { AppModal, TemplateEditor },

  computed: {
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

      if (!this.validateAutoMessageChannel) {
        return false
      }

      if (Object.values(this.templateValid).some(valid => !valid)) {
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

      return Boolean(this.models.autoMessage.cron?.match(constants.CRON_VALIDATION))
    },

    validateAutoMessageMessageLength() {
      return this.models.autoMessage.use_action ? 496 : 500
    },
  },

  data() {
    return {
      autoMessageSendModes: [
        { text: 'Cron', value: 'cron' },
        { text: 'Number of lines', value: 'lines' },
      ],

      autoMessages: [] as AutoMessage[],
      models: {
        autoMessage: {} as AutoMessageForm,
      },

      showAutoMessageEditModal: false,
      templateValid: {} as Record<string, boolean>,
    }
  },

  methods: {
    async deleteAutoMessage(uuid: string) {
      if (!await confirmDialog('Do you really want to delete this message?', {
        buttonSize: 'sm',
        cancelTitle: 'NO',
        centered: true,
        okTitle: 'YES',
        okVariant: 'danger',
        size: 'sm',
        title: 'Please Confirm',
      })) {
        return
      }

      try {
        await api.delete(`config-editor/auto-messages/${uuid}`)
        this.$bus.emit(constants.NOTIFY_CHANGE_PENDING, true)
      } catch (err) {
        this.$bus.emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    editAutoMessage(msg: AutoMessage) {
      this.models.autoMessage = {
        ...msg,
        sendMode: msg.cron ? 'cron' : 'lines',
      }
      this.templateValid = {}
      this.showAutoMessageEditModal = true
    },

    async fetchAutoMessages() {
      this.$bus.emit(constants.NOTIFY_LOADING_DATA, true)
      try {
        this.autoMessages = await api.get<AutoMessage[]>('config-editor/auto-messages') || []
        this.$bus.emit(constants.NOTIFY_CHANGE_PENDING, false)
      } catch (err) {
        this.$bus.emit(constants.NOTIFY_FETCH_ERROR, err)
      } finally {
        this.$bus.emit(constants.NOTIFY_LOADING_DATA, false)
      }
    },

    fieldStateClass(state: boolean) {
      return {
        'is-invalid': state === false,
        'is-valid': state === true,
      }
    },

    newAutoMessage() {
      this.models.autoMessage = {}
      this.templateValid = {}
      this.showAutoMessageEditModal = true
    },

    async saveAutoMessage(evt: { preventDefault: () => void }) {
      if (!this.validateAutoMessage) {
        evt.preventDefault()
        return
      }

      const obj = { ...this.models.autoMessage }

      if (obj.sendMode === 'cron') {
        delete obj.message_interval
      } else if (obj.sendMode === 'lines') {
        delete obj.cron
        obj.message_interval = Number(obj.message_interval)
      }

      try {
        if (obj.uuid) {
          await api.put(`config-editor/auto-messages/${obj.uuid}`, obj)
        } else {
          await api.post('config-editor/auto-messages', obj)
        }

        this.$bus.emit(constants.NOTIFY_CHANGE_PENDING, true)
      } catch (err) {
        evt.preventDefault()
        this.$bus.emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    updateTemplateValid(id: string, valid: boolean) {
      this.templateValid = {
        ...this.templateValid,
        [id]: valid,
      }
    },
  },

  mounted() {
    this.$bus.on(constants.NOTIFY_CONFIG_RELOAD, () => {
      this.fetchAutoMessages()
    })

    this.fetchAutoMessages()
  },

  name: 'TwitchBotAutoMessagesView',
})
</script>
