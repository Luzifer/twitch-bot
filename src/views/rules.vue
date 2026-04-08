<template>
  <div>
    <div class="row mb-2">
      <div class="col">
        <div class="input-group">
          <input
            v-model.lazy="filter"
            class="form-control"
            placeholder="Search in / filter rules..."
            type="text"
          >
          <button
            type="button"
            class="btn btn-secondary"
            @click="filter = ''"
          >
            <font-awesome-icon
              fixed-width
              :icon="['fas', 'trash']"
            />
          </button>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col">
        <div class="table-responsive">
          <table class="table table-hover table-striped align-middle">
            <thead>
              <tr>
                <th class="col-3 align-middle">
                  Match
                </th>
                <th class="col-8 align-middle">
                  Description
                </th>
                <th class="col-1 text-end align-middle">
                  <div class="btn-group btn-group-sm">
                    <button
                      type="button"
                      class="btn btn-success"
                      @click="newRule"
                    >
                      <font-awesome-icon
                        fixed-width
                        :icon="['fas', 'plus']"
                      />
                    </button>
                    <button
                      type="button"
                      class="btn btn-secondary"
                      @click="showRuleSubscribeModal = true"
                    >
                      <font-awesome-icon
                        fixed-width
                        :icon="['fas', 'download']"
                      />
                    </button>
                  </div>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!rules">
                <td
                  colspan="3"
                  class="text-center text-muted"
                >
                  Loading...
                </td>
              </tr>
              <tr v-else-if="filteredRules.length === 0">
                <td
                  colspan="3"
                  class="text-center text-muted"
                >
                  No entries found.
                </td>
              </tr>
              <template v-else>
                <tr
                  v-for="rule in filteredRules"
                  :key="rule.uuid"
                >
                  <td class="col-3 align-middle">
                    <span
                      v-for="badge in formatRuleMatch(rule)"
                      :key="badge.key"
                      class="badge bg-secondary-subtle text-secondary-emphasis m-1 text-truncate text-start col-12"
                      style="max-width: 250px;"
                    >
                      <strong>{{ badge.key }}</strong> <code class="ms-2">{{ badge.value }}</code>
                    </span>
                  </td>
                  <td class="col-8 align-middle">
                    <template v-if="rule.description">
                      {{ rule.description }}<br>
                    </template>
                    <span
                      v-if="rule.subscribe_from"
                      class="badge text-bg-primary mt-1 me-1"
                    >
                      Shared
                    </span>
                    <span
                      v-if="rule.disable"
                      class="badge text-bg-danger mt-1 me-1"
                    >
                      Disabled
                    </span>
                    <span
                      v-for="(badge, idx) in formatRuleActions(rule)"
                      :key="`${badge}-${idx}`"
                      class="badge bg-secondary-subtle text-secondary-emphasis mt-1 me-1"
                    >
                      {{ badge }}
                    </span>
                  </td>
                  <td class="col-1 text-end align-middle">
                    <div class="btn-group btn-group-sm">
                      <button
                        v-if="rule.subscribe_from"
                        type="button"
                        class="btn btn-secondary"
                        disabled
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'download']"
                        />
                      </button>
                      <button
                        v-else
                        type="button"
                        class="btn btn-secondary"
                        @click="editRule(rule)"
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'pen']"
                        />
                      </button>
                      <button
                        type="button"
                        class="btn btn-danger"
                        @click="rule.uuid && deleteRule(rule.uuid)"
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'minus']"
                        />
                      </button>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <AppModal
      v-if="showRuleSubscribeModal"
      :model-value="showRuleSubscribeModal"
      :ok-disabled="!models.subscriptionURL"
      ok-title="Subscribe"
      size="md"
      title="Subscribe Rule"
      @hidden="showRuleSubscribeModal = false"
      @update:model-value="showRuleSubscribeModal = $event"
      @ok="subscribeRule"
    >
      <div class="mb-3">
        <label
          class="form-label"
          for="formRuleSubURL"
        >Rule Subscription URL</label>
        <input
          id="formRuleSubURL"
          v-model="models.subscriptionURL"
          class="form-control"
          :class="{
            'is-invalid': !models.subscriptionURL,
            'is-valid': Boolean(models.subscriptionURL),
          }"
          type="text"
        >
      </div>
    </AppModal>

    <AppModal
      v-if="showRuleEditModal"
      :model-value="showRuleEditModal"
      :ok-disabled="!validateRule()"
      ok-title="Save"
      scrollable
      size="xl"
      title="Edit Rule"
      @hidden="showRuleEditModal = false"
      @update:model-value="showRuleEditModal = $event"
      @ok="saveRule"
    >
      <div class="row">
        <div class="col-6">
          <div class="mb-3">
            <label
              class="form-label"
              for="formRuleDescription"
            >Description</label>
            <input
              id="formRuleDescription"
              v-model="models.rule.description"
              class="form-control"
              type="text"
            >
            <div class="form-text">
              Human readable description for the rules list
            </div>
          </div>

          <hr>

          <ul class="nav nav-tabs">
            <li class="nav-item">
              <button
                type="button"
                class="nav-link"
                :class="{ active: activeRuleTab === 'matcher' }"
                @click="activeRuleTab = 'matcher'"
              >
                Matcher <span class="badge bg-secondary-subtle text-secondary-emphasis">{{ countRuleMatchers }}</span>
              </button>
            </li>
            <li class="nav-item">
              <button
                type="button"
                class="nav-link"
                :class="{ active: activeRuleTab === 'cooldown' }"
                @click="activeRuleTab = 'cooldown'"
              >
                Cooldown <span class="badge bg-secondary-subtle text-secondary-emphasis">{{ countRuleCooldowns }}</span>
              </button>
            </li>
            <li class="nav-item">
              <button
                type="button"
                class="nav-link"
                :class="{ active: activeRuleTab === 'conditions' }"
                @click="activeRuleTab = 'conditions'"
              >
                Conditions <span class="badge bg-secondary-subtle text-secondary-emphasis">{{ countRuleConditions }}</span>
              </button>
            </li>
            <li class="nav-item">
              <button
                type="button"
                class="nav-link"
                :class="{ active: activeRuleTab === 'exceptions' }"
                @click="activeRuleTab = 'exceptions'"
              >
                Exceptions <span class="badge bg-secondary-subtle text-secondary-emphasis">{{ countRuleExceptions }}</span>
              </button>
            </li>
          </ul>

          <div class="tab-content mt-3">
            <div
              v-if="activeRuleTab === 'matcher'"
              class="tab-pane active"
            >
              <div class="mb-3">
                <label
                  class="form-label"
                  for="formRuleMatchChannels"
                >Match Channels</label>
                <TagInput
                  id="formRuleMatchChannels"
                  v-model="models.rule.match_channels"
                  placeholder="Enter channels separated by space or comma"
                  :validator="(tag: string) => Boolean(tag.match(/^#[a-zA-Z0-9_]{4,25}$/))"
                />
                <div class="form-text">
                  Channel with leading hash: #mychannel - matches all channels if none are given
                </div>
              </div>

              <div class="mb-3">
                <label
                  class="form-label"
                  for="formRuleMatchEvent"
                >Match Event</label>
                <select
                  id="formRuleMatchEvent"
                  v-model="models.rule.match_event"
                  class="form-select"
                >
                  <option
                    v-for="option in availableEvents"
                    :key="String(option.value)"
                    :value="option.value"
                  >
                    {{ option.text }}
                  </option>
                </select>
                <div class="form-text">
                  Matches no events if not set
                </div>
              </div>

              <div class="mb-3">
                <label
                  class="form-label"
                  for="formRuleMatchMessage"
                >Match Message</label>
                <input
                  id="formRuleMatchMessage"
                  v-model="models.rule.match_message"
                  class="form-control"
                  :class="{
                    'is-invalid': models.rule.match_message__validation === false,
                    'is-valid': models.rule.match_message__validation === true,
                  }"
                  type="text"
                >
                <div class="form-text">
                  Regular expression to match the message, matches all messages when not set
                </div>
              </div>

              <div class="mb-3">
                <label
                  class="form-label"
                  for="formRuleMatchUsers"
                >Match Users</label>
                <TagInput
                  id="formRuleMatchUsers"
                  v-model="models.rule.match_users"
                  placeholder="Enter usernames separated by space or comma"
                  :validator="(tag: string) => Boolean(tag.match(/^[a-z0-9_]{4,25}$/))"
                />
                <div class="form-text">
                  Matches all users if none are given
                </div>
              </div>
            </div>

            <div
              v-if="activeRuleTab === 'cooldown'"
              class="tab-pane active"
            >
              <div class="row">
                <div class="col">
                  <div class="mb-3">
                    <label
                      class="form-label"
                      for="formRuleRuleCooldown"
                    >Rule Cooldown</label>
                    <input
                      id="formRuleRuleCooldown"
                      v-model="models.rule.cooldown"
                      class="form-control"
                      :class="{
                        'is-invalid': validateDuration(models.rule.cooldown, false) === false,
                        'is-valid': validateDuration(models.rule.cooldown, false) === true,
                      }"
                      placeholder="No Cooldown"
                      type="text"
                    >
                  </div>
                </div>
                <div class="col">
                  <div class="mb-3">
                    <label
                      class="form-label"
                      for="formRuleChannelCooldown"
                    >Channel Cooldown</label>
                    <input
                      id="formRuleChannelCooldown"
                      v-model="models.rule.channel_cooldown"
                      class="form-control"
                      :class="{
                        'is-invalid': validateDuration(models.rule.channel_cooldown, false) === false,
                        'is-valid': validateDuration(models.rule.channel_cooldown, false) === true,
                      }"
                      placeholder="No Cooldown"
                      type="text"
                    >
                  </div>
                </div>
                <div class="col">
                  <div class="mb-3">
                    <label
                      class="form-label"
                      for="formRuleUserCooldown"
                    >User Cooldown</label>
                    <input
                      id="formRuleUserCooldown"
                      v-model="models.rule.user_cooldown"
                      class="form-control"
                      :class="{
                        'is-invalid': validateDuration(models.rule.user_cooldown, false) === false,
                        'is-valid': validateDuration(models.rule.user_cooldown, false) === true,
                      }"
                      placeholder="No Cooldown"
                      type="text"
                    >
                  </div>
                </div>
              </div>

              <div class="mb-3">
                <label
                  class="form-label"
                  for="formRuleSkipCooldown"
                >Skip Cooldown for</label>
                <TagInput
                  id="formRuleSkipCooldown"
                  v-model="models.rule.skip_cooldown_for"
                  placeholder="Enter badges separated by space or comma"
                  :validator="validateTwitchBadge"
                />
                <div class="form-text">
                  Available badges: {{ appStore.vars.IRCBadges ? appStore.vars.IRCBadges.join(', ') : '' }}
                </div>
              </div>
            </div>

            <div
              v-if="activeRuleTab === 'conditions'"
              class="tab-pane active"
            >
              <p>Disable rule&hellip;</p>
              <div class="row">
                <div class="col">
                  <div class="mb-3">
                    <div class="form-check form-switch">
                      <input
                        id="rule-disable"
                        v-model="models.rule.disable"
                        class="form-check-input"
                        type="checkbox"
                      >
                      <label
                        class="form-check-label"
                        for="rule-disable"
                      >completely</label>
                    </div>
                  </div>
                </div>
                <div class="col">
                  <div class="mb-3">
                    <div class="form-check form-switch">
                      <input
                        id="rule-disable-offline"
                        v-model="models.rule.disable_on_offline"
                        class="form-check-input"
                        type="checkbox"
                      >
                      <label
                        class="form-check-label"
                        for="rule-disable-offline"
                      >when channel is offline</label>
                    </div>
                  </div>
                </div>
                <div class="col">
                  <div class="mb-3">
                    <div class="form-check form-switch">
                      <input
                        id="rule-disable-permit"
                        v-model="models.rule.disable_on_permit"
                        class="form-check-input"
                        type="checkbox"
                      >
                      <label
                        class="form-check-label"
                        for="rule-disable-permit"
                      >when user has permit</label>
                    </div>
                  </div>
                </div>
              </div>

              <div class="mb-3">
                <label
                  class="form-label"
                  for="formRuleDisableOn"
                >Disable Rule for</label>
                <TagInput
                  id="formRuleDisableOn"
                  v-model="models.rule.disable_on"
                  placeholder="Enter badges separated by space or comma"
                  :validator="validateTwitchBadge"
                />
                <div class="form-text">
                  Available badges: {{ appStore.vars.IRCBadges ? appStore.vars.IRCBadges.join(', ') : '' }}
                </div>
              </div>

              <div class="mb-3">
                <label
                  class="form-label"
                  for="formRuleEnableOn"
                >Enable Rule for</label>
                <TagInput
                  id="formRuleEnableOn"
                  v-model="models.rule.enable_on"
                  placeholder="Enter badges separated by space or comma"
                  :validator="validateTwitchBadge"
                />
                <div class="form-text">
                  Available badges: {{ appStore.vars.IRCBadges ? appStore.vars.IRCBadges.join(', ') : '' }}
                </div>
              </div>

              <div class="mb-3">
                <label
                  class="form-label"
                  for="formRuleDisableOnTemplate"
                >Disable on Template</label>
                <template-editor
                  id="formRuleDisableOnTemplate"
                  v-model="models.rule.disable_on_template"
                  @valid-template="(valid: boolean) => updateTemplateValid('rule.disable_on_template', valid)"
                />
                <div class="form-text">
                  <font-awesome-icon
                    fixed-width
                    class="me-1 text-success"
                    :icon="['fas', 'code']"
                    title="Supports Templating"
                  />
                  Template expression resulting in <code>true</code> to disable the rule or <code>false</code> to enable it
                </div>
              </div>
            </div>

            <div
              v-if="activeRuleTab === 'exceptions'"
              class="tab-pane active"
            >
              <div class="alert alert-info">
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'info-circle']"
                />
                If the message does match one of these regular expressions, the rule is not executed.
                (Use for example to disable link protection on certain messages.)
              </div>

              <div class="list-group list-group-flush">
                <div class="list-group-item">
                  <div class="input-group">
                    <input
                      v-model="models.addException"
                      class="form-control"
                      :class="{
                        'is-invalid': models.addException__validation === false,
                        'is-valid': models.addException__validation === true,
                      }"
                      type="text"
                      @keyup="validateExceptionRegex"
                      @paste="validateExceptionRegex"
                      @keyup.enter="addException"
                    >
                    <button
                      type="button"
                      class="btn btn-success"
                      :disabled="!models.addException__validation"
                      @click="addException"
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

                <div
                  v-for="ex in models.rule.disable_on_match_messages"
                  :key="ex"
                  class="list-group-item d-flex align-items-center align-middle"
                >
                  <code class="me-auto">{{ ex }}</code>
                  <button
                    type="button"
                    class="btn btn-danger btn-sm"
                    @click="removeException(ex)"
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
        </div>

        <div class="col-6">
          <div
            class="accordion"
            role="tablist"
          >
            <div
              v-for="(action, idx) in models.rule.actions"
              :key="`${models.rule.uuid}-action-${idx}`"
              class="card mb-1"
            >
              <div
                class="card-header p-1 d-flex"
                role="tab"
              >
                <div class="btn-group flex-fill">
                  <button
                    type="button"
                    class="btn btn-primary flex-grow-1 text-start"
                    @click="openActionIndex = openActionIndex === idx ? null : Number(idx)"
                  >
                    {{ getActionDefinitionByType(action.type).name }}
                    <font-awesome-icon
                      v-if="actionHasValidationError(Number(idx))"
                      fixed-width
                      class="me-1 text-danger"
                      :icon="['fas', 'exclamation-triangle']"
                    />
                  </button>
                  <button
                    type="button"
                    class="btn btn-secondary rule-action-icon-btn"
                    :disabled="idx === 0"
                    @click="moveAction(idx, -1)"
                  >
                    <font-awesome-icon
                      fixed-width
                      class="me-1"
                      :icon="['fas', 'chevron-up']"
                    />
                  </button>
                  <button
                    type="button"
                    class="btn btn-secondary rule-action-icon-btn"
                    :disabled="idx === models.rule.actions.length - 1"
                    @click="moveAction(idx, +1)"
                  >
                    <font-awesome-icon
                      fixed-width
                      class="me-1"
                      :icon="['fas', 'chevron-down']"
                    />
                  </button>
                  <button
                    type="button"
                    class="btn btn-danger rule-action-icon-btn"
                    @click="removeAction(idx)"
                  >
                    <font-awesome-icon
                      fixed-width
                      class="me-1"
                      :icon="['fas', 'trash']"
                    />
                  </button>
                </div>
              </div>
              <div
                class="collapse"
                :class="{ show: openActionIndex === idx }"
              >
                <div
                  v-if="getActionDefinitionByType(action.type).fields && getActionDefinitionByType(action.type).fields.length > 0"
                  class="card-body"
                >
                  <template v-for="field in getActionDefinitionByType(action.type).fields">
                    <div
                      v-if="field.type === 'bool'"
                      :key="`${field.name}-bool`"
                      class="mb-3"
                    >
                      <div class="form-text mb-2">
                        {{ field.description }}
                      </div>
                      <div class="form-check form-switch">
                        <input
                          :id="`${models.rule.uuid}-action-${idx}-${field.key}`"
                          v-model="models.rule.actions[idx].attributes[field.key]"
                          class="form-check-input"
                          type="checkbox"
                        >
                        <label
                          class="form-check-label"
                          :for="`${models.rule.uuid}-action-${idx}-${field.key}`"
                        >{{ field.name }}</label>
                      </div>
                    </div>

                    <div
                      v-else-if="field.type === 'stringslice'"
                      :key="`${field.name}-stringslice`"
                      class="mb-3"
                    >
                      <label
                        class="form-label"
                        :for="`${models.rule.uuid}-action-${idx}-${field.key}`"
                      >{{ field.name }}</label>
                      <TagInput
                        :id="`${models.rule.uuid}-action-${idx}-${field.key}`"
                        v-model="models.rule.actions[idx].attributes[field.key] as string[]"
                        :state="validateActionArgument(idx, field.key)"
                        placeholder="Enter elements and press enter to add the element"
                      />
                      <div class="form-text">
                        <font-awesome-icon
                          v-if="field.support_template"
                          fixed-width
                          class="me-1 text-success"
                          :icon="['fas', 'code']"
                          title="Supports Templating"
                        />
                        {{ field.description }}
                      </div>
                    </div>

                    <div
                      v-else-if="field.support_template"
                      :key="`${field.name}-tpl`"
                      class="mb-3"
                    >
                      <label
                        class="form-label"
                        :for="`${models.rule.uuid}-action-${idx}-${field.key}`"
                      >{{ field.name }}</label>
                      <template-editor
                        :id="`${models.rule.uuid}-action-${idx}-${field.key}`"
                        v-model="models.rule.actions[idx].attributes[field.key] as string"
                        :state="validateActionArgument(idx, field.key)"
                        @valid-template="(valid: boolean) => updateTemplateValid(`${models.rule.uuid}-action-${idx}-${field.key}`, valid)"
                      />
                      <div class="form-text">
                        <font-awesome-icon
                          fixed-width
                          class="me-1 text-success"
                          :icon="['fas', 'code']"
                          title="Supports Templating"
                        />
                        {{ field.description }}
                      </div>
                    </div>

                    <div
                      v-else
                      :key="field.name"
                      class="mb-3"
                    >
                      <label
                        class="form-label"
                        :for="`${models.rule.uuid}-action-${idx}-${field.key}`"
                      >{{ field.name }}</label>
                      <input
                        :id="`${models.rule.uuid}-action-${idx}-${field.key}`"
                        v-model="models.rule.actions[idx].attributes[field.key]"
                        class="form-control"
                        :class="{
                          'is-invalid': validateActionArgument(idx, field.key) === false,
                          'is-valid': validateActionArgument(idx, field.key) === true,
                        }"
                        :placeholder="field.default"
                        :required="!field.optional"
                        type="text"
                      >
                      <div class="form-text">
                        {{ field.description }}
                      </div>
                    </div>
                  </template>
                </div>
                <div
                  v-else
                  class="card-body"
                >
                  This action has no attributes.
                </div>
              </div>
            </div>

            <hr>

            <div class="mb-3">
              <label
                class="form-label"
                for="ruleAddAction"
              >Add Action</label>
              <div class="input-group">
                <select
                  id="ruleAddAction"
                  v-model="models.addAction"
                  class="form-select"
                >
                  <option
                    v-for="option in availableActionsForAdd"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.text }}
                  </option>
                </select>
                <button
                  type="button"
                  class="btn btn-success"
                  :disabled="!models.addAction"
                  @click="addAction"
                >
                  <font-awesome-icon
                    fixed-width
                    class="me-1"
                    :icon="['fas', 'plus']"
                  />
                  Add
                </button>
              </div>
              <div class="form-text">
                {{ addActionDescription }}
              </div>
            </div>
            <div />
          </div>
        </div>
      </div>
    </AppModal>
  </div>
</template>

<script lang="ts">
import * as constants from '../lib/const'
import type { ActionDocumentation, Rule } from '../types'
import { api } from '../api'
import AppModal from '../components/AppModal.vue'
import { confirmDialog } from '../lib/confirmModal'
import { defineComponent } from 'vue'
import TagInput from '../components/TagInput.vue'
import TemplateEditor from '../components/tplEditor.vue'
import { useAppStore } from '../stores/app'

type RuleAttributeValue = string | number | boolean | string[] | null | undefined

type RuleActionForm = {
  type: string
  attributes: Record<string, RuleAttributeValue>
}

type RuleModel = Omit<Rule, 'actions' | 'cooldown' | 'channel_cooldown' | 'user_cooldown'> & {
  actions: RuleActionForm[]
  cooldown?: string
  channel_cooldown?: string
  user_cooldown?: string
  match_message__validation?: boolean
}

export default defineComponent({
  components: {
    AppModal,
    TagInput,
    TemplateEditor,
  },

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
        ...(this.appStore.vars.KnownEvents || []).map((event: string) => ({ text: event, value: event })),
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

    countRuleExceptions() {
      return (this.models.rule?.disable_on_match_messages || []).length
    },

    countRuleMatchers() {
      let count = 0
      count += this.models.rule.match_channels ? 1 : 0
      count += this.models.rule.match_event ? 1 : 0
      count += this.models.rule.match_message ? 1 : 0
      count += this.models.rule.match_users ? 1 : 0
      return count
    },

    filteredRules() {
      const rules = [...this.rules]
        .filter(rule => !this.filter || rule.description?.toLocaleLowerCase().includes(this.filter?.toLocaleLowerCase()))
      rules.sort((a, b) => {
        const ad = a.description?.toLocaleLowerCase() || ''
        const bd = b.description?.toLocaleLowerCase() || ''
        return ad.localeCompare(bd)
      })
      return rules
    },
  },

  data() {
    return {
      actions: [] as ActionDocumentation[],
      activeRuleTab: 'matcher',
      appStore: useAppStore(),
      filter: '',
      models: {
        addAction: '',
        addException: '',
        addException__validation: false,
        rule: {} as RuleModel,
        subscriptionURL: '',
      },

      openActionIndex: null as number | null,
      rules: [] as Rule[],
      showRuleEditModal: false,
      showRuleSubscribeModal: false,
      templateValid: {} as Record<string, boolean>,
      validateReason: null as string | null,
    }
  },

  methods: {
    actionHasValidationError(idx: number): boolean {
      const action = this.models.rule.actions?.[idx]
      const def = this.getActionDefinitionByType(action?.type || '')

      for (const field of def?.fields || []) {
        if (!this.validateActionArgument(idx, field.key)) {
          return true
        }
      }

      return false
    },

    addAction(): void {
      if (!this.models.rule.actions) {
        this.models.rule.actions = []
      }

      this.models.rule.actions.push({ attributes: {}, type: this.models.addAction } as RuleActionForm)
    },

    async addException() {
      if (await this.validateRegex(this.models.addException, false)) {
        if (!this.models.rule.disable_on_match_messages) {
          this.models.rule.disable_on_match_messages = []
        }
        this.models.rule.disable_on_match_messages.push(this.models.addException)
        this.models.addException = ''
      }
    },

    deleteRule(uuid: string) {
      confirmDialog('Do you really want to delete this rule?', {
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

          return api.delete(`config-editor/rules/${uuid}`)
            .then(() => {
              this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
            })
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    editRule(msg: Rule) {
      this.models.rule = {
        ...msg,
        actions: msg.actions?.map(action => ({ ...action, attributes: action.attributes || {} } as RuleActionForm)) || [],
        channel_cooldown: this.fixDurationRepresentationToString(msg.channel_cooldown),
        cooldown: this.fixDurationRepresentationToString(msg.cooldown),
        user_cooldown: this.fixDurationRepresentationToString(msg.user_cooldown),
      }
      this.templateValid = {}
      this.activeRuleTab = 'matcher'
      this.showRuleEditModal = true
      this.openActionIndex = 0
      this.validateMatcherRegex()
    },

    async fetchActions() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      try {
        const resp = await api.get<ActionDocumentation[]>('config-editor/actions', false)
        this.actions = resp || []
      } catch (err) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    async fetchRules() {
      this.$bus.$emit(constants.NOTIFY_LOADING_DATA, true)
      try {
        const resp = await api.get<Rule[]>('config-editor/rules')
        this.rules = resp || []
        this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, false)
      } catch (err) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    fixDurationRepresentationToInt64(value: number | string) {
      let match: RegExpMatchArray | null

      switch (typeof value) {
      case 'string':
        match = value.match(/(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?/)
        if (!match) {
          throw new Error(`invalid duration value: ${value}`)
        }

        return ((Number(match[1]) || 0) * 3600 + (Number(match[2]) || 0) * 60 + (Number(match[3]) || 0)) * constants.NANO

      default:
        return value
      }
    },

    fixDurationRepresentationToString(value: number | string | undefined) {
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

      case 'undefined':
        return ''

      default:
        return value
      }
    },

    formatRuleActions(rule: Rule): string[] {
      const badges: string[] = []

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

    formatRuleMatch(rule: Rule): {key: string, value: string}[] {
      const badges: {key: string, value: string}[] = []

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

    getActionDefinitionByType(type: string): ActionDocumentation {
      const def = this.actions.find(ad => ad.type === type)
      if (!def) {
        throw new Error(`unknown action type: ${type}`)
      }

      return def
    },

    moveAction(idx: number, direction: number): void {
      const targetIdx = idx + direction
      const actions = this.models.rule.actions

      if (!actions?.[idx] || !actions[targetIdx]) {
        return
      }

      [actions[idx], actions[targetIdx]] = [actions[targetIdx], actions[idx]]
    },

    newRule() {
      this.models.rule = { match_message__validation: true } as RuleModel
      this.templateValid = {}
      this.activeRuleTab = 'matcher'
      this.openActionIndex = null
      this.showRuleEditModal = true
    },

    removeAction(idx: number) {
      this.models.rule.actions = this.models.rule.actions?.filter((_, i) => i !== idx) || []
    },

    removeException(ex: string) {
      this.models.rule.disable_on_match_messages = this.models.rule.disable_on_match_messages?.filter(r => r !== ex) || []
    },

    saveRule(evt: Event) {
      if (!this.validateRule()) {
        evt.preventDefault()
        return
      }

      const obj: Rule = {
        ...this.models.rule,
        actions: this.models.rule.actions?.map(action => ({
          ...action,
          attributes: Object.fromEntries(Object.entries(action.attributes)
            .filter(att => {
              const def = this.getActionDefinitionByType(action.type)
              const field = def?.fields.filter(field => field.key === att[0])[0]

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
                if (att[1] === false && field.optional) {
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
                if (!Array.isArray(att[1]) || att[1].length === 0) {
                  return false
                }
                break
              }

              return true
            })) || [],
        })),

        match_message: this.models.rule.match_message === '' ? null : this.models.rule.match_message,
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

      let promise: Promise<unknown>
      if (obj.uuid) {
        promise = api.put(`config-editor/rules/${obj.uuid}`, obj)
      } else {
        promise = api.post(`config-editor/rules`, obj)
      }

      promise.then(() => {
        this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
      })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    subscribeRule() {
      api.post(`config-editor/rules`, {
        subscribe_from: this.models.subscriptionURL,
      })
        .then(() => {
          this.models.subscriptionURL = ''
          this.$bus.$emit(constants.NOTIFY_CHANGE_PENDING, true)
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    updateTemplateValid(id: string, valid: boolean) {
      this.templateValid[id] = valid
    },

    validateActionArgument(idx: number, key: string) {
      const action = this.models.rule.actions?.[idx]
      const def = this.getActionDefinitionByType(action?.type || '')

      if (!def || !def.fields) {
        return false
      }

      for (const field of def.fields) {
        if (field.key !== key) {
          continue
        }

        switch (field.type) {
        case 'bool':
          if (!field.optional && typeof action?.attributes[field.key] !== 'boolean') {
            return false
          }
          break

        case 'duration':
          if (!this.validateDuration(action.attributes[field.key] as string, !field.optional)) {
            return false
          }
          break

        case 'int64':
          if (!field.optional && !action?.attributes[field.key]) {
            return false
          }

          if (action?.attributes[field.key] && isNaN(action.attributes[field.key] as number)) {
            return false
          }

          break

        case 'string':
          if (!field.optional && !action?.attributes[field.key]) {
            return false
          }
          break

        case 'stringslice':
          if (!field.optional && !action?.attributes[field.key]) {
            return false
          }
          break
        }
        break
      }

      return true
    },

    validateDuration(duration: string | undefined, required: boolean): boolean {
      if (!duration && !required) {
        return true
      }

      if (!duration && required) {
        return false
      }

      return Boolean(duration!.match(/^(?:\d+(?:s|m|h))+$/))
    },

    async validateExceptionRegex() {
      const res = await this.validateRegex(this.models.addException, false)
      this.models.addException__validation = res
    },

    validateMatcherRegex() {
      if (this.models.rule.match_message === '') {
        this.models.rule.match_message__validation = true
        return
      }

      return this.validateRegex(this.models.rule.match_message!, true)
        .then(res => {
          this.models.rule.match_message__validation = res
        })
    },

    async validateRegex(regex: string, allowEmpty: boolean = true): Promise<boolean> {
      if (regex === '' && !allowEmpty) {
        return new Promise(resolve => {
          resolve(false)
        })
      }

      try {
        await api.put(`config-editor/validate-regex?regexp=${encodeURIComponent(regex)}`, undefined, false)
        return true
      } catch {
        return false
      }
    },

    validateRule() {
      if (Object.entries(this.templateValid).filter(e => !e[1]).length > 0) {
        this.validateReason = 'templateValid'
        return false
      }

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
          if (!field.optional && action.attributes[field.key] === undefined) {
            this.validateReason = `${action.type} -> ${field.key} -> opt`
            return false
          }

          if (field.type === 'duration' && !this.validateDuration(action.attributes[field.key] as string, !field.optional)) {
            this.validateReason = `${action.type} -> ${field.key} -> duration`
            return false
          }
        }
      }

      return true
    },

    validateTwitchBadge(tag: string): boolean {
      return this.appStore.vars.IRCBadges.includes(tag)
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

  name: 'TwitchBotRulesView',

  watch: {
    'models.rule.match_message'(to, from) {
      if (to === from) {
        return
      }

      this.validateMatcherRegex()
    },
  },
})
</script>

<style scoped>
.rule-action-icon-btn {
  flex: 0 0 2.5rem;
  padding-left: 0;
  padding-right: 0;
}
</style>
