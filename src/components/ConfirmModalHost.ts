import { defineComponent, h } from 'vue'

import { confirmModalState, settleConfirm } from '../lib/confirmModal'
import AppModal from './AppModal.vue'

function renderConfirmContent() {
  if (!confirmModalState.content) {
    return null
  }

  if (typeof confirmModalState.content === 'string') {
    return h('p', { class: 'mb-0' }, confirmModalState.content)
  }

  return Array.isArray(confirmModalState.content)
    ? confirmModalState.content
    : [confirmModalState.content]
}

export default defineComponent({
  methods: {
    cancel() {
      settleConfirm(false)
    },

    confirm() {
      settleConfirm(true)
    },

    handleHidden() {
      if (confirmModalState.resolver) {
        settleConfirm(false)
      }
    },

    handleModelValue(value: boolean) {
      if (!value && confirmModalState.resolver) {
        settleConfirm(false)
      }
    },
  },

  name: 'TwitchBotConfirmModalHost',

  render() {
    return h(AppModal, {
      centered: confirmModalState.centered,
      modelValue: confirmModalState.visible,
      onHidden: this.handleHidden,
      'onUpdate:modelValue': this.handleModelValue,
      size: confirmModalState.size,
      title: confirmModalState.title,
    }, {
      default: () => renderConfirmContent(),
      footer: () => [
        h('button', {
          class: ['btn', 'btn-secondary', confirmModalState.buttonSize === 'md' ? null : `btn-${confirmModalState.buttonSize}`],
          onClick: this.cancel,
          type: 'button',
        }, confirmModalState.cancelTitle),
        h('button', {
          class: ['btn', `btn-${confirmModalState.okVariant}`, confirmModalState.buttonSize === 'md' ? null : `btn-${confirmModalState.buttonSize}`],
          onClick: this.confirm,
          type: 'button',
        }, confirmModalState.okTitle),
      ],
    })
  },
})
