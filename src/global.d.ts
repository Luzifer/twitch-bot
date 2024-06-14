/* eslint-disable no-unused-vars */

import { Emitter, EventType } from 'mitt'

type EditorVars = {
  DefaultBotScopes: string[]
  IRCBadges: string[]
  KnownEvents: string[]
  TemplateFunctions: string[]
  TwitchClientID: string
  Version: string
}

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
    userInfo: UserInfo | null
    vars: EditorVars | null
  }
}

export {} // Important! See note.
