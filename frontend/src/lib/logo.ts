export function hasCustomGradient(from?: string, to?: string): boolean {
  return Boolean(from && to)
}

export function logoGradient(from: string, to: string, angle: number): string {
  return `linear-gradient(${angle}deg, ${from} 0%, ${to} 100%)`
}
