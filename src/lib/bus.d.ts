import type { bus } from './eventBus'

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $bus: typeof bus
  }
}

export { }
