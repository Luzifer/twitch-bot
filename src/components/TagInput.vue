<template>
  <div>
    <div
      class="form-control d-flex flex-wrap gap-2 align-items-center"
      :class="stateClass"
    >
      <span
        v-for="tag in modelValue"
        :key="tag"
        class="badge bg-secondary-subtle text-secondary-emphasis d-inline-flex align-items-center gap-2"
      >
        {{ tag }}
        <button
          type="button"
          class="btn-close btn-close-white"
          aria-label="Remove"
          @click="removeTag(tag)"
        />
      </span>
      <input
        :id="id"
        v-model="draft"
        class="border-0 flex-grow-1"
        :placeholder="placeholder"
        @blur="flushDraft"
        @keydown="handleKeydown"
        @paste="handlePaste"
      >
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

function defaultValidator(value: string) {
  return value.trim().length > 0
}

export default defineComponent({
  computed: {
    stateClass() {
      return {
        'is-invalid': this.state === false,
        'is-valid': this.state === true,
      }
    },
  },

  data() {
    return {
      draft: '',
    }
  },

  emits: ['update:modelValue'],

  methods: {
    addTag(value: string) {
      const tag = value.trim()
      if (!tag || !this.validator(tag) || this.modelValue.includes(tag)) {
        return
      }

      this.$emit('update:modelValue', [...this.modelValue, tag])
    },

    flushDraft() {
      const parts = this.draft.split(/[ ,]+/).filter(Boolean)
      for (const part of parts) {
        this.addTag(part)
      }
      this.draft = ''
    },

    handleKeydown(evt: KeyboardEvent) {
      if (![' ', ',', 'Enter'].includes(evt.key)) {
        return
      }

      evt.preventDefault()
      this.flushDraft()
    },

    handlePaste(evt: ClipboardEvent) {
      const text = evt.clipboardData?.getData('text') || ''
      if (!/[ ,]/.test(text)) {
        return
      }

      evt.preventDefault()
      for (const part of text.split(/[ ,]+/).filter(Boolean)) {
        this.addTag(part)
      }
      this.draft = ''
    },

    removeTag(tag: string) {
      this.$emit('update:modelValue', this.modelValue.filter((entry: string) => entry !== tag))
    },
  },

  name: 'TwitchBotTagInput',

  props: {
    id: {
      default: '',
      type: String,
    },

    modelValue: {
      default: () => [],
      type: Array<string>,
    },

    placeholder: {
      default: '',
      type: String,
    },

    state: {
      default: null,
      type: Boolean,
    },

    validator: {
      default: defaultValidator,
      type: Function,
    },
  },
})
</script>

<style scoped>
input {
  background: transparent;
  min-width: 8rem;
  outline: none;
}
</style>
