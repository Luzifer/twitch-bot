<template>
  <div>
    <b-row>
      <b-col>
        <b-table
          key="rulesTable"
          :busy="!rules"
          :fields="rulesFields"
          hover
          :items="rules"
          striped
        >
          <template #cell(_actions)="data">
            <b-button-group size="sm">
              <b-button @click="editRule(data.item)">
                <font-awesome-icon
                  fixed-width
                  class="mr-1"
                  :icon="['fas', 'pen']"
                />
              </b-button>
              <b-button
                variant="danger"
                @click="deleteRule(data.item.uuid)"
              >
                <font-awesome-icon
                  fixed-width
                  class="mr-1"
                  :icon="['fas', 'minus']"
                />
              </b-button>
            </b-button-group>
          </template>

          <template #cell(_match)="data">
            <b-badge
              v-for="badge in formatRuleMatch(data.item)"
              :key="badge.key"
              class="m-1 text-truncate text-left col-12"
              style="max-width: 250px;"
            >
              <strong>{{ badge.key }}</strong> <code class="ml-2">{{ badge.value }}</code>
            </b-badge>
          </template>

          <template #cell(_description)="data">
            <template v-if="data.item.description">
              {{ data.item.description }}<br>
            </template>
            <b-badge
              v-if="data.item.disable"
              class="mt-1 mr-1"
              variant="danger"
            >
              Disabled
            </b-badge>
            <b-badge
              v-for="badge in formatRuleActions(data.item)"
              :key="badge"
              class="mt-1 mr-1"
            >
              {{ badge }}
            </b-badge>
          </template>

          <template #head(_actions)="">
            <b-button-group size="sm">
              <b-button
                variant="success"
                @click="newRule"
              >
                <font-awesome-icon
                  fixed-width
                  class="mr-1"
                  :icon="['fas', 'plus']"
                />
              </b-button>
            </b-button-group>
          </template>
        </b-table>
      </b-col>
    </b-row>

    <!-- Rule Editor -->
    <b-modal
      v-if="showRuleEditModal"
      hide-header-close
      :ok-disabled="!validateRule()"
      ok-title="Save"
      scrollable
      size="xl"
      :visible="showRuleEditModal"
      title="Edit Rule"
      @hidden="showRuleEditModal=false"
      @ok="saveRule"
    >
      <b-row>
        <b-col cols="6">
          <b-form-group
            description="Human readable description for the rules list"
            label="Description"
            label-for="formRuleDescription"
          >
            <b-form-input
              id="formRuleDescription"
              v-model="models.rule.description"
              type="text"
            />
          </b-form-group>

          <hr>

          <b-tabs content-class="mt-3">
            <b-tab>
              <div slot="title">
                Matcher <b-badge>{{ countRuleMatchers }}</b-badge>
              </div>

              <b-form-group
                description="Channel with leading hash: #mychannel - matches all channels if none are given"
                label="Match Channels"
                label-for="formRuleMatchChannels"
              >
                <b-form-tags
                  id="formRuleMatchChannels"
                  v-model="models.rule.match_channels"
                  no-add-on-enter
                  placeholder="Enter channels separated by space or comma"
                  remove-on-delete
                  separator=" ,"
                  :tag-validator="(tag) => Boolean(tag.match(/^#[a-zA-Z0-9_]{4,25}$/))"
                />
              </b-form-group>

              <b-form-group
                description="Matches no events if not set"
                label="Match Event"
                label-for="formRuleMatchEvent"
              >
                <b-form-select
                  id="formRuleMatchEvent"
                  v-model="models.rule.match_event"
                  :options="availableEvents"
                />
              </b-form-group>

              <b-form-group
                description="Regular expression to match the message, matches all messages when not set"
                label="Match Message"
                label-for="formRuleMatchMessage"
              >
                <b-form-input
                  id="formRuleMatchMessage"
                  v-model="models.rule.match_message"
                  :state="models.rule.match_message__validation"
                  type="text"
                />
              </b-form-group>

              <b-form-group
                description="Matches all users if none are given"
                label="Match Users"
                label-for="formRuleMatchUsers"
              >
                <b-form-tags
                  id="formRuleMatchUsers"
                  v-model="models.rule.match_users"
                  no-add-on-enter
                  placeholder="Enter usernames separated by space or comma"
                  remove-on-delete
                  separator=" ,"
                  :tag-validator="(tag) => Boolean(tag.match(/^[a-z0-9_]{4,25}$/))"
                />
              </b-form-group>
            </b-tab>
            <b-tab>
              <div slot="title">
                Cooldown <b-badge>{{ countRuleCooldowns }}</b-badge>
              </div>

              <b-row>
                <b-col>
                  <b-form-group
                    label="Rule Cooldown"
                    label-for="formRuleRuleCooldown"
                  >
                    <b-form-input
                      id="formRuleRuleCooldown"
                      v-model="models.rule.cooldown"
                      placeholder="No Cooldown"
                      :state="validateDuration(models.rule.cooldown, false)"
                      type="text"
                    />
                  </b-form-group>
                </b-col>
                <b-col>
                  <b-form-group
                    label="Channel Cooldown"
                    label-for="formRuleChannelCooldown"
                  >
                    <b-form-input
                      id="formRuleChannelCooldown"
                      v-model="models.rule.channel_cooldown"
                      placeholder="No Cooldown"
                      :state="validateDuration(models.rule.channel_cooldown, false)"
                      type="text"
                    />
                  </b-form-group>
                </b-col>
                <b-col>
                  <b-form-group
                    label="User Cooldown"
                    label-for="formRuleUserCooldown"
                  >
                    <b-form-input
                      id="formRuleUserCooldown"
                      v-model="models.rule.user_cooldown"
                      placeholder="No Cooldown"
                      :state="validateDuration(models.rule.user_cooldown, false)"
                      type="text"
                    />
                  </b-form-group>
                </b-col>
              </b-row>

              <b-form-group
                :description="`Available badges: ${$root.vars.IRCBadges ? $root.vars.IRCBadges.join(', ') : ''}`"
                label="Skip Cooldown for"
                label-for="formRuleSkipCooldown"
              >
                <b-form-tags
                  id="formRuleSkipCooldown"
                  v-model="models.rule.skip_cooldown_for"
                  no-add-on-enter
                  placeholder="Enter badges separated by space or comma"
                  remove-on-delete
                  separator=" ,"
                  :tag-validator="validateTwitchBadge"
                />
              </b-form-group>
            </b-tab>
            <b-tab>
              <div slot="title">
                Conditions <b-badge>{{ countRuleConditions }}</b-badge>
              </div>

              <p>Disable rule&hellip;</p>
              <b-row>
                <b-col>
                  <b-form-group>
                    <b-form-checkbox
                      v-model="models.rule.disable"
                      switch
                    >
                      completely
                    </b-form-checkbox>
                  </b-form-group>
                </b-col>
                <b-col>
                  <b-form-group>
                    <b-form-checkbox
                      v-model="models.rule.disable_on_offline"
                      switch
                    >
                      when channel is offline
                    </b-form-checkbox>
                  </b-form-group>
                </b-col>
                <b-col>
                  <b-form-group>
                    <b-form-checkbox
                      v-model="models.rule.disable_on_permit"
                      switch
                    >
                      when user has permit
                    </b-form-checkbox>
                  </b-form-group>
                </b-col>
              </b-row>

              <b-form-group
                :description="`Available badges: ${$root.vars.IRCBadges ? $root.vars.IRCBadges.join(', ') : ''}`"
                label="Disable Rule for"
                label-for="formRuleDisableOn"
              >
                <b-form-tags
                  id="formRuleDisableOn"
                  v-model="models.rule.disable_on"
                  no-add-on-enter
                  placeholder="Enter badges separated by space or comma"
                  remove-on-delete
                  separator=" ,"
                  :tag-validator="validateTwitchBadge"
                />
              </b-form-group>

              <b-form-group
                :description="`Available badges: ${$root.vars.IRCBadges ? $root.vars.IRCBadges.join(', ') : ''}`"
                label="Enable Rule for"
                label-for="formRuleEnableOn"
              >
                <b-form-tags
                  id="formRuleEnableOn"
                  v-model="models.rule.enable_on"
                  no-add-on-enter
                  placeholder="Enter badges separated by space or comma"
                  remove-on-delete
                  separator=" ,"
                  :tag-validator="validateTwitchBadge"
                />
              </b-form-group>

              <b-form-group
                label="Disable on Template"
                label-for="formRuleDisableOnTemplate"
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
                  id="formRuleDisableOnTemplate"
                  v-model="models.rule.disable_on_template"
                />
              </b-form-group>
            </b-tab>
          </b-tabs>
        </b-col>

        <b-col cols="6">
          <div
            class="accordion"
            role="tablist"
          >
            <b-card
              v-for="(action, idx) in models.rule.actions"
              :key="`${models.rule.uuid}-action-${idx}`"
              no-body
              class="mb-1"
            >
              <b-card-header
                header-tag="header"
                class="p-1 d-flex"
                role="tab"
              >
                <b-button-group class="flex-fill">
                  <b-button
                    v-b-toggle="`${models.rule.uuid}-action-${idx}`"
                    block
                    variant="primary"
                  >
                    {{ getActionDefinitionByType(action.type).name }}
                    <font-awesome-icon
                      v-if="actionHasValidationError(idx)"
                      fixed-width
                      class="mr-1 text-danger"
                      :icon="['fas', 'exclamation-triangle']"
                    />
                  </b-button>
                  <b-button
                    :disabled="idx === 0"
                    variant="secondary"
                    @click="moveAction(idx, -1)"
                  >
                    <font-awesome-icon
                      fixed-width
                      class="mr-1"
                      :icon="['fas', 'chevron-up']"
                    />
                  </b-button>
                  <b-button
                    :disabled="idx === models.rule.actions.length - 1"
                    variant="secondary"
                    @click="moveAction(idx, +1)"
                  >
                    <font-awesome-icon
                      fixed-width
                      class="mr-1"
                      :icon="['fas', 'chevron-down']"
                    />
                  </b-button>
                  <b-button
                    variant="danger"
                    @click="removeAction(idx)"
                  >
                    <font-awesome-icon
                      fixed-width
                      class="mr-1"
                      :icon="['fas', 'trash']"
                    />
                  </b-button>
                </b-button-group>
              </b-card-header>
              <b-collapse
                :id="`${models.rule.uuid}-action-${idx}`"
                accordion="my-accordion"
                role="tabpanel"
              >
                <b-card-body v-if="getActionDefinitionByType(action.type).fields && getActionDefinitionByType(action.type).fields.length > 0">
                  <template
                    v-for="field in getActionDefinitionByType(action.type).fields"
                  >
                    <b-form-group
                      v-if="field.type === 'bool'"
                      :key="field.name"
                    >
                      <div slot="description">
                        {{ field.description }}
                      </div>

                      <b-form-checkbox
                        v-model="models.rule.actions[idx].attributes[field.key]"
                        switch
                      >
                        {{ field.name }}
                      </b-form-checkbox>
                    </b-form-group>

                    <b-form-group
                      v-else-if="field.type === 'stringslice'"
                      :key="field.name"
                      :label="field.name"
                      :label-for="`${models.rule.uuid}-action-${idx}-${field.key}`"
                    >
                      <div slot="description">
                        <font-awesome-icon
                          v-if="field.support_template"
                          fixed-width
                          class="mr-1 text-success"
                          :icon="['fas', 'code']"
                          title="Supports Templating"
                        />
                        {{ field.description }}
                      </div>

                      <b-form-tags
                        :id="`${models.rule.uuid}-action-${idx}-${field.key}`"
                        v-model="models.rule.actions[idx].attributes[field.key]"
                        :state="validateActionArgument(idx, field.key)"
                        placeholder="Enter elements and press enter to add the element"
                        remove-on-delete
                      />
                    </b-form-group>

                    <b-form-group
                      v-else-if="field.support_template"
                      :key="field.name"
                      :label="field.name"
                      :label-for="`${models.rule.uuid}-action-${idx}-${field.key}`"
                    >
                      <div slot="description">
                        <font-awesome-icon
                          fixed-width
                          class="mr-1 text-success"
                          :icon="['fas', 'code']"
                          title="Supports Templating"
                        />
                        {{ field.description }}
                      </div>

                      <template-editor
                        :id="`${models.rule.uuid}-action-${idx}-${field.key}`"
                        v-model="models.rule.actions[idx].attributes[field.key]"
                        :state="validateActionArgument(idx, field.key)"
                      />
                    </b-form-group>

                    <b-form-group
                      v-else
                      :key="field.name"
                      :label="field.name"
                      :label-for="`${models.rule.uuid}-action-${idx}-${field.key}`"
                    >
                      <div slot="description">
                        {{ field.description }}
                      </div>

                      <b-form-input
                        :id="`${models.rule.uuid}-action-${idx}-${field.key}`"
                        v-model="models.rule.actions[idx].attributes[field.key]"
                        :placeholder="field.default"
                        :required="!field.optional"
                        :state="validateActionArgument(idx, field.key)"
                        type="text"
                      />
                    </b-form-group>
                  </template>
                </b-card-body>
                <b-card-body v-else>
                  This action has no attributes.
                </b-card-body>
              </b-collapse>
            </b-card>

            <hr>

            <b-form-group
              :description="addActionDescription"
              label="Add Action"
              label-for="ruleAddAction"
            >
              <b-input-group>
                <b-form-select
                  id="ruleAddAction"
                  v-model="models.addAction"
                  :options="availableActionsForAdd"
                />
                <b-input-group-append>
                  <b-button
                    :disabled="!models.addAction"
                    variant="success"
                    @click="addAction"
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
            </b-form-group>
            <div />
          </div>
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
    addActionDescription() {
      if (!this.models.addAction) {
        return ''
      }

      for (const action of this.actions) {
        if (action.type === this.models.addAction) {
          return action.description
        }
      }

      throw new Error('found no description for action')
    },

    availableActionsForAdd() {
      return this.actions.map(a => ({ text: a.name, value: a.type }))
    },

    availableEvents() {
      return [
        { text: 'Clear Event-Matching', value: null },
        ...this.$root.vars.KnownEvents,
      ]
    },

    countRuleConditions() {
      let count = 0
      count += this.models.rule.disable ? 1 : 0
      count += this.models.rule.disable_on_offline ? 1 : 0
      count += this.models.rule.disable_on_permit ? 1 : 0
      count += this.models.rule.disable_on ? 1 : 0
      count += this.models.rule.enable_on ? 1 : 0
      count += this.models.rule.disable_on_template ? 1 : 0
      return count
    },

    countRuleCooldowns() {
      let count = 0
      count += this.models.rule.cooldown ? 1 : 0
      count += this.models.rule.channel_cooldown ? 1 : 0
      count += this.models.rule.user_cooldown ? 1 : 0
      count += this.models.rule.skip_cooldown_for ? 1 : 0
      return count
    },

    countRuleMatchers() {
      let count = 0
      count += this.models.rule.match_channels ? 1 : 0
      count += this.models.rule.match_event ? 1 : 0
      count += this.models.rule.match_message ? 1 : 0
      count += this.models.rule.match_users ? 1 : 0
      return count
    },
  },

  data() {
    return {
      actions: [],
      models: {
        addAction: '',
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

      showRuleEditModal: false,
      validateReason: null,
    }
  },

  methods: {
    actionHasValidationError(idx) {
      const action = this.models.rule.actions[idx]
      const def = this.getActionDefinitionByType(action.type)

      for (const field of def.fields || []) {
        if (!this.validateActionArgument(idx, field.key)) {
          return true
        }
      }

      return false
    },

    addAction() {
      if (!this.models.rule.actions) {
        Vue.set(this.models.rule, 'actions', [])
      }

      this.models.rule.actions.push({ attributes: {}, type: this.models.addAction })
    },

    deleteRule(uuid) {
      this.$bvModal.msgBoxConfirm('Do you really want to delete this rule?', {
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

          return axios.delete(`config-editor/rules/${uuid}`, this.$root.axiosOptions)
            .then(() => {
              this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
            })
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    editRule(msg) {
      Vue.set(this.models, 'rule', {
        ...msg,
        actions: msg.actions?.map(action => ({ ...action, attributes: action.attributes || {} })) || [],
        channel_cooldown: this.fixDurationRepresentationToString(msg.channel_cooldown),
        cooldown: this.fixDurationRepresentationToString(msg.cooldown),
        user_cooldown: this.fixDurationRepresentationToString(msg.user_cooldown),
      })
      this.showRuleEditModal = true
      this.validateMatcherRegex()
    },

    fetchActions() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      return axios.get('config-editor/actions')
        .then(resp => {
          this.actions = resp.data
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    fetchRules() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      return axios.get('config-editor/rules', this.$root.axiosOptions)
        .then(resp => {
          this.rules = resp.data
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, false)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    fixDurationRepresentationToInt64(value) {
      let match = null

      switch (typeof value) {
      case 'string':
        match = value.match(/(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?/)
        return ((Number(match[1]) || 0) * 3600 + (Number(match[2]) || 0) * 60 + (Number(match[3]) || 0)) * constants.NANO

      default:
        return value
      }
    },

    fixDurationRepresentationToString(value) {
      let repr = ''

      switch (typeof value) {
      case 'number':
        value /= constants.NANO

        if (value >= 3600) {
          const h = Math.floor(value / 3600)
          repr += `${h}h`
          value -= h * 3600
        }

        if (value >= 60) {
          const m = Math.floor(value / 60)
          repr += `${m}m`
          value -= m * 60
        }

        if (value > 0) {
          repr += `${value}s`
        }

        return repr

      default:
        return value
      }
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

    moveAction(idx, direction) {
      const tmp = [...this.models.rule.actions]

      const eltmp = tmp[idx]
      tmp[idx] = tmp[idx + direction]
      tmp[idx + direction] = eltmp

      Vue.set(this.models.rule, 'actions', tmp)
    },

    newRule() {
      Vue.set(this.models, 'rule', { match_message__validation: true })
      this.showRuleEditModal = true
    },

    removeAction(idx) {
      this.models.rule.actions = this.models.rule.actions.filter((_, i) => i !== idx)
    },

    saveRule(evt) {
      if (!this.validateRule()) {
        evt.preventDefault()
        return
      }

      const obj = {
        ...this.models.rule,
        actions: this.models.rule.actions.map(action => ({
          ...action,
          attributes: Object.fromEntries(Object.entries(action.attributes)
            .filter(att => {
              const def = this.getActionDefinitionByType(action.type)
              const field = def.fields.filter(field => field.key === att[0])[0]

              if (!field) {
                // The field is not defined, drop it
                return false
              }

              if (att[1] === null || att[1] === undefined) {
                // Drop null / undefined values
                return false
              }

              // Check for zero-values and drop the field on zero-value
              switch (field.type) {
              case 'bool':
                if (att[1] === false) {
                  return false
                }
                break

              case 'duration':
                if (att[1] === '0s' || att[1] === '') {
                  return false
                }
                break

              case 'int64':
                if (att[1] === 0 || att[1] === '') {
                  return false
                }
                break

              case 'string':
                if (att[1] === '') {
                  return false
                }
                break

              case 'stringslice':
                if (att[1] === null || att[1].length === 0) {
                  return false
                }
                break
              }

              return true
            })),
        })),
      }

      if (obj.cooldown) {
        obj.cooldown = this.fixDurationRepresentationToInt64(obj.cooldown)
      } else {
        delete obj.cooldown
      }

      if (obj.channel_cooldown) {
        obj.channel_cooldown = this.fixDurationRepresentationToInt64(obj.channel_cooldown)
      } else {
        delete obj.channel_cooldown
      }

      if (obj.user_cooldown) {
        obj.user_cooldown = this.fixDurationRepresentationToInt64(obj.user_cooldown)
      } else {
        delete obj.user_cooldown
      }

      let promise = null
      if (obj.uuid) {
        promise = axios.put(`config-editor/rules/${obj.uuid}`, obj, this.$root.axiosOptions)
      } else {
        promise = axios.post(`config-editor/rules`, obj, this.$root.axiosOptions)
      }

      promise.then(() => {
        this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
      })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    validateActionArgument(idx, key) {
      const action = this.models.rule.actions[idx]
      const def = this.getActionDefinitionByType(action.type)

      if (!def || !def.fields) {
        return false
      }

      for (const field of def.fields) {
        if (field.key !== key) {
          continue
        }

        switch (field.type) {
        case 'bool':
          if (!field.optional && !action.attributes[field.key]) {
            return false
          }
          break

        case 'duration':
          if (!this.validateDuration(action.attributes[field.key], !field.optional)) {
            return false
          }
          break

        case 'int64':
          if (!field.optional && !action.attributes[field.key]) {
            return false
          }

          if (action.attributes[field.key] && isNaN(action.attributes[field.key])) {
            return false
          }

          break

        case 'string':
          if (!field.optional && !action.attributes[field.key]) {
            return false
          }
          break

        case 'stringslice':
          if (!field.optional && !action.attributes[field.key]) {
            return false
          }
          break
        }
        break
      }

      return true
    },

    validateDuration(duration, required) {
      if (!duration && !required) {
        return true
      }

      if (!duration && required) {
        return false
      }

      return Boolean(duration.match(/^(?:\d+(?:s|m|h))+$/))
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

          if (field.type === 'duration' && !this.validateDuration(action.attributes[field.key], field.optional)) {
            this.validateReason = `${action.type} -> ${field.key} -> duration`
            return false
          }
        }
      }

      return true
    },

    validateTwitchBadge(tag) {
      return this.$root.vars.IRCBadges.includes(tag)
    },
  },

  mounted() {
    this.$bus.$on(constants.NOTIFY_CONFIG_RELOAD, () => {
      this.fetchRules()
        .then(() => this.$bus.$emit(constants.NOTIFY_LOADING_DATA, false))
    })

    Promise.all([
      this.fetchRules(),
      this.fetchActions(),
    ]).then(() => this.$bus.$emit(constants.NOTIFY_LOADING_DATA, false))
  },

  name: 'TwitchBotEditorAppRules',

  watch: {
    'models.rule.match_message'(to, from) {
      if (to === from) {
        return
      }

      this.validateMatcherRegex()
    },
  },
}
</script>
