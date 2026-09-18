import EventClient from './eventclient.js'

const client = new EventClient({
  handlers: {
    _: event => {
      if (event.type === 'alpha') {
        event.fields.duration.toFixed()
      }
    },
    alpha: event => {
      event.fields.duration.toFixed()
      // @ts-expect-error alpha events do not have subscription plans
      event.fields.plan
    },
  },
  token: 'token',
})

client.apiBase().toUpperCase()
client.paramOptionFallback('channel')
client.renderTemplate('template').then(output => output.toUpperCase())
client.replayEvent('1').then(response => response.text())
