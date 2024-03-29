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

<script>
import * as constants from './const.js'
import axios from 'axios'
import { CodeJar } from 'codejar/codejar.js'
import Prism from 'prismjs'
import { withLineNumbers } from 'codejar/linenumbers.js'

export default {
  computed: {
    grammar() {
      return {
        template: {
          inside: {
            boolean: /\b(?:true|false)\b/,
            comment: /\/\*[\s\S]*\*\//,
            function: RegExp(`\\b(?:${[...constants.BUILTIN_TEMPLATE_FUNCTIONS, ...this.$root.vars.TemplateFunctions].join('|')})\\b`),
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
      emittedCode: '',
      isValid: true,
      jar: null,
      validationError: '',
    }
  },

  methods: {
    highlight(editor) {
      const code = editor.textContent
      editor.innerHTML = Prism.highlight(code, this.grammar, 'template')
    },

    validateTemplate(template) {
      if (template === '') {
        this.isValid = true
        this.validationError = ''
        this.$emit('valid-template', true)
        return
      }

      return axios.put(`config-editor/validate-template?template=${encodeURIComponent(template)}`)
        .then(() => {
          this.isValid = true
          this.validationError = ''
          this.$emit('valid-template', true)
        })
        .catch(resp => {
          this.isValid = false
          this.validationError = resp.response.data.split(':1:')[1]
          this.$emit('valid-template', false)
        })
    },
  },

  mounted() {
    this.jar = CodeJar(this.$refs.editor, withLineNumbers(this.highlight), {
      indentOn: /[{(]$/,
      tab: ' '.repeat(2),
    })
    this.jar.onUpdate(code => {
      this.validateTemplate(code)
      this.emittedCode = code
      this.$emit('input', code)
    })
    this.jar.updateCode(this.value)
  },

  name: 'TwitchBotEditorAppTemplateEditor',

  props: {
    state: {
      default: null,
      required: false,
      type: Boolean,
    },

    value: {
      default: '',
      required: false,
      type: String,
    },
  },

  watch: {
    value(to, from) {
      if (to === from || to === this.emittedCode) {
        return
      }

      this.jar.updateCode(to)
    },
  },
}
</script>

<style>
.template-editor {
  color: #444;
  font-family: SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace;
  font-size: 87.5%;
  height: fit-content;
  padding: 0;
}

.template-editor .codejar-wrap {
  background-color: #fff;
  border-radius: 0.25rem;
}

.template-editor .codejar-linenumbers {
  padding-right: 0.5em;
  text-align: right;
}

.template-editor .codejar-linenumbers div {
  padding-bottom: 0.5em;
  padding-top: 0.5em;
}

.template-editor .codejar-linenumbers + div {
  margin-left: 35px;
  padding-bottom: 0.5em;
  padding-left: 0.5em !important;
  padding-top: 0.5em;
}

.template-editor .token.comment {
	color: #7D8B99;
}

.template-editor .token.operator {
	color: #5F6364;
}

.template-editor .token.boolean,
.template-editor .token.number {
	color: #c92c2c;
}

.template-editor .token.keyword {
  color: #e83e8c;
}

.template-editor .token.string {
	color: #2f9c0a;
}

.template-editor .token.variable {
	color: #a67f59;
	background: rgba(255, 255, 255, 0.5);
}

.template-editor .token.function {
	color: #1990b8;
}
</style>
