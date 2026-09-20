<overlay-title>Debug</overlay-title>

<template>
  <table>
    <thead>
      <tr>
        <th>Time</th>
        <th>Reason</th>
        <th>Event</th>
        <th>Fields</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="event in events"
        :key="event.eventKey"
      >
        <td>{{ formatTime(event.time) }}</td>
        <td>{{ event.reason }}</td>
        <td>{{ event.event }}</td>
        <td>
          <span
            v-for="field in formatFields(event.fields)"
            :key="field"
            class="event"
          >{{ field }}</span>
        </td>
      </tr>
    </tbody>
  </table>
</template>

<script lang="ts">
import { createApp, defineComponent } from 'vue'
import EventClient from './eventclient.js'
import type { EventSocketMessage } from './eventTypes.js'

type DebugEvent = {
  event: string
  eventKey: number
  fields: EventSocketMessage['fields']
  reason: EventSocketMessage['reason']
  time: Date
}

const component = defineComponent({
  created() {
    const eventClient = new EventClient({
      handlers: {
        _: ({ fields, reason, time, type }) => {
          if ((eventClient.paramOptionFallback('hide', '') as string)
            .split(',')
            .includes(type)) {
            return
          }

          this.events = [
            {
              event: type,
              eventKey: (this.events[0]?.eventKey ?? 0) + 1,
              fields,
              reason,
              time,
            },
            ...this.events,
          ]
        },
      },

      replay: true,
    })

    ;(window as typeof window & { botClient: EventClient }).botClient = eventClient
  },

  data() {
    return {
      events: [] as DebugEvent[],
    }
  },

  methods: {
    formatFields(fields: EventSocketMessage['fields']) {
      return Object.entries(fields || {})
        .map(([key, value]) => `${key}="${String(value)}"`)
        .sort()
    },

    formatTime(time: Date) {
      const date = time.toLocaleDateString('sv-SE')
      const clock = time.toLocaleTimeString('sv-SE', { hour12: false })
      return `${date} ${clock}`
    },
  },

  name: 'DebugOverlay',
})

queueMicrotask(() => createApp(component).mount('#app'))

export default component
</script>

<style>
html {
  background-color: #333;
  color: #fff;
  font-family: monospace;
}

span.event {
  background-color: #e3e3ff3f;
  border-radius: 0.5rem;
  display: inline-block;
  margin-bottom: 0.5rem;
  margin-right: 5px;
  padding: 0.1rem 0.5rem;
  white-space: pre;
}

table {
  border-spacing: 10px;
  margin: 0 auto;
  max-width: 1200px;
}

td {
  vertical-align: top;
}

th {
  text-align: left;
}
</style>
