export const BUILTIN_TEMPLATE_FUNCTIONS = [
  'and',
  'call',
  'html',
  'index',
  'slice',
  'js',
  'len',
  'not',
  'or',
  'print',
  'printf',
  'println',
  'urlquery',
]

export const CRON_VALIDATION = /^(?:(?:@every (?:\d+(?:s|m|h))+)|(?:(?:(?:(?:\d+,)+\d+|(?:\d+(?:\/|-)\d+)|\d+|\*|\*\/\d+)(?: |$)){5}))$/
export const NANO = 1000000000

export const NOTIFY_CHANGE_PENDING = 'changePending'
export const NOTIFY_CONFIG_RELOAD = 'configReload'
export const NOTIFY_ERROR = 'error'
export const NOTIFY_FETCH_ERROR = 'fetchError'
export const NOTIFY_LOADING_DATA = 'loadingData'

export const REGEXP_USER = /^[a-z0-9_]{4,25}$/
