export const motion = $state({
  reduced: false,
})

if (typeof window !== 'undefined') {
  const media = window.matchMedia('(prefers-reduced-motion: reduce)')
  motion.reduced = media.matches
  media.addEventListener('change', (event) => {
    motion.reduced = event.matches
  })
}
