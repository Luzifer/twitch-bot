import BusEventTypes from './busevents'

class ConfigNotifyListener {
  private backoff: number = 100

  private eventListener: Function

  private socket: WebSocket | null = null

  constructor(listener: Function) {
    this.eventListener = listener
    this.connect()
  }

  private connect(): void {
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }

    const baseURL = window.location.href.split('#')[0].replace(/^http/, 'ws')
    this.socket = new WebSocket(`${baseURL}config-editor/notify-config`)

    this.socket.onopen = () => {
      console.debug('[notify] Socket connected')
      this.eventListener(BusEventTypes.NotifySocketConnected)
    }

    this.socket.onmessage = evt => {
      const msg = JSON.parse(evt.data)

      console.debug(`[notify] Socket message received type=${msg.msg_type}`)
      this.backoff = 100 // We've received a message, reset backoff

      if (msg.msg_type !== 'ping') {
        this.eventListener(msg.msg_type)
      }
    }

    this.socket.onclose = evt => {
      console.debug(`[notify] Socket was closed wasClean=${evt.wasClean}`)
      this.eventListener(BusEventTypes.NotifySocketDisconnected)
      this.updateBackoffAndReconnect()
    }
  }

  private updateBackoffAndReconnect(): void {
    this.backoff = Math.min(this.backoff * 1.5, 10000)
    window.setTimeout(() => this.connect(), this.backoff)
  }
}

export default ConfigNotifyListener