import type { EventHandlers, EventSocketMessage } from './eventTypes.js'

const HOUR = 3600 * 1000
const initialSocketBackoff = 500
const maxSocketBackoff = 10000
const socketBackoffMultiplier = 1.25

type LegacyHandler = (type: string, fields: EventSocketMessage['fields'], time: Date, isLive: boolean) => unknown

type InternalOptions = {
  channel?: string
  handlers?: EventHandlers
  maxReplayAge?: number
  replay?: boolean
  token?: string
}

type Options = InternalOptions & Record<string, unknown>

/** EventClient abstracts the connection to the bot websocket for events. */
export default class EventClient {
  private handlers: EventHandlers

  private options: Options

  private params: URLSearchParams

  private socket: WebSocket | null = null

  private socketBackoff: number

  private token: string

  /**
   * Creates, initializes and connects the EventClient.
   *
   * @param opts EventClient options
   */
  constructor(opts: Options = {}) {
    this.params = new URLSearchParams(window.location.hash.substring(1))
    this.handlers = { ...opts.handlers || {} }
    this.options = { ...opts }
    this.token = this.paramOptionFallback('token', '')

    if (!this.token) {
      throw new Error('token for socket not present in hash or opts')
    }

    this.socketBackoff = initialSocketBackoff
    this.connect()

    if (this.paramOptionFallback('replay', false) && this.paramOptionFallback('channel', '')) {
      this.fetchReplayForChannel(
        this.paramOptionFallback('channel', ''),
        Number(this.paramOptionFallback('maxReplayAge', -1)),
      )
    }
  }

  /**
   * Returns the API base URL without trailing slash.
   *
   * @returns API base URL
   */
  apiBase(): string {
    return window.location.href.substring(0, window.location.href.indexOf('/overlays/'))
  }

  paramOptionFallback<K extends keyof InternalOptions>(
    key: K,
    fallback: NonNullable<InternalOptions[K]>,
  ): NonNullable<InternalOptions[K]>

  paramOptionFallback<K extends keyof InternalOptions>(
    key: K,
  ): InternalOptions[K] | null

  paramOptionFallback(key: string, fallback?: unknown): unknown

  /**
   * Resolves a URL hash parameter with a fallback to the constructor options.
   *
   * @param key Option key to resolve
   * @param fallback Value returned when the option is absent
   * @returns Resolved option value
   */
  paramOptionFallback(key: string, fallback: unknown = null): unknown {
    return this.params.get(key) || this.options[key] || fallback
  }

  /**
   * Renders a template using the bot's msgformat API.
   *
   * The token requires the `msgformat` permission in addition to `overlays`.
   *
   * @param template Template to render
   * @returns Rendered template
   */
  async renderTemplate(template: string): Promise<string> {
    const resp = await fetch(`${this.apiBase()}/msgformat/format?template=${encodeURIComponent(template)}`, {
      headers: { authorization: this.token },
    })
    return await resp.text()
  }

  /**
   * Triggers a replay of an event for all connected overlays.
   *
   * @param eventId Event ID received in a socket message
   * @returns Response from the replay request
   */
  replayEvent(eventId: string): Promise<Response> {
    return fetch(`${this.apiBase()}/overlays/event/${eventId}/replay`, {
      headers: { authorization: this.token },
      method: 'PUT',
    })
  }

  private connect(): void {
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
      const data = this.parseMessage(evt.data as string)

      if (data.type === '_auth') {
        this.socketBackoff = initialSocketBackoff
        return
      }

      const channel = this.paramOptionFallback('channel', '')
      if (channel && !data.fields?.channel?.match(channel)) {
        return
      }

      this.runHandlers(data)
    }

    this.socket.onopen = () => {
      this.socket?.send(JSON.stringify({ fields: { token: this.token }, type: '_auth' }))
    }
  }

  private async fetchReplayForChannel(channel: string, hours = -1): Promise<unknown[]> {
    const params = new URLSearchParams()
    if (hours > -1) {
      params.set('since', new Date(new Date().getTime() - hours * HOUR).toISOString())
    }

    const resp = await fetch(`${this.apiBase()}/overlays/events/${encodeURIComponent(channel)}?${params.toString()}`, {
      headers: { authorization: this.token },
    })
    const data = await (resp.json() as Promise<EventSocketMessage[]>)
    return await Promise.all(data.flatMap(msg => this.runHandlers(msg)))
  }

  private parseMessage(data: string): EventSocketMessage | { fields: { channel?: string }, type: '_auth' } {
    return JSON.parse(data) as EventSocketMessage | { fields: { channel?: string }, type: '_auth' }
  }

  private runHandlers(data: EventSocketMessage): unknown[] {
    const handler = this.handlers[data.type] as ((event: EventSocketMessage) => unknown) | LegacyHandler | undefined
    const fallback = this.handlers._
    const normalized = {
      ...data,
      time: new Date(data.time as unknown as string),
    } as EventSocketMessage

    return [handler, fallback]
      .filter((fn): fn is ((event: EventSocketMessage) => unknown) | LegacyHandler => Boolean(fn))
      .map(fn => fn.length === 1
        ? (fn as (event: EventSocketMessage) => unknown)(normalized)
        : (fn as LegacyHandler)(normalized.type, normalized.fields, normalized.time, normalized.is_live))
  }

  private socketAddr(): string {
    return `${this.apiBase().replace(/^http/, 'ws')}/overlays/events.sock`
  }
}
