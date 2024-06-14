import { type ToastContent } from '../components/_toast.vue'

/**
 * Create the content of an error-toast
 *
 * @param text The message to display to the user
 * @returns The {ToastContent} for usage in `this.bus.emit(BusEventTypes.Toast, errorToast(...))`
 */
const errorToast = (text: string): ToastContent => ({
  autoHide: false,
  color: 'danger',
  id: crypto.randomUUID(),
  text,
})

/**
 * Create the content of an info-toast
 *
 * @param text The message to display to the user
 * @returns The {ToastContent} for usage in `this.bus.emit(BusEventTypes.Toast, infoToast(...))`
 */
const infoToast = (text: string): ToastContent => ({
  color: 'info',
  id: crypto.randomUUID(),
  text,
})

/**
 * Create the content of an success-toast
 *
 * @param text The message to display to the user
 * @returns The {ToastContent} for usage in `this.bus.emit(BusEventTypes.Toast, successToast(...))`
 */
const successToast = (text: string): ToastContent => ({
  color: 'success',
  id: crypto.randomUUID(),
  text,
})

export { errorToast, infoToast, successToast }
