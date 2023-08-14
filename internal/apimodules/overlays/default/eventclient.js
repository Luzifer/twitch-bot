/**
 * Options to pass to the EventClient constructor
 * @typedef {Object} Options
 * @prop {string} [channel] - Filter for specific channel events (format: `#channel`)
 * @prop {Object} [handlers={}] - Map event types to callback functions `(event, fields, time, live) => {...}`
 * @prop {number} [maxReplayAge=-1] - Number of hours to replay the events for (-1 = infinite)
 * @prop {boolean} [replay=false] - Request a replay at connect (requires channel to be set to a channel name)
 * @prop {string} [token] - API access token to use to connect to the WebSocket (if not set, must be provided through URL hash)
 */

const HOUR = 3600 * 1000

const initialSocketBackoff = 500
const maxSocketBackoff = 10000
const socketBackoffMultiplier = 1.25

/**
 * @class EventClient abstracts the connection to the bot websocket for events
 */
class EventClient {
  /**
   * Creates, initializes and connects the EventClient
   *
   * @param {Options} opts Options for the EventClient
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

    // If reply is enabled and channel is provided, fetch the replay
    if (this.paramOptionFallback('replay', false) && this.paramOptionFallback('channel')) {
      this.fetchReplayForChannel(
        this.paramOptionFallback('channel'),
        Number(this.paramOptionFallback('maxReplayAge', -1)),
      )
    }
  }

  /**
   * Returns the API base URL without trailing slash
   *
   * @returns {string} API base URL
   */
  apiBase() {
    return window.location.href.substr(0, window.location.href.indexOf('/overlays/'))
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

      if (data.type === '_auth') {
        // Special handling for auth confirmation
        this.socketBackoff = initialSocketBackoff
        return
      }

      if (this.paramOptionFallback('channel') && !data.fields?.channel?.match(this.paramOptionFallback('channel'))) {
        // Channel filter is active and channel does not match
        return
      }

      for (const fn of [this.handlers[data.type], this.handlers._].filter(fn => fn)) {
        fn(data.type, data.fields, new Date(data.time), data.is_live)
      }
    }

    this.socket.onopen = () => {
      this.socket.send(JSON.stringify({
        fields: { token: this.token },
        type: '_auth',
      }))
    }
  }

  /**
   * Requests past events from the API and feed them through the registered handlers
   *
   * @param {string} channel The channel to fetch the events for
   * @param {number} hours The amount of hours to fetch into the past (-1 = infinite)
   * @private
   * @returns {Promise} Can be listened for failures using `.catch`
   */
  fetchReplayForChannel(channel, hours = -1) {
    const params = new URLSearchParams()
    if (hours > -1) {
      params.set('since', new Date(new Date().getTime() - hours * HOUR).toISOString())
    }

    return fetch(`${this.apiBase()}/overlays/events/${encodeURIComponent(channel)}?${params.toString()}`, {
      headers: {
        authorization: this.paramOptionFallback('token'),
      },
    })
      .then(resp => resp.json())
      .then(data => {
        const handlers = []

        for (const msg of data) {
          for (const fn of [this.handlers[msg.type], this.handlers._].filter(fn => fn)) {
            handlers.push(fn(msg.type, msg.fields, new Date(msg.time), msg.is_live))
          }
        }

        return Promise.all(handlers)
      })
  }

  /**
   * Resolves the given key through url hash parameters with fallback to constructor options
   *
   * @param {string} key The key to resolve
   * @param {*} [fallback=null] Fallback to return if neither params nor options contained that key
   * @returns {*} Value of the key or `null`
   */
  paramOptionFallback(key, fallback = null) {
    return this.params.get(key) || this.options[key] || fallback
  }

  /**
   * Renders a given template using the bots msgformat API (supports all templating you can use in bot messages). To use this function the token passed through the constructor or the URL hash must have the `msgformat` permission in addition to the `overlays` permission.
   *
   * @param {string} template The template to render
   * @returns {Promise} Promise resolving to the rendered output of the template
   */
  renderTemplate(template) {
    return fetch(`${this.apiBase()}/msgformat/format?template=${encodeURIComponent(template)}`, {
      headers: {
        authorization: this.paramOptionFallback('token'),
      },
    })
      .then(resp => resp.text())
  }

  /**
   * Modifies the overlay address to the websocket address the bot listens to
   *
   * @private
   * @returns {string} Websocket address in form ws://... or wss://...
   */
  socketAddr() {
    return `${this.apiBase().replace(/^http/, 'ws')}/overlays/events.sock`
  }
}

export default EventClient
