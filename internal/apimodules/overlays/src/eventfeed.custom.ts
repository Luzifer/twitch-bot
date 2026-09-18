import { CustomSocketMessage } from './eventTypes'

type Event = {
  eventId: string
  extraData?: Record<string, any>
  filterKey: string
  hasReplay?: boolean
  icon: string
  isMeta?: boolean
  originId?: string
  subtext?: string | (() => string | undefined)
  text?: string
  time: Date
  title: string
}

type Filter = {
  name: string
  visible: boolean
}

/**
 * Allows to add filters for custom events created through the customHandler
 *
 * @returns Custom filter definitions as `filterKey: {name: "Name", visible: true}`
 */
export const customFilters = (): Record<string, Filter> => ({})

/**
 * Handles custom events and creates feed items from them
 *
 * @param eventObj Event-Object as returned by the websocket
 * @returns Event to add to the event list of the feed
 */
export const customHandler = (eventObj: CustomSocketMessage): Event | null => {
  console.log('custom event unhandled:', eventObj)
  return null
}
