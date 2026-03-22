<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="modal fade show d-block"
      tabindex="-1"
      @click.self="close"
    >
      <div
        class="modal-dialog"
        :class="dialogClasses"
      >
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ title }}
            </h5>
            <button
              type="button"
              class="btn-close"
              aria-label="Close"
              @click="close"
            />
          </div>

          <div class="modal-body">
            <slot />
          </div>

          <div
            v-if="!hideFooter"
            class="modal-footer"
          >
            <slot name="footer">
              <button
                type="button"
                class="btn btn-secondary"
                @click="close"
              >
                Cancel
              </button>
              <button
                type="button"
                class="btn btn-primary"
                :disabled="okDisabled"
                @click="handleOk"
              >
                {{ okTitle }}
              </button>
            </slot>
          </div>
        </div>
      </div>
    </div>
    <div
      v-if="modelValue"
      class="modal-backdrop fade show"
    />
  </Teleport>
</template>

<script lang="ts">
import { defineComponent, watch } from 'vue'

export default defineComponent({
  computed: {
    dialogClasses() {
      return {
        'modal-dialog-centered': this.centered,
        'modal-dialog-scrollable': this.scrollable,
        'modal-lg': this.size === 'lg',
        'modal-md': this.size === 'md',
        'modal-sm': this.size === 'sm',
        'modal-xl': this.size === 'xl',
      }
    },
  },

  emits: ['hidden', 'ok', 'update:modelValue'],

  methods: {
    close() {
      this.$emit('update:modelValue', false)
      this.$emit('hidden')
    },

    handleOk() {
      const evt = {
        defaultPrevented: false,
        preventDefault() {
          this.defaultPrevented = true
        },
      }

      this.$emit('ok', evt)
      if (!evt.defaultPrevented) {
        this.close()
      }
    },
  },

  mounted() {
    document.body.classList.add('modal-open')
  },

  name: 'TwitchBotAppModal',

  props: {
    centered: {
      default: false,
      type: Boolean,
    },

    hideFooter: {
      default: false,
      type: Boolean,
    },

    modelValue: {
      required: true,
      type: Boolean,
    },

    okDisabled: {
      default: false,
      type: Boolean,
    },

    okTitle: {
      default: 'Save',
      type: String,
    },

    scrollable: {
      default: false,
      type: Boolean,
    },

    size: {
      default: 'md',
      type: String,
    },

    title: {
      default: '',
      type: String,
    },
  },

  setup(props) {
    watch(() => props.modelValue, value => {
      document.body.classList.toggle('modal-open', value)
    }, { immediate: true })

    return {}
  },

  unmounted() {
    document.body.classList.remove('modal-open')
  },
})
</script>
