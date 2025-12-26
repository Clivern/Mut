import { i18n } from '@/i18n'

export const USER_LANGUAGE_EN = 'en'
export const USER_LANGUAGE_FR = 'fr'
export const USER_THEME_DEFAULT = 'default'
export const USER_THEME_BLUE = 'blue'
export const USER_THEME_SLATE = 'slate'
export const USER_THEME_EMERALD = 'emerald'
export const USER_THEME_DARK = 'dark'

const THEME_CLASS_MAP = {
  [USER_THEME_BLUE]: 'theme-blue',
  [USER_THEME_SLATE]: 'theme-slate',
  [USER_THEME_EMERALD]: 'theme-emerald',
  [USER_THEME_DARK]: 'theme-dark',
}

const VALID_THEMES = new Set([
  USER_THEME_DEFAULT,
  USER_THEME_BLUE,
  USER_THEME_SLATE,
  USER_THEME_EMERALD,
  USER_THEME_DARK,
])

export function applyTheme(name) {
  const root = document.documentElement
  Object.values(THEME_CLASS_MAP).forEach((className) => {
    root.classList.remove(className)
  })
  const themeClass = THEME_CLASS_MAP[name]
  if (themeClass) {
    root.classList.add(themeClass)
  }
  try {
    localStorage.setItem('_ftheme', name)
  } catch (_) {}
  return name
}

export function applyLocale(locale) {
  i18n.global.locale.value = locale
  try {
    localStorage.setItem('_flocale', locale)
  } catch (_) {}
  return locale
}

export function applyUserPreferences(user) {
  if (!user) return
  applyTheme(user.theme)
  applyLocale(user.language)
}

export function readStoredTheme() {
  try {
    const stored = localStorage.getItem('_ftheme')
    if (stored && VALID_THEMES.has(stored)) return stored
  } catch (_) {}
  return USER_THEME_DEFAULT
}

export function readStoredLocale() {
  try {
    const stored = localStorage.getItem('_flocale')
    if (stored === USER_LANGUAGE_FR) return USER_LANGUAGE_FR
  } catch (_) {}
  return USER_LANGUAGE_EN
}
