import { defineStore } from 'pinia'

import { api, setApiAuthToken } from '../api'
import type { EditorVarsResponse, StatusResponse } from '../types'

export const useAppStore = defineStore('app', {
  actions: {
    async fetchVars() {
      const vars = await api.get<EditorVarsResponse>('editor/vars.json', false)
      this.vars = typeof vars === 'string' ? JSON.parse(vars) : vars
    },

    pushToast(message: string, variant: string) {
      const id = this.toastId++
      this.toasts.push({ id, message, variant })
      window.setTimeout(() => {
        this.toasts = this.toasts.filter(toast => toast.id !== id)
      }, 3000)
    },

    setAuthToken(token: string | null) {
      this.authToken = token
      setApiAuthToken(token)
    },

    setChangePending(value: boolean) {
      this.changePending = value
    },

    setError(message: string | null) {
      this.error = message
    },

    setLoadingData(value: boolean) {
      this.loadingData = value
    },

    setStatus(status: StatusResponse) {
      this.status = status
    },

    toastError(message: string) {
      this.pushToast(message, 'danger')
    },

    toastInfo(message: string) {
      this.pushToast(message, 'info')
    },

    toastSuccess(message: string) {
      this.pushToast(message, 'success')
    },
  },

  getters: {
    isAuthenticated: state => Boolean(state.authToken),
  },

  state: () => ({
    authToken: null as string | null,
    changePending: false,
    error: null as string | null,
    loadingData: false,
    status: {
      checks: [],
      overall_status_success: false,
    } as StatusResponse,
    toastId: 0,
    toasts: [] as Array<{ id: number, message: string, variant: string }>,
    vars: {
      DefaultBotScopes: [],
      IRCBadges: [],
      KnownEvents: [],
      TemplateFunctions: [],
      TwitchClientID: '',
      Version: '',
    } as EditorVarsResponse,
  }),
})
