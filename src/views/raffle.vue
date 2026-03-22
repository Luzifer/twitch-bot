<template>
  <div>
    <div class="row">
      <div class="col">
        <div class="table-responsive">
          <table class="table table-hover table-striped align-middle">
            <thead>
              <tr>
                <th class="col-1 text-center align-middle">
                  Status
                </th>
                <th class="col-1 align-middle">
                  Channel
                </th>
                <th class="col-1 align-middle">
                  Keyword
                </th>
                <th class="col-7 align-middle">
                  Title
                </th>
                <th class="col-2 text-end align-middle">
                  <div class="btn-group btn-group-sm">
                    <button
                      type="button"
                      class="btn btn-success"
                      title="Create New Raffle"
                      @click="newRaffle"
                    >
                      <font-awesome-icon
                        fixed-width
                        :icon="['fas', 'plus']"
                      />
                    </button>

                    <button
                      type="button"
                      class="btn btn-secondary"
                      title="Refresh Raffle"
                      @click="fetchRaffles"
                    >
                      <font-awesome-icon
                        fixed-width
                        :icon="['fas', 'rotate']"
                      />
                    </button>
                  </div>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!raffleTableItems">
                <td
                  colspan="5"
                  class="text-center text-muted"
                >
                  Loading...
                </td>
              </tr>
              <tr v-else-if="raffleTableItems.length === 0">
                <td
                  colspan="5"
                  class="text-center text-muted"
                >
                  No entries found.
                </td>
              </tr>
              <template v-else>
                <tr
                  v-for="raffle in raffleTableItems"
                  :key="raffle.id"
                >
                  <td class="text-center align-middle">
                    <font-awesome-icon
                      v-if="raffle.status == 'planned'"
                      fixed-width
                      :icon="['fas', 'pen-ruler']"
                      title="Planned"
                    />
                    <span v-else-if="raffle.status == 'active'">
                      <font-awesome-icon
                        fixed-width
                        :icon="['fas', 'spinner']"
                        spin-pulse
                        title="In Progress"
                        class="me-1"
                      />
                      <small class="text-muted">{{ raffleTimer(raffle) }}</small>
                    </span>
                    <font-awesome-icon
                      v-else-if="raffle.status == 'ended'"
                      fixed-width
                      :icon="['fas', 'stop']"
                      title="Ended"
                    />
                    <font-awesome-icon
                      v-else
                      fixed-width
                      :icon="['fas', 'question']"
                    />
                  </td>
                  <td class="align-middle">
                    {{ raffle.channel }}
                  </td>
                  <td class="align-middle">
                    {{ raffle.keyword }}
                  </td>
                  <td class="align-middle">
                    {{ raffle.title }}
                  </td>
                  <td class="text-end align-middle">
                    <div class="btn-group btn-group-sm">
                      <button
                        v-if="raffle.status === 'planned'"
                        type="button"
                        class="btn btn-success"
                        title="Start Raffle"
                        @click="startRaffle(raffle.id!)"
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'play']"
                        />
                      </button>
                      <button
                        v-else-if="raffle.status === 'active'"
                        type="button"
                        class="btn btn-warning"
                        title="Close Raffle"
                        @click="closeRaffle(raffle.id!)"
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'stop']"
                        />
                      </button>
                      <button
                        v-else-if="raffle.status === 'ended'"
                        type="button"
                        class="btn btn-warning"
                        title="Re-Open Raffle"
                        @click="reopenRaffle(raffle.id!)"
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'play']"
                        />
                      </button>

                      <button
                        type="button"
                        class="btn btn-info"
                        :disabled="raffle.status === 'planned'"
                        title="Manage Entries"
                        @click="showEntryDialog(raffle.id!)"
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'people-group']"
                        />
                      </button>

                      <button
                        type="button"
                        class="btn btn-primary"
                        title="Edit Raffle"
                        @click="editRaffle(raffle)"
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'pen']"
                        />
                      </button>

                      <button
                        type="button"
                        class="btn btn-secondary"
                        title="Duplicate Raffle"
                        @click="cloneRaffle(raffle.id!)"
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'clone']"
                        />
                      </button>

                      <button
                        type="button"
                        class="btn btn-warning"
                        :disabled="raffle.status !== 'ended'"
                        title="Reset Raffle"
                        @click="resetRaffle(raffle.id!)"
                      >
                        <font-awesome-icon
                          fixed-width
                          :icon="['fas', 'recycle']"
                        />
                      </button>

                      <button
                        type="button"
                        class="btn btn-danger"
                        title="Delete Raffle"
                        @click="deleteRaffle(raffle.id!)"
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
      v-if="showRaffleEntriesModal"
      :model-value="showRaffleEntriesModal"
      hide-footer
      scrollable
      size="lg"
      :title="`Entries: ${openedRaffle.title}`"
      @hidden="closeEntriesModal"
      @update:model-value="showRaffleEntriesModal = $event"
    >
      <div
        v-if="raffleEntries.length > 0"
        class="list-group list-group-flush"
      >
        <div class="list-group-item d-flex justify-content-between align-items-center">
          <strong>{{ raffleEntries.length }} entrant(s)</strong>
          <div class="btn-group btn-group-sm">
            <button
              type="button"
              class="btn btn-success"
              :disabled="openedRaffle.status !== 'ended'"
              @click="pickWinner"
            >
              <font-awesome-icon
                fixed-width
                :icon="['fas', 'crown']"
              />
              Pick Winner
            </button>
            <button
              type="button"
              class="btn btn-secondary"
              title="Refresh Entries"
              @click="refreshOpenendRaffle"
            >
              <font-awesome-icon
                fixed-width
                :icon="['fas', 'rotate']"
                :spin="openedRaffleReloading"
              />
            </button>
          </div>
        </div>
        <div
          v-for="entry in raffleEntries"
          :key="entry.id"
          class="list-group-item"
        >
          <div class="d-flex align-items-center">
            <span
              v-if="entry.wasPicked"
              class="position-relative d-inline-flex align-items-center me-1"
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
                class="text-danger position-absolute top-50 start-50 translate-middle"
              />
            </span>

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
              v-else-if="entry.enteredAs === 'reward'"
              fixed-width
              :icon="['fas', 'coins']"
              title="Subscriber"
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

            <span class="ms-1">{{ entry.userDisplayName }}</span>
            <span class="badge bg-secondary-subtle text-secondary-emphasis ms-auto">
              {{ new Date(entry.enteredAt).toLocaleString() }}
            </span>
          </div>
          <div class="row">
            <div class="col-11">
              <pre
                v-if="entry.drawResponse"
                class="mb-0 mt-2"
              ><code>{{ entry.drawResponse }}</code></pre>
            </div>
            <div class="col-1">
              <button
                v-if="entry.wasPicked && !entry.wasRedrawn"
                type="button"
                class="btn btn-danger btn-sm"
                title="Re-Draw Winner"
                @click="repickWinner(entry.id)"
              >
                <font-awesome-icon
                  fixed-width
                  :icon="['fas', 'recycle']"
                />
              </button>
            </div>
          </div>
        </div>
      </div>
      <p
        v-else
        class="mb-0"
      >
        No one entered into this raffle.
      </p>
    </AppModal>

    <AppModal
      v-if="showRaffleEditModal"
      :model-value="showRaffleEditModal"
      :ok-disabled="!validateRaffle()"
      ok-title="Save"
      scrollable
      size="xl"
      title="Edit Raffle"
      @hidden="showRaffleEditModal = false"
      @update:model-value="showRaffleEditModal = $event"
      @ok="saveRaffle"
    >
      <div class="row">
        <div class="col-3">
          <div class="mb-3">
            <label
              class="form-label"
              for="formRaffleChannel"
            >Channel</label>
            <div class="input-group">
              <span class="input-group-text">#</span>
              <input
                id="formRaffleChannel"
                v-model="models.raffle.channel"
                class="form-control"
                :class="{ 'is-invalid': validateRaffleChannel() === false }"
                :disabled="models.raffle.status !== 'planned'"
                type="text"
              >
            </div>
            <div class="form-text">
              Where to do the raffle?
            </div>
          </div>
        </div>
        <div class="col-3">
          <div class="mb-3">
            <label
              class="form-label"
              for="formRaffleKeyword"
            >Keyword</label>
            <input
              id="formRaffleKeyword"
              v-model="models.raffle.keyword"
              class="form-control"
              :class="{ 'is-invalid': validateRaffleNonEmpty(models.raffle.keyword) === false }"
              :disabled="models.raffle.status !== 'planned'"
              type="text"
            >
            <div class="form-text">
              Keyword to use in chat to enter
            </div>
          </div>
        </div>
        <div class="col-6">
          <div class="mb-3">
            <label
              class="form-label"
              for="formRaffleTitle"
            >Title</label>
            <input
              id="formRaffleTitle"
              v-model="models.raffle.title"
              class="form-control"
              :class="{ 'is-invalid': validateRaffleNonEmpty(models.raffle.title) === false }"
              type="text"
            >
            <div class="form-text">
              Title of the raffle (displayed in overview and available in chat)
            </div>
          </div>
        </div>
      </div>

      <div class="row mt-3">
        <div class="col">
          <div class="card">
            <div class="card-header">
              Allowed Entries
            </div>
            <div class="list-group list-group-flush">
              <div class="list-group-item">
                <div class="form-check form-switch">
                  <input
                    id="raffleAllowEveryone"
                    v-model="models.raffle.allowEveryone"
                    class="form-check-input"
                    :class="{ 'is-invalid': validateRaffleHasAllowedEntries() === false }"
                    :disabled="models.raffle.status !== 'planned'"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="raffleAllowEveryone"
                  >Everyone</label>
                </div>
              </div>
              <div class="list-group-item">
                <div class="d-flex flex-wrap gap-2 align-items-center">
                  <div class="form-check form-switch">
                    <input
                      id="raffleAllowFollower"
                      v-model="models.raffle.allowFollower"
                      class="form-check-input"
                      :class="{ 'is-invalid': validateRaffleHasAllowedEntries() === false }"
                      :disabled="models.raffle.status !== 'planned'"
                      type="checkbox"
                    >
                    <label
                      class="form-check-label"
                      for="raffleAllowFollower"
                    >Followers, since</label>
                  </div>
                  <div class="input-group input-group-sm raffle-inline-input ms-auto">
                    <input
                      v-model="models.raffle.minFollowAge"
                      class="form-control text-end"
                      :class="{ 'is-invalid': validateRaffleIsNumber(models.raffle.minFollowAge) === false }"
                      :disabled="models.raffle.status !== 'planned'"
                      min="0"
                      placeholder="min. Age"
                      style="width: 5rem;"
                      type="number"
                    >
                    <span class="input-group-text">min</span>
                  </div>
                </div>
              </div>
              <div class="list-group-item">
                <div class="form-check form-switch">
                  <input
                    id="raffleAllowSubscriber"
                    v-model="models.raffle.allowSubscriber"
                    class="form-check-input"
                    :class="{ 'is-invalid': validateRaffleHasAllowedEntries() === false }"
                    :disabled="models.raffle.status !== 'planned'"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="raffleAllowSubscriber"
                  >Subscribers</label>
                </div>
              </div>
              <div class="list-group-item">
                <div class="form-check form-switch">
                  <input
                    id="raffleAllowVIP"
                    v-model="models.raffle.allowVIP"
                    class="form-check-input"
                    :class="{ 'is-invalid': validateRaffleHasAllowedEntries() === false }"
                    :disabled="models.raffle.status !== 'planned'"
                    type="checkbox"
                  >
                  <label
                    class="form-check-label"
                    for="raffleAllowVIP"
                  >VIPs</label>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="col">
          <div class="card">
            <div class="card-header">
              Luck Modifiers
            </div>
            <div class="card-body">
              <div class="row">
                <div class="col">
                  <div class="mb-3">
                    <label
                      class="form-label"
                      for="formRaffleMultiFollower"
                    >Followers</label>
                    <input
                      id="formRaffleMultiFollower"
                      v-model="models.raffle.multiFollower"
                      class="form-control form-control-sm"
                      :class="{ 'is-invalid': validateRaffleIsNumber(models.raffle.multiFollower) === false }"
                      :disabled="models.raffle.status !== 'planned'"
                      min="0"
                      step="0.1"
                      type="number"
                    >
                  </div>
                </div>

                <div class="col">
                  <div class="mb-3">
                    <label
                      class="form-label"
                      for="formRaffleMultiSubs"
                    >Subs</label>
                    <input
                      id="formRaffleMultiSubs"
                      v-model="models.raffle.multiSubscriber"
                      class="form-control form-control-sm"
                      :class="{ 'is-invalid': validateRaffleIsNumber(models.raffle.multiSubscriber) === false }"
                      :disabled="models.raffle.status !== 'planned'"
                      min="0"
                      step="0.1"
                      type="number"
                    >
                  </div>
                </div>

                <div class="col">
                  <div class="mb-3">
                    <label
                      class="form-label"
                      for="formRaffleMultiVIP"
                    >VIPs</label>
                    <input
                      id="formRaffleMultiVIP"
                      v-model="models.raffle.multiVIP"
                      class="form-control form-control-sm"
                      :class="{ 'is-invalid': validateRaffleIsNumber(models.raffle.multiVIP) === false }"
                      :disabled="models.raffle.status !== 'planned'"
                      min="0"
                      step="0.1"
                      type="number"
                    >
                  </div>
                </div>
              </div>
              <div class="form-text">
                The base amount of tickets for <strong>Everyone</strong> is always <code>1.0</code>. You can
                increase chances for certain user groups. Groups are checked from right to left: If the user
                is VIP, Sub and Follower they are assigned <strong>VIPkeyword</strong> multiplier.
              </div>
            </div>
          </div>
        </div>

        <div class="col">
          <div class="card">
            <div class="card-header">
              Times
            </div>
            <div class="list-group list-group-flush">
              <div class="list-group-item d-flex flex-wrap gap-2 align-items-center justify-content-between">
                <span class="me-auto">Auto-Start*</span>
                <input
                  v-model="models.raffle.autoStartAt"
                  class="form-control form-control-sm raffle-inline-control"
                  :disabled="models.raffle.status !== 'planned'"
                  :min="transformISOToDateTimeLocal(new Date().toISOString())"
                  type="datetime-local"
                >
              </div>

              <div class="list-group-item d-flex flex-wrap gap-2 align-items-center justify-content-between">
                <span class="me-auto">Duration</span>
                <div class="input-group input-group-sm raffle-inline-input">
                  <input
                    v-model="models.raffle.closeAfter"
                    class="form-control text-end"
                    :class="{ 'is-invalid': validateRaffleIsNumber(models.raffle.closeAfter) === false }"
                    min="0"
                    step="1"
                    style="width: 3rem;"
                    type="number"
                  >
                  <span class="input-group-text">min</span>
                </div>
              </div>

              <div class="list-group-item d-flex flex-wrap gap-2 align-items-center justify-content-between">
                <span class="me-auto">Close At*</span>
                <input
                  v-model="models.raffle.closeAt"
                  class="form-control form-control-sm raffle-inline-control"
                  :min="transformISOToDateTimeLocal(new Date().toISOString())"
                  type="datetime-local"
                >
              </div>

              <div class="list-group-item d-flex flex-wrap gap-2 align-items-center justify-content-between">
                <span class="me-auto">Respond in</span>
                <div class="input-group input-group-sm raffle-inline-input">
                  <input
                    v-model="models.raffle.waitForResponse"
                    class="form-control text-end"
                    :class="{ 'is-invalid': validateRaffleIsNumber(models.raffle.waitForResponse) === false }"
                    min="0"
                    step="1"
                    style="width: 3rem;"
                    type="number"
                  >
                  <span class="input-group-text">sec</span>
                </div>
              </div>

              <div class="list-group-item">
                <div class="form-text">
                  * Optional, when <strong>Close At</strong> is specified, <strong>Duration</strong> is ignored,
                  when no <strong>Auto-Start</strong> is specified the raffle must be started manually. If you
                  set <strong>Respond In</strong> to 0, no chat responses from that user are recorded after
                  picking them as winner.
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="row mt-3">
        <div class="col">
          <div class="card">
            <div class="card-header">
              Texts
            </div>
            <div class="list-group list-group-flush">
              <div class="list-group-item">
                <div class="d-flex align-items-center mb-1">
                  <div class="form-check form-switch">
                    <input
                      id="raffleTextEntryPost"
                      v-model="models.raffle.textEntryPost"
                      class="form-check-input"
                      type="checkbox"
                    >
                    <label
                      class="form-check-label"
                      for="raffleTextEntryPost"
                    >Message on successful entry</label>
                  </div>
                </div>
                <template-editor v-model="models.raffle.textEntry" />
              </div>

              <div class="list-group-item">
                <div class="d-flex align-items-center mb-1">
                  <div class="form-check form-switch">
                    <input
                      id="raffleTextEntryFailPost"
                      v-model="models.raffle.textEntryFailPost"
                      class="form-check-input"
                      type="checkbox"
                    >
                    <label
                      class="form-check-label"
                      for="raffleTextEntryFailPost"
                    >Message on failed entry</label>
                  </div>
                </div>
                <template-editor v-model="models.raffle.textEntryFail" />
              </div>

              <div class="list-group-item">
                <div class="d-flex align-items-center mb-1">
                  <div class="form-check form-switch">
                    <input
                      id="raffleTextWinPost"
                      v-model="models.raffle.textWinPost"
                      class="form-check-input"
                      type="checkbox"
                    >
                    <label
                      class="form-check-label"
                      for="raffleTextWinPost"
                    >Message on winner draw</label>
                  </div>
                </div>
                <template-editor v-model="models.raffle.textWin" />
              </div>

              <div class="list-group-item">
                <div class="d-flex flex-wrap gap-2 align-items-center justify-content-between mb-1">
                  <div class="form-check form-switch">
                    <input
                      id="raffleTextReminderPost"
                      v-model="models.raffle.textReminderPost"
                      class="form-check-input"
                      type="checkbox"
                    >
                    <label
                      class="form-check-label"
                      for="raffleTextReminderPost"
                    >Periodic reminder</label>
                  </div>
                  <div class="input-group input-group-sm raffle-inline-input">
                    <span class="input-group-text">every</span>
                    <input
                      v-model="models.raffle.textReminderInterval"
                      class="form-control text-end"
                      :class="{ 'is-invalid': validateRaffleIsNumber(models.raffle.textReminderInterval) === false }"
                      min="0"
                      step="1"
                      style="width: 5rem;"
                      type="number"
                    >
                    <span class="input-group-text">min</span>
                  </div>
                </div>
                <template-editor v-model="models.raffle.textReminder" />
              </div>

              <div class="list-group-item">
                <div class="d-flex align-items-center mb-1">
                  <div class="form-check form-switch">
                    <input
                      id="raffleTextClosePost"
                      v-model="models.raffle.textClosePost"
                      class="form-check-input"
                      type="checkbox"
                    >
                    <label
                      class="form-check-label"
                      for="raffleTextClosePost"
                    >Message on raffle close</label>
                  </div>
                </div>
                <template-editor v-model="models.raffle.textClose" />
              </div>

              <div class="list-group-item">
                <div class="form-text">
                  Available variables are <code>.user</code> and <code>.raffle</code> with
                  access to all of these configurations most notably <code>.raffle.Title</code>
                  and <code>.raffle.Keyword</code>.
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </AppModal>
  </div>
</template>

<script lang="ts">
import * as constants from '../lib/const'
import { defineComponent, h } from 'vue'
import type { Raffle, RaffleEntry } from '../types'
import { api } from '../api'
import AppModal from '../components/AppModal.vue'
import { confirmDialog } from '../lib/confirmModal'
import TemplateEditor from '../components/tplEditor.vue'
import { useAppStore } from '../stores/app'

const ONE_MINUTE = 60000000000 // nanoseconds
const ONE_SECOND = 1000000000 // nanoseconds

export default defineComponent({
  components: {
    AppModal,
    TemplateEditor,
  },

  computed: {
    raffleEntries() {
      const entries = [...this.openedRaffle.entries || []]
      entries.sort((a, b) => {
        const isWinner = (e: RaffleEntry) => e.wasPicked && !e.wasRedrawn
        const wasWinnner = (e: RaffleEntry) => e.wasPicked && e.wasRedrawn

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
        const scores: Record<string, number> = { active: 0, ended: 2, planned: 1 }

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
      appStore: useAppStore(),
      models: { raffle: {} as Raffle },
      now: new Date(),
      openedRaffle: {} as Raffle,
      openedRaffleReloading: false,
      raffles: [] as Raffle[],
      reopenRaffleDuration: 60,
      showRaffleEditModal: false,
      showRaffleEntriesModal: false,
    }
  },

  methods: {
    async cloneRaffle(id: number | string) {
      try {
        await api.put(`raffle/${id}/clone`, {})
        return this.appStore.toastSuccess('Raffle cloned')
      } catch (err) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    closeEntriesModal() {
      this.openedRaffle = {} as Raffle
      this.showRaffleEntriesModal = false
    },

    closeRaffle(id: number | string) {
      confirmDialog('Do you really want to close entries for this raffle?', {
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

          return api.put(`raffle/${id}/close`, {})
            .then(() => this.appStore.toastSuccess('Raffle closed'))
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    deleteRaffle(id: number | string) {
      confirmDialog('Do you really want to delete this raffle and all its entries?', {
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

          return api.delete(`raffle/${id}`)
            .then(() => this.appStore.toastSuccess('Raffle deleted'))
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    editRaffle(raffle: Raffle) {
      this.models.raffle = raffle
      this.showRaffleEditModal = true
    },

    async fetchRaffles() {
      try {
        const resp = await api.get<Raffle[]>('raffle/')
        this.raffles = (resp || []).map(raffle => this.transformRaffleFromDB(raffle))
      } catch (err) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    newRaffle() {
      this.models.raffle = {
        /* eslint-disable sort-keys */
        id: 0,

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

        textClose: 'The raffle "{{ .raffle.Title }}" is closed now, you can no longer enter!',
        textClosePost: false,

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
      }
      this.showRaffleEditModal = true
    },

    async pickWinner() {
      this.openedRaffleReloading = true
      try {
        await api.put(`raffle/${this.openedRaffle.id}/pick`, {})
        return this.appStore.toastSuccess('Winner picked!')
      } catch {
        return this.appStore.toastError('Could not pick winner!')
      }
    },

    raffleTimer(raffle: Raffle) {
      const parts = []
      let tte = new Date(raffle.closeAt!).getTime() - this.now.getTime()

      for (const d of [3600000, 60000, 1000]) {
        const pt = Math.floor(tte / d)
        parts.push(String(pt).padStart(2, '0'))
        tte -= pt * d
      }

      return parts.join(':')
    },

    async refreshOpenendRaffle() {
      this.openedRaffleReloading = true
      const resp = await api.get<Raffle>(`raffle/${this.openedRaffle.id}`)
      this.openedRaffle = (resp ? this.transformRaffleFromDB(resp) : {}) as Raffle
      this.openedRaffleReloading = false
    },

    reopenRaffle(raffleId: number | string) {
      let duration = 10

      const content = h('div', {}, [
        h('div', { class: 'input-group input-group-sm mb-2' }, [
          h('input', {
            class: 'form-control text-end',
            min: '0',
            onInput: (event: Event) => {
              duration = Number((event.target as HTMLInputElement).value)
            },
            step: '1',
            type: 'number',
            value: String(duration),
          }),
          h('span', { class: 'input-group-text' }, 'min'),
        ]),
        h('div', { class: 'form-text' }, 'The raffle will be re-opened and the "Close At" attribute will be set to the given duration from now.'),
        h('div', { class: 'form-text' }, [
          'This will ',
          h('strong', 'NOT'),
          ' clear the entrants, so do not use this for another raffle, use the "Duplicate Raffle" functionality for that.',
        ]),
      ])

      confirmDialog(content, {
        buttonSize: 'sm',
        centered: true,
        size: 'sm',
        title: 'Re-Open the Raffle?',
      })
        .then(val => {
          if (!val) {
            return
          }

          return api.put(`raffle/${raffleId}/reopen?duration=${duration * 60}`, {})
            .then(() => this.appStore.toastSuccess('Raffle re-opened'))
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    async repickWinner(winnerId: number | string) {
      this.openedRaffleReloading = true
      try {
        await api.put(`raffle/${this.openedRaffle.id}/repick/${winnerId}`, {})
        return this.appStore.toastSuccess('Winner re-picked!')
      } catch {
        return this.appStore.toastError('Could not re-pick winner!')
      }
    },

    resetRaffle(id: number | string) {
      confirmDialog('Do you really want to reset this raffle?', {
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

          return api.put(`raffle/${id}/reset`, {})
            .then(() => this.appStore.toastSuccess('Raffle reset'))
            .catch(err => this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err))
        })
    },

    async saveRaffle() {
      if (this.models.raffle.id) {
        try {
          await api.put(`raffle/${this.models.raffle.id}`, this.transformRaffleToDB(this.models.raffle))
          return this.appStore.toastSuccess('Raffle updated')
        } catch (err) {
          return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
        }
      }

      try {
        await api.post('raffle/', this.transformRaffleToDB(this.models.raffle))
        return this.appStore.toastSuccess('Raffle created')
      } catch (err_1) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err_1)
      }
    },

    async showEntryDialog(raffleId: number) {
      this.openedRaffle = { id: raffleId } as Raffle
      await this.refreshOpenendRaffle()
      this.showRaffleEntriesModal = true
    },

    async startRaffle(id: number | string) {
      try {
        await api.put(`raffle/${id}/start`, {})
        return this.appStore.toastSuccess('Raffle started')
      } catch (err) {
        return this.$bus.$emit(constants.NOTIFY_FETCH_ERROR, err)
      }
    },

    transformISOToDateTimeLocal(t: string) {
      const d = new Date(t)
      d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
      return d.toISOString().slice(0, 16)
    },

    transformRaffleFromDB(raffle: Raffle): Raffle {
      return {
        /* eslint-disable sort-keys */
        ...raffle,

        channel: raffle.channel.replace(/^#/, ''),

        id: raffle.id ? Number(raffle.id) : 0,

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

    transformRaffleToDB(raffle: Raffle): Raffle {
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
      ] as const) {
        // You must not put text in numeric fields
        if (this.validateRaffleIsNumber(this.models.raffle[nf]) === false) {
          return false
        }
      }

      if (this.validateRaffleChannel() === false) {
        return false
      }

      for (const nef of ['keyword', 'title'] as const) {
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
      if (!constants.REGEXP_USER.test(this.models.raffle.channel)) {
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

    validateRaffleIsNumber(n: any) {
      if (isNaN(Number(n))) {
        return false
      }
      return null
    },

    validateRaffleNonEmpty(str: string) {
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

  name: 'TwitchBotRaffleView',
})
</script>

<style scoped>
.raffle-inline-control {
  flex: 0 1 fit-content;
  max-width: 100%;
  min-width: 13rem;
}

.raffle-inline-input {
  flex: 0 1 fit-content;
  max-width: 100%;
  min-width: 7rem;
}
</style>
