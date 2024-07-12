/**
 * Allows to add filters for custom events created through the customHandler
 *
 * @returns {Object} Custom filter definitions as `filterKey: {name: "Name", visible: true}`
 */
const customFilters = () => ({})

/**
 * Handles custom events and creates feed items from them
 *
 * @param {*} param0 Event-Object as returned by the websocket
 * @returns {Object} Event to add to the event list of the feed
 */
const customHandler = eventObj => {
  console.log('custom event unhandled:', eventObj)
  return null
}

export { customFilters, customHandler }
