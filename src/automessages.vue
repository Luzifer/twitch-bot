<template>
  <div>
    <b-row>
      <b-col>
        <b-table
          key="autoMessagesTable"
          :busy="!autoMessages"
          :fields="autoMessageFields"
          hover
          :items="autoMessages"
          striped
        >
          <template #cell(actions)="data">
            <b-button-group size="sm">
              <b-button @click="editAutoMessage(data.item)">
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'pen']"
                />
              </b-button>
              <b-button
                variant="danger"
                @click="deleteAutoMessage(data.item.uuid)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'minus']"
                />
              </b-button>
            </b-button-group>
          </template>

          <template #cell(channel)="data">
            <font-awesome-icon
              fixed-width
              class="mr-1"
              :icon="['fas', 'hashtag']"
            />
            {{ data.value }}
          </template>

          <template #cell(cron)="data">
            <code>{{ data.value }}</code>
          </template>

          <template #cell(message)="data">
            {{ data.value }}<br>
            <b-badge
              v-if="data.item.disable"
              class="mt-1 mr-1"
              variant="danger"
            >
              Disabled
            </b-badge>
            <b-badge
              v-if="data.item.disable_on_template"
              class="mt-1 mr-1"
            >
              Disable on Template
            </b-badge>
            <b-badge
              v-if="data.item.only_on_live"
              class="mt-1 mr-1"
            >
              Only during Stream
            </b-badge>
          </template>

          <template #head(actions)="">
            <b-button-group size="sm">
              <b-button
                variant="success"
                @click="newAutoMessage"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'plus']"
                />
              </b-button>
            </b-button-group>
          </template>
        </b-table>
      </b-col>
    </b-row>

    <!-- Auto-Message Editor -->
    <b-modal
      v-if="showAutoMessageEditModal"
      hide-header-close
      :ok-disabled="!validateAutoMessage"
      ok-title="Save"
      size="lg"
      :visible="showAutoMessageEditModal"
      title="Edit Auto-Message"
      @hidden="showAutoMessageEditModal=false"
      @ok="saveAutoMessage"
    >
      <b-row>
        <b-col cols="8">
          <b-form-group
            label="Channel"
            label-for="formAutoMessageChannel"
          >
            <b-input-group
              prepend="#"
            >
              <b-form-input
                id="formAutoMessageChannel"
                v-model="models.autoMessage.channel"
                :state="validateAutoMessageChannel"
                type="text"
                required
              />
            </b-input-group>
          </b-form-group>

          <hr>

          <b-form-group
            label="Message"
            label-for="formAutoMessageMessage"
          >
            <template-editor
              id="formAutoMessageMessage"
              v-model="models.autoMessage.message"
              :state="models.autoMessage.message ? models.autoMessage.message.length <= validateAutoMessageMessageLength : false"
            />
            <div slot="description">
              <font-awesome-icon
                fixed-width
                class="mr-1 text-success"
                :icon="['fas', 'code']"
                title="Supports Templating"
              />
              {{ models.autoMessage.message && models.autoMessage.message.length || 0 }} / {{ validateAutoMessageMessageLength }}
            </div>
          </b-form-group>

          <b-form-group>
            <b-form-checkbox
              v-model="models.autoMessage.use_action"
              switch
            >
              Send message as action (<code>/me</code>)
            </b-form-checkbox>
          </b-form-group>

          <hr>

          <b-form-group
            label="Sending Mode"
            label-for="formAutoMessageSendMode"
          >
            <b-form-select
              id="formAutoMessageSendMode"
              v-model="models.autoMessage.sendMode"
              :options="autoMessageSendModes"
            />
          </b-form-group>

          <b-form-group
            v-if="models.autoMessage.sendMode === 'cron'"
            label="Send at"
            label-for="formAutoMessageCron"
          >
            <b-form-input
              id="formAutoMessageCron"
              v-model="models.autoMessage.cron"
              :state="validateAutoMessageCron"
              type="text"
            />

            <div slot="description">
              <code>@every [time]</code> or Cron syntax
            </div>
          </b-form-group>

          <b-form-group
            v-if="models.autoMessage.sendMode === 'lines'"
            label="Send every"
            label-for="formAutoMessageNLines"
          >
            <b-input-group
              append="Lines"
            >
              <b-form-input
                id="formAutoMessageNLines"
                v-model="models.autoMessage.message_interval"
                type="number"
              />
            </b-input-group>
          </b-form-group>

          <hr>

          <b-form-group>
            <b-form-checkbox
              v-model="models.autoMessage.only_on_live"
              switch
            >
              Send only when channel is live
            </b-form-checkbox>
          </b-form-group>

          <b-form-group>
            <b-form-checkbox
              v-model="models.autoMessage.disable"
              switch
            >
              Disable Auto-Message entirely
            </b-form-checkbox>
          </b-form-group>

          <b-form-group
            label="Disable on Template"
            label-for="formAutoMessageDisableOnTemplate"
          >
            <div slot="description">
              <font-awesome-icon
                fixed-width
                class="mr-1 text-success"
                :icon="['fas', 'code']"
                title="Supports Templating"
              />
              Template expression resulting in <code>true</code> to disable the rule or <code>false</code> to enable it
            </div>
            <template-editor
              id="formAutoMessageDisableOnTemplate"
              v-model="models.autoMessage.disable_on_template"
            />
          </b-form-group>
        </b-col>

        <b-col cols="4">
          <h6>Getting Help</h6>
          <p>
            For information about available template functions and variables to use in the <strong>Message</strong> see the <a
              href="https://github.com/Luzifer/twitch-bot/wiki#templating"
              rel="noopener noreferrer"
              target="_blank"
            >Templating</a> section of the Wiki.
          </p>
          <p>
            For information about the <strong>Cron</strong> syntax have a look at the <a
              href="https://cron.help/"
              rel="noopener noreferrer"
              target="_blank"
            >cron.help</a> site. Aditionally you can use <code>@every [time]</code> syntax. The <code>[time]</code> part is in format <code>1h30m20s</code>. You can leave out every segment but need to specify the unit of every segment. So for example <code>@every 1h</code> or <code>@every 10m</code> would be a valid specification.
          </p>
        </b-col>
      </b-row>
    </b-modal>
  </div>
</template>

<script>
import * as constants from './const.js'

import axios from 'axios'
import TemplateEditor from './tplEditor.vue'
import Vue from 'vue'

export default {
  components: { TemplateEditor },
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

      return Boolean(this.models.autoMessage.cron?.match(constants.CRON_VALIDATION))
    },

    validateAutoMessageMessageLength() {
      return this.models.autoMessage.use_action ? 496 : 500
    },
  },

  data() {
    return {
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

      models: {
        autoMessage: {},
      },

      showAutoMessageEditModal: false,
    }
  },

  methods: {
    deleteAutoMessage(uuid) {
      this.$bvModal.msgBoxConfirm('Do you really want to delete this message?', {
        buttonSize: 'sm',
        cancelTitle: 'NO',
        centered: true,
        okTitle: 'YES',
        okVariant: 'danger',
        size: 'sm',
        title: 'Please Confirm',
      })
        .then(val => {
          if (!val) {
            return
          }

          return axios.delete(`config-editor/auto-messages/${uuid}`, this.$root.axiosOptions)
            .then(() => {
              this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
            })
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    editAutoMessage(msg) {
      Vue.set(this.models, 'autoMessage', {
        ...msg,
        sendMode: msg.cron ? 'cron' : 'lines',
      })
      this.showAutoMessageEditModal = true
    },

    fetchAutoMessages() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      return axios.get('config-editor/auto-messages', this.$root.axiosOptions)
        .then(resp => {
          this.autoMessages = resp.data
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, false)
          this.$bus.$emit(constants.NOTIFY_LOADING_DATA, false)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    newAutoMessage() {
      Vue.set(this.models, 'autoMessage', {})
      this.showAutoMessageEditModal = true
    },

    saveAutoMessage(evt) {
      if (!this.validateAutoMessage) {
        evt.preventDefault()
        return
      }

      const obj = { ...this.models.autoMessage }

      if (this.models.autoMessage.sendMode === 'cron') {
        delete obj.message_interval
      } else if (this.models.autoMessage.sendMode === 'lines') {
        delete obj.cron
      }

      let promise = null
      if (obj.uuid) {
        promise = axios.put(`config-editor/auto-messages/${obj.uuid}`, obj, this.$root.axiosOptions)
      } else {
        promise = axios.post(`config-editor/auto-messages`, obj, this.$root.axiosOptions)
      }

      promise.then(() => {
        this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
      })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },
  },

  mounted() {
    this.$bus.$on(constants.NOTIFY_CONFIG_RELOAD, () => {
      this.fetchAutoMessages()
    })

    this.fetchAutoMessages()
  },

  name: 'TwitchBotEditorAppAutomessages',
}
</script>
