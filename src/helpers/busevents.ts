/* eslint-disable no-unused-vars */

enum BusEventTypes {
  ChangePending = 'changePending',
  ConfigReload = 'configReload',
  Error = 'error',
  FetchError = 'fetchError',
  LoadingData = 'loadingData',
  LoginProcessing = 'loginProcessing',
  NotifySocketConnected = 'notifySocketConnected',
  NotifySocketDisconnected = 'notifySocketDisconnected',
  RaffleChanged = 'raffleChanged',
  RaffleEntryChanged = 'raffleEntryChanged',
  Toast = 'toast',
}

export default BusEventTypes
