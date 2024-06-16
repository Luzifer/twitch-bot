/* eslint-disable no-unused-vars */
/* global RequestInit, TimerHandler */

import { Emitter, EventType } from 'mitt'

type CheckAccessFunction = (resp: Response) => Response

type EditorVars = {
  DefaultBotScopes: string[]
  IRCBadges: string[]
  KnownEvents: string[]
  TemplateFunctions: string[]
  TwitchClientID: string
  Version: string
}

type ParseResponseFunction = (resp: Response) => Promise<any>
type TickerRegisterFunction = (id: string, func: TimerHandler, intervalMs: number) => void
type TickerUnregisterFunction = (id: string) => void

type UserInfo = {
  display_name: string
  id: string
  login: string
  profile_image_url: string
}

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    bus: Emitter<Record<EventType, unknown>>

    // On the $root
    check403: CheckAccessFunction
    fetchOpts: RequestInit
    parseResponseFromJSON: ParseResponseFunction
    registerTicker: TickerRegisterFunction
    unregisterTicker: TickerUnregisterFunction
    userInfo: UserInfo | null
    vars: EditorVars | null
  }
}

export {} // Important! See note.
