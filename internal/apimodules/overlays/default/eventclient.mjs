/**
 * Options to pass to the EventClient constructor
 * @typedef {Object} EventClient~Options
 * @prop {string} [channel] - Filter for specific channel events (matched as `event.channel.match(...)`)
 * @prop {Object} handlers - Map event types to callback functions `(event, fields) => {...}`
 * @prop {string} [token] - API access token to use to connect to the WebSocket
 */

const initialSocketBackoff = 500
const maxSocketBackoff = 10000
const socketBackoffMultiplier = 1.25

/**
 * @class EventClient abstracts the connection to the bot websocket for events
 */
export default class EventClient {
  /**
   * Creates, initializes and connects the EventClient
   *
   * @param {EventClient~Options} opts {@link EventClient~Options} for the EventClient
   */
  constructor(opts) {
    this.params = new URLSearchParams(window.location.hash.substr(1))
    this.handlers = { ...opts.handlers || {} }
    this.options = { ...opts }

    this.token = this.paramOptionFallback('token')
    if (!this.token) {
      throw new Error('token for socket not present in hash or opts')
    }

    this.socketBackoff = initialSocketBackoff

    this.connect()
  }

  /**
   * Connects the EventClient to the socket
   *
   * @private
   */
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

      if (this.paramOptionFallback('channel') && !data.fields?.channel?.match(this.paramOptionFallback('channel'))) {
        // Channel filter is active and channel does not match
        return
      }

      for (const fn of [this.handlers[data.type], this.handlers._].filter(fn => fn)) {
        fn(data.type, data.fields)
      }
    }

    this.socket.onopen = () => {
      this.socketBackoff = initialSocketBackoff
    }
  }

  /**
   * Resolves the given key through url hash parameters with fallback to constructor options
   *
   * @params {string} key The key to resolve
   * @returns {*} Value of the key or null
   */
  paramOptionFallback(key) {
    return this.params.get(key) || this.options[key] || null
  }

  /**
   * Modifies the overlay address to the websocket address the bot listens to
   *
   * @private
   * @returns {string} Websocket address in form ws://... or wss://...
   */
  socketAddr() {
    const base = window.location.href.substr(0, window.location.href.indexOf('/overlays/') + '/overlays/'.length)
    return `${base.replace(/^http/, 'ws')}events.sock`
  }
}
