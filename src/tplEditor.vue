<template>
  <div :class="wrapClasses">
    <div ref="editor" />
  </div>
</template>

<script>
import * as constants from './const.js'
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
        'is-invalid': this.state === false,
        'is-valid': this.state === true,
        'template-editor': true,
      }
    },
  },

  data() {
    return {
      emittedCode: '',
      jar: null,
    }
  },

  methods: {
    highlight(editor) {
      const code = editor.textContent
      editor.innerHTML = Prism.highlight(code, this.grammar, 'template')
    },
  },

  mounted() {
    this.jar = CodeJar(this.$refs.editor, withLineNumbers(this.highlight))
    this.jar.onUpdate(code => {
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
  background-color: #fff;
  border-radius: 0.25rem;
  color: #444;
  font-family: SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace;
  font-size: 87.5%;
  height: fit-content;
  padding: 0;
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
