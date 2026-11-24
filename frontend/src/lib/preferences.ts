const PREFIX = 'tablon.'

function storage(): Storage | null {
  if (typeof window === 'undefined') return null
  try {
    return window.localStorage
  } catch {
    return null
  }
}

export function readPreference(key: string): string | null {
  try {
    return storage()?.getItem(`${PREFIX}${key}`) ?? null
  } catch {
    return null
  }
}

export function writePreference(key: string, value: string): void {
  try {
    storage()?.setItem(`${PREFIX}${key}`, value)
  } catch {
    // Preferences are best-effort; private mode or a full quota must not break the UI.
  }
}

export function readBooleanPreference(key: string, fallback = false): boolean {
  const value = readPreference(key)
  if (value === null) return fallback
  return value === 'true'
}

export function writeBooleanPreference(key: string, value: boolean): void {
  writePreference(key, value ? 'true' : 'false')
}
