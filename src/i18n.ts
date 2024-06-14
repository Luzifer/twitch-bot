import { createI18n } from 'vue-i18n'
import en from './langs/en.json'

const cookieSet = Object.fromEntries(document.cookie.split('; ')
  .map(el => el.split('=')
    .map(el => decodeURIComponent(el))))

export default createI18n({
  fallbackLocale: 'en',
  globalInjection: true,
  legacy: false,
  locale: cookieSet.lang || navigator?.language || 'en',
  messages: {
    en,
  },
})
