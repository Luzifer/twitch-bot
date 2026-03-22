type EventHandler = (payload?: any) => void

class EventBus {
  private handlers = new Map<string, Set<EventHandler>>()

  emit(event: string, payload?: any) {
    for (const handler of this.handlers.get(event) || []) {
      handler(payload)
    }
  }

  $emit(event: string, payload?: any) {
    this.emit(event, payload)
  }

  off(event: string, handler: EventHandler) {
    this.handlers.get(event)?.delete(handler)
  }

  $off(event: string, handler: EventHandler) {
    this.off(event, handler)
  }

  on(event: string, handler: EventHandler) {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, new Set())
    }

    this.handlers.get(event)?.add(handler)

    return () => this.off(event, handler)
  }

  $on(event: string, handler: EventHandler) {
    return this.on(event, handler)
  }
}

export const bus = new EventBus()
