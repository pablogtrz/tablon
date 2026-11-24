import { motion } from './motion.svelte'

const DURATION = 1200
const PARTICLE_COUNT = 110
const COLORS = ['--brand', '--status-done', '--status-todo', '--status-progress', '--status-review']

interface Particle {
  x: number
  y: number
  vx: number
  vy: number
  size: number
  color: string
  rotation: number
  spin: number
  shape: 'rect' | 'circle'
}

let activeCleanup: (() => void) | null = null

function readColors(): string[] {
  const style = getComputedStyle(document.documentElement)
  return COLORS.map((token) => style.getPropertyValue(token).trim()).filter(Boolean)
}

export function burstConfetti(): void {
  if (typeof window === 'undefined' || motion.reduced) return
  activeCleanup?.()

  const canvas = document.createElement('canvas')
  canvas.setAttribute('aria-hidden', 'true')
  canvas.style.cssText =
    'position:fixed;inset:0;width:100%;height:100%;pointer-events:none;z-index:100;'
  document.body.appendChild(canvas)

  const context = canvas.getContext('2d')!

  const ratio = Math.min(window.devicePixelRatio || 1, 2)
  const width = window.innerWidth
  const height = window.innerHeight
  canvas.width = Math.round(width * ratio)
  canvas.height = Math.round(height * ratio)
  context.scale(ratio, ratio)

  const colors = readColors()
  const originX = width / 2
  const originY = height * 0.42
  const particles: Particle[] = Array.from({ length: PARTICLE_COUNT }, () => {
    const angle = Math.random() * Math.PI * 2
    const speed = 3 + Math.random() * 7
    return {
      x: originX + (Math.random() - 0.5) * 40,
      y: originY + (Math.random() - 0.5) * 20,
      vx: Math.cos(angle) * speed,
      vy: Math.sin(angle) * speed - 3,
      size: 4 + Math.random() * 6,
      color: colors[Math.floor(Math.random() * colors.length)] ?? '#FFA443',
      rotation: Math.random() * Math.PI,
      spin: (Math.random() - 0.5) * 0.3,
      shape: Math.random() > 0.45 ? 'rect' : 'circle',
    }
  })

  const start = performance.now()
  let frame = 0
  let finished = false

  function cleanup() {
    if (finished) return
    finished = true
    cancelAnimationFrame(frame)
    canvas.remove()
    if (activeCleanup === cleanup) activeCleanup = null
  }

  activeCleanup = cleanup

  function draw(now: number) {
    const elapsed = now - start
    const progress = Math.min(elapsed / DURATION, 1)
    context.clearRect(0, 0, width, height)
    context.globalAlpha = Math.max(0, 1 - progress * 1.1)

    for (const particle of particles) {
      particle.vy += 0.16
      particle.vx *= 0.995
      particle.x += particle.vx
      particle.y += particle.vy
      particle.rotation += particle.spin

      context.save()
      context.translate(particle.x, particle.y)
      context.rotate(particle.rotation)
      context.fillStyle = particle.color
      if (particle.shape === 'rect') {
        context.fillRect(-particle.size / 2, -particle.size / 4, particle.size, particle.size / 2)
      } else {
        context.beginPath()
        context.arc(0, 0, particle.size / 3, 0, Math.PI * 2)
        context.fill()
      }
      context.restore()
    }

    if (progress < 1) {
      frame = requestAnimationFrame(draw)
    } else {
      cleanup()
    }
  }

  frame = requestAnimationFrame(draw)
}
