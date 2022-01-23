/*
 *  {
 *    handlers: {
 *      join: (event, data) => { ... },
 *    },
 *    token: '...',
 *  }
 */

const initialSocketBackoff = 500
const maxSocketBackoff = 10000
const socketBackoffMultiplier = 1.25

export default class EventClient {
  constructor(opts) {
    this.params = new URLSearchParams(window.location.hash.substr(1))
    this.handlers = { ...opts.handlers || {} }

    this.token = this.params.get('token') || opts.token || null
    if (!this.token) {
      throw new Error('token for socket not present in hash or opts')
    }

    this.socketBackoff = initialSocketBackoff

    this.connect()
  }

  connect() {
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }

    this.socket = new WebSocket(this.socketAddr())

    this.socket.onclose = () => {
      this.socketBackoff = Math.min(this.socketBackoff * socketBackoffMultiplier, maxSocketBackoff)
      window.setTimeout(() => this.connect(), this.socketBackoff)
    }

    this.socket.onmessage = evt => {
      const data = JSON.parse(evt.data)

      console.log(data)

      for (const fn of [this.handlers[data.type], this.handlers._].filter(fn => fn)) {
        fn(data.type, data.fields)
      }
    }

    this.socket.onopen = () => {
      this.socketBackoff = initialSocketBackoff
    }
  }

  socketAddr() {
    const base = window.location.href.substr(0, window.location.href.indexOf('/overlays/') + '/overlays/'.length)
    return `${base.replace(/^http/, 'ws')}events.sock`
  }
}
