export type JsonPrimitive = boolean | number | string | null
export type JsonValue = JsonPrimitive | JsonValue[] | { [key: string]: JsonValue }
export type JsonObject = { [key: string]: JsonValue }

export interface StatusResponseCheck {
  description: string
  error?: string
  name: string
  success: boolean
}

export interface StatusResponse {
  checks: StatusResponseCheck[]
  overall_status_success: boolean
}

export interface EditorVarsResponse {
  DefaultBotScopes: string[]
  IRCBadges: string[]
  KnownEvents: string[]
  TemplateFunctions: string[]
  TwitchClientID: string
  Version: string
}

export interface ConfigAuthToken {
  modules: string[]
  name: string
  token: string
}

export type AuthTokensResponse = Record<string, ConfigAuthToken>

export interface AuthURLsResponse {
  available_extended_scopes: Record<string, string>
  update_bot_token: string
  update_channel_scopes: string
}

export interface GeneralConfig {
  bot_editors: string[]
  bot_name?: string
  channel_has_token: Record<string, boolean>
  channel_scopes: Record<string, string[]>
  channels: string[]
}

export interface TwitchUser {
  display_name: string
  id: string
  login: string
  profile_image_url: string
}

export type ActionDocumentationFieldType = 'bool' | 'duration' | 'int64' | 'string' | 'stringslice'

export interface ActionDocumentationField {
  default: string
  default_comment: string
  description: string
  key: string
  long: boolean
  name: string
  optional: boolean
  support_template: boolean
  type: ActionDocumentationFieldType
}

export interface ActionDocumentation {
  description: string
  fields: ActionDocumentationField[]
  name: string
  type: string
}

export interface RuleAction {
  attributes: Record<string, unknown>
  type: string
}

export interface Rule {
  actions?: RuleAction[]
  channel_cooldown?: number | string
  cooldown?: number | string
  description?: string
  disable?: boolean
  disable_on?: string[]
  disable_on_match_messages?: string[]
  disable_on_offline?: boolean
  disable_on_permit?: boolean
  disable_on_template?: string
  enable_on?: string[]
  match_channels?: string[]
  match_event?: string
  match_message?: string | null
  match_users?: string[]
  skip_cooldown_for?: string[]
  subscribe_from?: string
  user_cooldown?: number | string
  uuid?: string
}

export interface AutoMessage {
  channel?: string
  cron?: string
  disable?: boolean
  disable_on_template?: string
  message?: string
  message_interval?: number
  only_on_live?: boolean
  use_action?: boolean
  uuid?: string
}

export type RaffleStatus = 'active' | 'ended' | 'planned'

export interface RaffleEntry {
  drawResponse?: string
  enteredAs: string
  enteredAt: string
  id: number
  multiplier: number
  speakUpUntil?: string
  userDisplayName: string
  userID: string
  userLogin: string
  wasPicked: boolean
  wasRedrawn: boolean
}

export interface Raffle {
  allowEveryone: boolean
  allowFollower: boolean
  allowSubscriber: boolean
  allowVIP: boolean
  autoStartAt: string | null
  channel: string
  closeAfter: number
  closeAt: string | null
  entries?: RaffleEntry[]
  id: number
  keyword: string
  minFollowAge: number
  multiFollower: number
  multiSubscriber: number
  multiVIP: number
  status: RaffleStatus
  textClose: string
  textClosePost: boolean
  textEntry: string
  textEntryFail: string
  textEntryFailPost: boolean
  textEntryPost: boolean
  textReminder: string
  textReminderInterval: number
  textReminderPost: boolean
  textWin: string
  textWinPost: boolean
  title: string
  waitForResponse: number
}

export interface ConfigNotifyMessage {
  msg_type: string
}
