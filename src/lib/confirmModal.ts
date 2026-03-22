import { reactive, type VNode } from 'vue'

type ConfirmContent = string | VNode | VNode[] | null

type ConfirmOptions = {
  buttonSize?: 'sm' | 'md' | 'lg'
  cancelTitle?: string
  centered?: boolean
  okTitle?: string
  okVariant?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  title?: string
}

export const confirmModalState = reactive({
  buttonSize: 'sm' as 'sm' | 'md' | 'lg',
  cancelTitle: 'Cancel',
  centered: true,
  content: null as ConfirmContent,
  okTitle: 'OK',
  okVariant: 'primary',
  resolver: null as null | ((value: boolean) => void),
  size: 'sm' as 'sm' | 'md' | 'lg' | 'xl',
  title: 'Please Confirm',
  visible: false,
})

function settleConfirm(value: boolean) {
  const resolver = confirmModalState.resolver

  confirmModalState.resolver = null
  confirmModalState.visible = false
  confirmModalState.content = null

  resolver?.(value)
}

export function confirmDialog(content: ConfirmContent, options: ConfirmOptions = {}) {
  if (confirmModalState.resolver) {
    settleConfirm(false)
  }

  confirmModalState.buttonSize = options.buttonSize || 'sm'
  confirmModalState.cancelTitle = options.cancelTitle || 'Cancel'
  confirmModalState.centered = options.centered !== false
  confirmModalState.content = content
  confirmModalState.okTitle = options.okTitle || 'OK'
  confirmModalState.okVariant = options.okVariant || 'primary'
  confirmModalState.size = options.size || 'sm'
  confirmModalState.title = options.title || 'Please Confirm'
  confirmModalState.visible = true

  return new Promise<boolean>(resolve => {
    confirmModalState.resolver = resolve
  })
}

export { settleConfirm }
