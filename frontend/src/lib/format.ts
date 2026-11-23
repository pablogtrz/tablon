const relative = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' })
const absolute = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' })

const units: Array<[Intl.RelativeTimeFormatUnit, number]> = [
  ['year', 31_536_000_000],
  ['month', 2_592_000_000],
  ['week', 604_800_000],
  ['day', 86_400_000],
  ['hour', 3_600_000],
  ['minute', 60_000],
]

export function relativeTime(value: string): string {
  const date = new Date(value)
  const diff = date.getTime() - Date.now()
  for (const [unit, size] of units) {
    if (Math.abs(diff) >= size) {
      return relative.format(Math.round(diff / size), unit)
    }
  }
  return 'just now'
}

export function formatDate(value: string): string {
  return absolute.format(new Date(value))
}

export function plural(count: number, singular: string, pluralForm = `${singular}s`): string {
  return `${count} ${count === 1 ? singular : pluralForm}`
}
