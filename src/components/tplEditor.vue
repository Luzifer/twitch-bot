<template>
  <div>
    <div :class="wrapClasses">
      <div ref="editor" />
    </div>
    <div
      v-if="!isValid && validationError"
      class="d-block invalid-feedback"
    >
      {{ validationError }}
    </div>
  </div>
</template>

<script lang="ts">
import * as constants from '../lib/const'
import { api, HttpError } from '../api'
import { CodeJar } from 'codejar/codejar.js'
import type { CodeJar as CodeJarInstance } from 'codejar'
import { defineComponent } from 'vue'
import Prism from 'prismjs'
import { useAppStore } from '../stores/app'
import { withLineNumbers } from 'codejar/linenumbers.js'

export default defineComponent({
  computed: {
    grammar() {
      return {
        template: {
          inside: {
            boolean: /\b(?:true|false)\b/,
            comment: /\/\*[\s\S]*\*\//,
            function: RegExp(`\\b(?:${[...constants.BUILTIN_TEMPLATE_FUNCTIONS, ...this.appStore.vars.TemplateFunctions].join('|')})\\b`),
            keyword: /\b(?:if|else|end|range)\b/,
            number: /\b0x[\da-f]+\b|(?:\b\d+(?:\.\d*)?|\B\.\d+)(?:e[+-]?\d+)?/i,
            operator: /\b(?:eq|ne|lt|le|gt|ge)\b/,
            string: {
              greedy: true,
              pattern: /(["'])(?:\\(?:\r\n|[\s\S])|(?!\1)[^\\\r\n])*\1/,
            },

            variable: /(^|\s)\.\w+\b/,
          },

          pattern: /\{\{.*?\}\}/s,
        },
      }
    },

    wrapClasses() {
      return {
        'form-control': true,
        'is-invalid': this.state === false || !this.isValid,
        'is-valid': this.state === true && this.isValid,
        'template-editor': true,
      }
    },
  },

  data() {
    return {
      appStore: useAppStore(),
      emittedCode: '',
      isValid: true,
      jar: null as CodeJarInstance | null,
      validationError: '',
    }
  },

  emits: ['update:modelValue', 'valid-template'],

  methods: {
    highlight(editor: HTMLElement) {
      const code = editor.textContent || ''
      editor.innerHTML = Prism.highlight(code, this.grammar, 'template')
    },

    async validateTemplate(template: string) {
      if (template === '') {
        this.isValid = true
        this.validationError = ''
        this.$emit('valid-template', true)
        return
      }

      try {
        await api.put(`config-editor/validate-template?template=${encodeURIComponent(template)}`, undefined, false)
        this.isValid = true
        this.validationError = ''
        this.$emit('valid-template', true)
      } catch (err) {
        const httpErr = err as HttpError
        this.isValid = false
        this.validationError = String(httpErr.data || '').split(':1:')[1] || String(httpErr.data || err)
        this.$emit('valid-template', false)
      }
    },
  },

  mounted() {
    this.jar = CodeJar(this.$refs.editor as HTMLElement, withLineNumbers((editor: HTMLElement) => this.highlight(editor)), {
      indentOn: /[{(]$/,
      tab: ' '.repeat(2),
    })
    this.jar.onUpdate((code: string) => {
      this.validateTemplate(code)
      this.emittedCode = code
      this.$emit('update:modelValue', code)
    })
    this.jar.updateCode(this.modelValue)
  },

  name: 'TwitchBotTemplateEditor',

  props: {
    modelValue: {
      default: '',
      type: String,
    },

    state: {
      default: null,
      type: Boolean,
    },
  },

  watch: {
    modelValue(to: string, from: string) {
      if (to === from || to === this.emittedCode || !this.jar) {
        return
      }

      this.jar.updateCode(to)
    },
  },
})
</script>

<style>
.template-editor {
  color: #d7dde7;
  font-family: SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 87.5%;
  height: fit-content;
  padding: 0;
}

.template-editor .codejar-wrap {
  background-color: #111827;
  border: 1px solid #374151;
  border-radius: 0.375rem;
}

.template-editor .codejar-linenumbers {
  background-color: #0f172a;
  color: #7c8aa5;
  padding-right: 0.5em;
  text-align: right;
}

.template-editor .codejar-linenumbers div {
  padding-bottom: 0.5em;
  padding-top: 0.5em;
}

.template-editor .codejar-linenumbers + div {
  color: #d7dde7;
  margin-left: 35px;
  padding-bottom: 0.5em;
  padding-left: 0.5em !important;
  padding-top: 0.5em;
}

.template-editor .token.comment {
  color: #7c8aa5;
}

.template-editor .token.operator {
  color: #9aa6b2;
}

.template-editor .token.boolean,
.template-editor .token.number {
  color: #f59e0b;
}

.template-editor .token.keyword {
  color: #f472b6;
}

.template-editor .token.string {
  color: #86efac;
}

.template-editor .token.variable {
  background: rgba(255, 255, 255, 0.04);
  color: #c4b5fd;
}

.template-editor .token.function {
  color: #67e8f9;
}
</style>
