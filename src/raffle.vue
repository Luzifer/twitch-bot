<template>
  <div>
    <b-row>
      <b-col>
        <b-table
          :busy="!raffleTableItems"
          :fields="raffleFields"
          hover
          :items="raffleTableItems"
          striped
          primary-key="id"
        >
          <template #cell(_actions)="data">
            <b-button-group size="sm">
              <b-button
                v-if="data.item.status === 'planned'"
                variant="success"
                title="Start Raffle"
                @click="startRaffle(data.item.id)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'play']"
                />
              </b-button>
              <b-button
                v-else-if="data.item.status === 'active'"
                variant="warning"
                title="Close Raffle"
                @click="closeRaffle(data.item.id)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'stop']"
                />
              </b-button>
              <b-button
                v-else-if="data.item.status === 'ended'"
                variant="warning"
                title="Re-Open Raffle"
                @click="reopenRaffle(data.item.id)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'play']"
                />
              </b-button>

              <b-button
                variant="info"
                :disabled="data.item.status === 'planned'"
                title="Manage Entries"
                @click="showEntryDialog(data.item.id)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'people-group']"
                />
              </b-button>

              <b-button
                variant="primary"
                title="Edit Raffle"
                @click="editRaffle(data.item)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'pen']"
                />
              </b-button>

              <b-button
                title="Duplicate Raffle"
                @click="cloneRaffle(data.item.id)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'clone']"
                />
              </b-button>

              <b-button
                variant="danger"
                title="Delete Raffle"
                @click="deleteRaffle(data.item.id)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'minus']"
                />
              </b-button>
            </b-button-group>
          </template>

          <template #cell(_status)="data">
            <font-awesome-icon
              v-if="data.item.status == 'planned'"
              fixed-width
              :icon="['fas', 'pen-ruler']"
              title="Planned"
            />
            <span v-else-if="data.item.status == 'active'">
              <font-awesome-icon
                fixed-width
                :icon="['fas', 'spinner']"
                spin-pulse
                title="In Progress"
                class="mr-1"
              />
              <small class="text-muted">{{ raffleTimer(data.item) }}</small>
            </span>
            <font-awesome-icon
              v-else-if="data.item.status == 'ended'"
              fixed-width
              :icon="['fas', 'stop']"
              title="Ended"
            />
            <font-awesome-icon
              v-else
              fixed-width
              :icon="['fas', 'question']"
            />
          </template>

          <template #head(_actions)="">
            <b-button-group size="sm">
              <b-button
                variant="success"
                title="Create New Raffle"
                @click="newRaffle"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'plus']"
                />
              </b-button>

              <b-button
                variant="secondary"
                title="Refresh Raffle"
                @click="fetchRaffles"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'rotate']"
                />
              </b-button>
            </b-button-group>
          </template>
        </b-table>
      </b-col>
    </b-row>

    <!-- Entries Modal -->
    <b-modal
      v-if="showRaffleEntriesModal"
      scrollable
      size="lg"
      :visible="showRaffleEntriesModal"
      :title="`Entries: ${openedRaffle.title}`"
      hide-footer
      @hidden="closeEntriesModal"
    >
      <b-list-group
        v-if="raffleEntries.length > 0"
        flush
      >
        <b-list-group-item class="d-flex justify-content-between align-items-center">
          <strong>{{ raffleEntries.length }} entrant(s)</strong>
          <b-button-group size="sm">
            <b-button
              variant="success"
              :disabled="openedRaffle.status !== 'ended'"
              @click="pickWinner"
            >
              <font-awesome-icon
                fixed-width
                :icon="['fas', 'crown']"
              />
              Pick Winner
            </b-button>
            <b-button
              variant="secondary"
              title="Refresh Entries"
              @click="refreshOpenendRaffle"
            >
              <font-awesome-icon
                fixed-width
                :icon="['fas', 'rotate']"
                :spin="openedRaffleReloading"
              />
            </b-button>
          </b-button-group>
        </b-list-group-item>
        <b-list-group-item
          v-for="entry in raffleEntries"
          :key="entry.id"
        >
          <div class="d-flex align-items-center">
            <font-awesome-layers
              v-if="entry.wasPicked"
              class="mr-1"
              :title="entry.wasRedrawn ? 'Was Redrawn' : 'Has Won'"
            >
              <font-awesome-icon
                fixed-width
                :icon="['fas', 'crown']"
                :class="entry.wasRedrawn ? 'text-muted' : 'text-warning'"
              />
              <font-awesome-icon
                v-if="entry.wasRedrawn"
                fixed-width
                :icon="['fas', 'slash']"
                class="text-danger"
              />
            </font-awesome-layers>

            <font-awesome-icon
              v-if="entry.enteredAs === 'everyone'"
              fixed-width
              :icon="['fas', 'user']"
              title="Does not Follow"
            />
            <font-awesome-icon
              v-else-if="entry.enteredAs === 'follower'"
              fixed-width
              :icon="['fas', 'heart']"
              title="Follower"
            />
            <font-awesome-icon
              v-else-if="entry.enteredAs === 'subscriber'"
              fixed-width
              :icon="['fas', 'star']"
              title="Subscriber"
            />
            <font-awesome-icon
              v-else-if="entry.enteredAs === 'vip'"
              fixed-width
              :icon="['fas', 'gem']"
              title="VIP"
            />
            <font-awesome-icon
              v-else
              fixed-width
              :icon="['fas', 'question']"
            />

            <span class="ml-1">{{ entry.userDisplayName }}</span>
            <b-badge
              variant="secondary"
              class="ml-auto"
            >
              {{ new Date(entry.enteredAt).toLocaleString() }}
            </b-badge>
          </div>
          <b-row>
            <b-col cols="11">
              <pre
                v-if="entry.drawResponse"
                class="mb-0 mt-2"
              ><code>{{ entry.drawResponse }}</code></pre>
            </b-col>
            <b-col cols="1">
              <b-button
                v-if="entry.wasPicked && !entry.wasRedrawn"
                variant="danger"
                size="sm"
                title="Re-Draw Winner"
                @click="repickWinner(entry.id)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'recycle']"
                />
              </b-button>
            </b-col>
          </b-row>
        </b-list-group-item>
      </b-list-group>
      <p
        v-else
        class="mb-0"
      >
        No one entered into this raffle.
      </p>
    </b-modal>

    <!-- Raffle Editor -->
    <b-modal
      v-if="showRaffleEditModal"
      hide-header-close
      :ok-disabled="!validateRaffle()"
      ok-title="Save"
      scrollable
      size="xl"
      :visible="showRaffleEditModal"
      title="Edit Raffle"
      @hidden="showRaffleEditModal=false"
      @ok="saveRaffle"
    >
      <b-row>
        <!-- Keyword / Title -->
        <b-col cols="3">
          <b-form-group
            description="Where to do the raffle?"
            label="Channel"
            label-for="formRaffleChannel"
          >
            <b-input-group prepend="#">
              <b-form-input
                id="formRaffleChannel"
                v-model="models.raffle.channel"
                type="text"
                :state="validateRaffleChannel()"
                :disabled="models.raffle.status !== 'planned'"
              />
            </b-input-group>
          </b-form-group>
        </b-col>
        <b-col cols="3">
          <b-form-group
            description="Keyword to use in chat to enter"
            label="Keyword"
            label-for="formRaffleKeyword"
          >
            <b-form-input
              id="formRaffleKeyword"
              v-model="models.raffle.keyword"
              type="text"
              :state="validateRaffleNonEmpty(models.raffle.keyword)"
              :disabled="models.raffle.status !== 'planned'"
            />
          </b-form-group>
        </b-col>
        <b-col cols="6">
          <b-form-group
            description="Title of the raffle (displayed in overview and available in chat)"
            label="Title"
            label-for="formRaffleTitle"
          >
            <b-form-input
              id="formRaffleTitle"
              v-model="models.raffle.title"
              type="text"
              :state="validateRaffleNonEmpty(models.raffle.title)"
            />
          </b-form-group>
        </b-col>
      </b-row>

      <b-row class="mt-3">
        <!-- Allow / Multiplier / Times -->
        <b-col>
          <b-card
            header="Allowed Entries"
            no-body
          >
            <b-list-group flush>
              <b-list-group-item>
                <b-form-checkbox
                  v-model="models.raffle.allowEveryone"
                  switch
                  :state="validateRaffleHasAllowedEntries()"
                  :disabled="models.raffle.status !== 'planned'"
                >
                  Everyone
                </b-form-checkbox>
              </b-list-group-item>
              <b-list-group-item>
                <b-form
                  inline
                  class="d-flex justify-content-between"
                >
                  <b-form-checkbox
                    v-model="models.raffle.allowFollower"
                    switch
                    :state="validateRaffleHasAllowedEntries()"
                    :disabled="models.raffle.status !== 'planned'"
                  >
                    Followers, since
                  </b-form-checkbox>
                  <b-input-group
                    size="sm"
                    append="min"
                    class="col-5 px-0"
                  >
                    <b-form-input
                      v-model="models.raffle.minFollowAge"
                      class="text-right"
                      type="number"
                      placeholder="min. Age"
                      min="0"
                      :state="validateRaffleIsNumber(models.raffle.minFollowAge)"
                      :disabled="models.raffle.status !== 'planned'"
                    />
                  </b-input-group>
                </b-form>
              </b-list-group-item>
              <b-list-group-item>
                <b-form-checkbox
                  v-model="models.raffle.allowSubscriber"
                  switch
                  :state="validateRaffleHasAllowedEntries()"
                  :disabled="models.raffle.status !== 'planned'"
                >
                  Subscribers
                </b-form-checkbox>
              </b-list-group-item>
              <b-list-group-item>
                <b-form-checkbox
                  v-model="models.raffle.allowVIP"
                  switch
                  :state="validateRaffleHasAllowedEntries()"
                  :disabled="models.raffle.status !== 'planned'"
                >
                  VIPs
                </b-form-checkbox>
              </b-list-group-item>
            </b-list-group>
          </b-card>
        </b-col>

        <b-col>
          <b-card header="Luck Modifiers">
            <b-row>
              <b-col>
                <b-form-group
                  label="Followers"
                  label-for="formRaffleMultiFollower"
                >
                  <b-form-input
                    id="formRaffleMultiFollower"
                    v-model="models.raffle.multiFollower"
                    type="number"
                    min="0"
                    step="0.1"
                    size="sm"
                    :state="validateRaffleIsNumber(models.raffle.multiFollower)"
                    :disabled="models.raffle.status !== 'planned'"
                  />
                </b-form-group>
              </b-col>

              <b-col>
                <b-form-group
                  label="Subs"
                  label-for="formRaffleMultiSubs"
                >
                  <b-form-input
                    id="formRaffleMultiSubs"
                    v-model="models.raffle.multiSubscriber"
                    type="number"
                    min="0"
                    step="0.1"
                    size="sm"
                    :state="validateRaffleIsNumber(models.raffle.multiSubscriber)"
                    :disabled="models.raffle.status !== 'planned'"
                  />
                </b-form-group>
              </b-col>

              <b-col>
                <b-form-group
                  label="VIPs"
                  label-for="formRaffleMultiVIP"
                >
                  <b-form-input
                    id="formRaffleMultiVIP"
                    v-model="models.raffle.multiVIP"
                    type="number"
                    min="0"
                    step="0.1"
                    size="sm"
                    :state="validateRaffleIsNumber(models.raffle.multiVIP)"
                    :disabled="models.raffle.status !== 'planned'"
                  />
                </b-form-group>
              </b-col>
            </b-row>
            <b-row>
              <b-col>
                <b-form-text>
                  The base amount of tickets for <strong>Everyone</strong> is always <code>1.0</code>. You can
                  increase chances for certain user groups. Groups are checked from right to left: If the user
                  is VIP, Sub and Follower they are assigned <strong>VIPkeyword</strong> multiplier.
                </b-form-text>
              </b-col>
            </b-row>
            <!-- -->
          </b-card>
        </b-col>

        <b-col>
          <b-card
            header="Times"
            no-body
          >
            <b-list-group flush>
              <b-list-group-item class="d-flex justify-content-between align-items-center">
                <span>Auto-Start*</span>
                <b-form-input
                  v-model="models.raffle.autoStartAt"
                  class="col-7"
                  type="datetime-local"
                  size="sm"
                  :min="transformISOToDateTimeLocal(new Date())"
                  :disabled="models.raffle.status !== 'planned'"
                />
              </b-list-group-item>

              <b-list-group-item class="d-flex justify-content-between align-items-center">
                <span>Duration</span>
                <b-input-group
                  size="sm"
                  append="min"
                  class="col-5 px-0"
                >
                  <b-form-input
                    v-model="models.raffle.closeAfter"
                    type="number"
                    step="1"
                    min="0"
                    class="text-right"
                    :state="validateRaffleIsNumber(models.raffle.closeAfter)"
                  />
                </b-input-group>
              </b-list-group-item>

              <b-list-group-item class="d-flex justify-content-between align-items-center">
                <span>Close At*</span>
                <b-form-input
                  v-model="models.raffle.closeAt"
                  class="col-7"
                  type="datetime-local"
                  size="sm"
                  :min="transformISOToDateTimeLocal(new Date())"
                />
              </b-list-group-item>

              <b-list-group-item class="d-flex justify-content-between align-items-center">
                <span>Respond in</span>
                <b-input-group
                  size="sm"
                  append="sec"
                  class="col-5 px-0"
                >
                  <b-form-input
                    v-model="models.raffle.waitForResponse"
                    type="number"
                    step="1"
                    min="0"
                    class="text-right"
                    :state="validateRaffleIsNumber(models.raffle.waitForResponse)"
                  />
                </b-input-group>
              </b-list-group-item>

              <b-list-group-item>
                <b-form-text>
                  * Optional, when <strong>Close At</strong> is specified, <strong>Duration</strong> is ignored,
                  when no <strong>Auto-Start</strong> is specified the raffle must be started manually. If you
                  set <strong>Respond In</strong> to 0, no chat responses from that user are recorded after
                  picking them as winner.
                </b-form-text>
              </b-list-group-item>
            </b-list-group>
          </b-card>
        </b-col>
      </b-row>

      <b-row class="mt-3">
        <!-- Texts / Entries -->
        <b-col>
          <b-card
            header="Texts"
            no-body
          >
            <b-list-group flush>
              <b-list-group-item class="">
                <div class="d-flex align-items-center mb-1">
                  <b-form-checkbox
                    v-model="models.raffle.textEntryPost"
                    switch
                    class="mr-n2"
                  >
                    Message on successful entry
                  </b-form-checkbox>
                </div>
                <template-editor
                  v-model="models.raffle.textEntry"
                />
              </b-list-group-item>

              <b-list-group-item class="">
                <div class="d-flex align-items-center mb-1">
                  <b-form-checkbox
                    v-model="models.raffle.textEntryFailPost"
                    switch
                    class="mr-n2"
                  >
                    Message on failed entry
                  </b-form-checkbox>
                </div>
                <template-editor
                  v-model="models.raffle.textEntryFail"
                />
              </b-list-group-item>

              <b-list-group-item class="">
                <div class="d-flex align-items-center mb-1">
                  <b-form-checkbox
                    v-model="models.raffle.textWinPost"
                    switch
                    class="mr-n2"
                  >
                    Message on winner draw
                  </b-form-checkbox>
                </div>
                <template-editor
                  v-model="models.raffle.textWin"
                />
              </b-list-group-item>

              <b-list-group-item class="">
                <div class="d-flex justify-content-between align-items-center mb-1">
                  <b-form-checkbox
                    v-model="models.raffle.textReminderPost"
                    switch
                    class="mr-n2"
                  >
                    Periodic reminder
                  </b-form-checkbox>
                  <b-input-group
                    prepend="every"
                    append="min"
                    size="sm"
                    class="col-2 px-0"
                  >
                    <b-form-input
                      v-model="models.raffle.textReminderInterval"
                      type="number"
                      step="1"
                      min="0"
                      class="text-right"
                      :state="validateRaffleIsNumber(models.raffle.textReminderInterval)"
                    />
                  </b-input-group>
                </div>
                <template-editor
                  v-model="models.raffle.textReminder"
                />
              </b-list-group-item>
              <b-list-group-item>
                <b-form-text>
                  Available variables are <code>.user</code> and <code>.raffle</code> with
                  access to all of these configurations most notably <code>.raffle.Title</code>
                  and <code>.raffle.Keyword</code>.
                </b-form-text>
              </b-list-group-item>
            </b-list-group>
          </b-card>
        </b-col>
      </b-row>
    </b-modal>
  </div>
</template>

<script>
import * as constants from './const.js'

import axios from 'axios'
import TemplateEditor from './tplEditor.vue'

const ONE_MINUTE = 60000000000 // nanoseconds
const ONE_SECOND = 1000000000 // nanoseconds

export default {
  components: { TemplateEditor },
  computed: {
    raffleEntries() {
      const entries = [...this.openedRaffle.entries || []]
      entries.sort((a, b) => {
        const isWinner = e => e.wasPicked && !e.wasRedrawn
        const wasWinnner = e => e.wasPicked && e.wasRedrawn

        if (isWinner(a) && !isWinner(b)) {
          return -1
        } else if (isWinner(b) && !isWinner(a)) {
          return 1
        } else if (wasWinnner(a) && !wasWinnner(b)) {
          return -1
        } else if (wasWinnner(b) && !wasWinnner(a)) {
          return 1
        }

        // Everything else: Order by ID DESC
        return b.id - a.id
      })

      return entries
    },

    raffleTableItems() {
      const raffles = [...this.raffles]
      raffles.sort((a, b) => {
        const scores = { active: 0, ended: 2, planned: 1 }

        if (scores[a.status] !== scores[b.status]) {
          return scores[a.status] - scores[b.status]
        }

        return b.id - a.id
      })
      return raffles
    },
  },

  data() {
    return {
      models: { raffle: {} },
      now: new Date(),
      openedRaffle: {},
      openedRaffleReloading: false,
      raffleFields: [
        {
          class: 'col-1',
          key: '_status',
          label: 'Status',
          tdClass: 'text-center align-middle',
          thClass: 'align-middle text-center',
        },
        {
          class: 'col-1',
          key: 'channel',
          label: 'Channel',
          tdClass: 'align-middle',
          thClass: 'align-middle',
        },
        {
          class: 'col-1',
          key: 'keyword',
          label: 'Keyword',
          tdClass: 'align-middle',
          thClass: 'align-middle',
        },
        {
          class: 'col-7',
          key: 'title',
          label: 'Title',
          tdClass: 'align-middle',
          thClass: 'align-middle',
        },
        {
          class: 'col-2 text-right',
          key: '_actions',
          label: '',
          tdClass: 'align-middle',
          thClass: 'align-middle',
        },
      ],

      raffles: [],
      reopenRaffleDuration: 60,
      showRaffleEditModal: false,
      showRaffleEntriesModal: false,
    }
  },

  methods: {
    cloneRaffle(id) {
      return axios.put(`raffle/${id}/clone`, {}, this.$root.axiosOptions)
        .then(() => this.$root.toastSuccess('Raffle cloned'))
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    closeEntriesModal() {
      this.openedRaffle = {}
      this.showRaffleEntriesModal = false
    },

    closeRaffle(id) {
      this.$bvModal.msgBoxConfirm('Do you really want to close entries for this raffle?', {
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

          return axios.put(`raffle/${id}/close`, {}, this.$root.axiosOptions)
            .then(() => this.$root.toastSuccess('Raffle closed'))
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    deleteRaffle(id) {
      this.$bvModal.msgBoxConfirm('Do you really want to delete this raffle and all its entries?', {
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

          return axios.delete(`raffle/${id}`, this.$root.axiosOptions)
            .then(() => this.$root.toastSuccess('Raffle deleted'))
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    editRaffle(raffle) {
      this.$set(this.models, 'raffle', raffle)
      this.showRaffleEditModal = true
    },

    fetchRaffles() {
      return axios.get('raffle/', this.$root.axiosOptions)
        .then(resp => {
          this.raffles = resp.data.map(raffle => this.transformRaffleFromDB(raffle))
        })
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    newRaffle() {
      this.$set(this.models, 'raffle', {
        /* eslint-disable sort-keys */
        channel: '',
        keyword: '!enter',
        status: 'planned',
        title: '',

        allowEveryone: false,
        allowFollower: true,
        allowSubscriber: true,
        allowVIP: true,
        minFollowAge: 0,

        multiFollower: 1.0,
        multiSubscriber: 1.0,
        multiVIP: 1.0,

        autoStartAt: null,
        closeAfter: 0,
        closeAt: null,
        waitForResponse: 30,

        textEntry: '{{ mention .user }} you are registered, good luck!',
        textEntryPost: true,

        textEntryFail: '{{ mention .user }} couldn\'t register you, sorry.',
        textEntryFailPost: false,

        textWin: '{{ mention .user }} you won! Please speak up in chat to claim your price!',
        textWinPost: false,

        textReminder: 'We are currently doing a raffle until {{ dateInZone "15:04" .raffle.CloseAt "Europe/Berlin" }}: "{{ .raffle.Title }}" - type "{{ .raffle.Keyword }}" to enter!',
        textReminderInterval: 15,
        textReminderPost: false,
        /* eslint-enable sort-keys */
      })
      this.showRaffleEditModal = true
    },

    pickWinner() {
      this.openedRaffleReloading = true
      return axios.put(`raffle/${this.openedRaffle.id}/pick`, {}, this.$root.axiosOptions)
        .then(() => this.$root.toastSuccess('Winner picked!'))
        .catch(() => this.$root.toastError('Could not pick winner!'))
    },

    raffleTimer(raffle) {
      const parts = []
      let tte = new Date(raffle.closeAt) - this.now

      for (const d of [3600000, 60000, 1000]) {
        const pt = Math.floor(tte / d)
        parts.push(String(pt).padStart(2, '0'))
        tte -= pt * d
      }

      return parts.join(':')
    },

    refreshOpenendRaffle() {
      this.openedRaffleReloading = true
      return axios.get(`raffle/${this.openedRaffle.id}`, this.$root.axiosOptions)
        .then(resp => {
          this.openedRaffle = resp.data
          this.openedRaffleReloading = false
        })
    },

    reopenRaffle(raffleId) {
      let duration = 10

      const h = this.$createElement
      const content = h('div', {}, [
        h('b-input-group', { props: { append: 'min' } }, [
          h('b-form-input', {
            class: 'text-right',
            on: {
              input(value) {
                duration = Number(value)
              },
            },
            props: {
              min: '0',
              step: '1',
              type: 'number',
              value: duration,
            },
          }),
        ]),
        h('b-form-text', { domProps: { innerHTML: 'The raffle will be re-opened and the "Close At" attribute will be set to the given duration from now.' } }),
        h('b-form-text', { domProps: { innerHTML: 'This will <strong>NOT</strong> clear the entrants, so don\'t use this for another raffle, use the "Duplicate Raffle" functionality for that.' } }),
      ])


      this.$bvModal.msgBoxConfirm([content], {
        buttonSize: 'sm',
        centered: true,
        size: 'sm',
        title: 'Re-Open the Raffle?',
      })
        .then(val => {
          if (!val) {
            return
          }

          return axios.put(`raffle/${raffleId}/reopen?duration=${duration * 60}`, {}, this.$root.axiosOptions)
            .then(() => this.$root.toastSuccess('Raffle re-opened'))
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    repickWinner(winnerId) {
      this.openedRaffleReloading = true
      return axios.put(`raffle/${this.openedRaffle.id}/repick/${winnerId}`, {}, this.$root.axiosOptions)
        .then(() => this.$root.toastSuccess('Winner re-picked!'))
        .catch(() => this.$root.toastError('Could not re-pick winner!'))
    },

    saveRaffle() {
      if (this.models.raffle.id) {
        return axios.put(`raffle/${this.models.raffle.id}`, this.transformRaffleToDB(this.models.raffle), this.$root.axiosOptions)
          .then(() => this.$root.toastSuccess('Raffle updated'))
          .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
      }

      return axios.post('raffle/', this.transformRaffleToDB(this.models.raffle), this.$root.axiosOptions)
        .then(() => this.$root.toastSuccess('Raffle created'))
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    showEntryDialog(raffleId) {
      this.openedRaffle = { id: raffleId }
      return this.refreshOpenendRaffle()
        .then(() => {
          this.showRaffleEntriesModal = true
        })
    },

    startRaffle(id) {
      return axios.put(`raffle/${id}/start`, {}, this.$root.axiosOptions)
        .then(() => this.$root.toastSuccess('Raffle started'))
        .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
    },

    transformISOToDateTimeLocal(t) {
      const d = new Date(t)
      d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
      return d.toISOString().slice(0, 16)
    },

    transformRaffleFromDB(raffle) {
      return {
        /* eslint-disable sort-keys */
        ...raffle,

        channel: raffle.channel.replace(/^#/, ''),

        // Transform durations
        closeAfter: raffle.closeAfter / ONE_MINUTE,
        minFollowAge: raffle.minFollowAge / ONE_MINUTE,
        textReminderInterval: raffle.textReminderInterval / ONE_MINUTE,
        waitForResponse: raffle.waitForResponse / ONE_SECOND,

        // Transform date values
        autoStartAt: raffle.autoStartAt ? this.transformISOToDateTimeLocal(raffle.autoStartAt) : '',
        closeAt: raffle.closeAt ? this.transformISOToDateTimeLocal(raffle.closeAt) : '',
        /* eslint-enable sort-keys */
      }
    },

    transformRaffleToDB(raffle) {
      return {
        /* eslint-disable sort-keys */
        ...raffle,

        channel: `#${raffle.channel.replace(/^#/, '')}`,

        // Transform durations
        closeAfter: Number(raffle.closeAfter) * ONE_MINUTE,
        minFollowAge: Number(raffle.minFollowAge) * ONE_MINUTE,
        textReminderInterval: Number(raffle.textReminderInterval) * ONE_MINUTE,
        waitForResponse: Number(raffle.waitForResponse) * ONE_SECOND,

        // Transform date values
        autoStartAt: !raffle.autoStartAt ? null : new Date(raffle.autoStartAt).toISOString(),
        closeAt: !raffle.closeAt ? null : new Date(raffle.closeAt).toISOString(),

        // Transform numeric values
        multiFollower: Number(raffle.multiFollower),
        multiSubscriber: Number(raffle.multiSubscriber),
        multiVIP: Number(raffle.multiVIP),
        /* eslint-enable sort-keys */
      }
    },

    validateRaffle() {
      if (this.models.raffle.status === 'ended') {
        // You must not modify running raffles
        return false
      }

      for (const nf of [
        'closeAfter', 'minFollowAge', 'textReminderInterval', 'waitForResponse',
        'multiFollower', 'multiSubscriber', 'multiVIP',
      ]) {
        // You must not put text in numeric fields
        if (this.validateRaffleIsNumber(this.models.raffle[nf]) === false) {
          return false
        }
      }

      if (this.validateRaffleChannel() === false) {
        return false
      }

      for (const nef of ['keyword', 'title']) {
        if (this.validateRaffleNonEmpty(this.models.raffle[nef]) === false) {
          // You must fill certain fields
          return false
        }
      }

      if (!this.models.raffle.closeAfter && !this.models.raffle.closeAt) {
        return false
      }

      if (this.validateRaffleHasAllowedEntries() === false) {
        // You must allow someone to enter
        return false
      }

      return true
    },

    validateRaffleChannel() {
      if (!/^[a-zA-Z0-9]{4,25}$/.test(this.models.raffle.channel)) {
        return false
      }
      return null
    },

    validateRaffleHasAllowedEntries() {
      if (this.models.raffle.allowEveryone || this.models.raffle.allowFollower || this.models.raffle.allowSubscriber || this.models.raffle.allowVIP) {
        return null
      }
      return false
    },

    validateRaffleIsNumber(n) {
      if (isNaN(Number(n))) {
        return false
      }
      return null
    },

    validateRaffleNonEmpty(str) {
      if (!str || str.trim().length === 0) {
        return false
      }
      return null
    },
  },

  mounted() {
    this.fetchRaffles()

    this.$bus.$on('raffleChanged', () => {
      this.fetchRaffles()

      if (this.openedRaffle?.id) {
        this.refreshOpenendRaffle()
      }
    })

    this.$bus.$on('raffleEntryChanged', () => {
      if (!this.openedRaffle?.id) {
        // We ignore this when there is no opened raffle
        return
      }

      this.refreshOpenendRaffle()
    })

    window.setInterval(() => {
      this.now = new Date()
    }, 1000)
  },

  name: 'TwitchBotEditorAppRaffle',
}
</script>
