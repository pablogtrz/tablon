<script lang="ts">
  import { hasCustomGradient, logoGradient } from '$lib/logo'
  import { cn } from '$lib/utils'

  let {
    gradient = 0,
    emoji = '',
    from = '',
    to = '',
    angle = 135,
    name,
    size = 'md',
    class: className,
  }: {
    gradient?: number
    emoji?: string
    from?: string
    to?: string
    angle?: number
    name: string
    size?: 'xs' | 'sm' | 'md' | 'lg'
    class?: string
  } = $props()

  const index = $derived(Math.abs(Math.trunc(gradient)) % 8)
  const initial = $derived(name.trim().charAt(0).toUpperCase() || 'T')
  const background = $derived(
    hasCustomGradient(from, to) ? logoGradient(from, to, angle) : `var(--logo-${index})`,
  )
  const sizes = {
    xs: 'size-6 rounded-md text-[11px]',
    sm: 'size-8 rounded-lg text-sm',
    md: 'size-10 rounded-[10px] text-base',
    lg: 'size-12 rounded-xl text-lg',
  }
</script>

<span
  class={cn(
    'inline-grid shrink-0 place-items-center overflow-hidden font-semibold text-white/95 shadow-[inset_0_1px_0_rgba(255,255,255,.18)]',
    sizes[size],
    className,
  )}
  style="background-image: {background}"
  aria-hidden="true"
>
  {#if emoji}
    <span class="max-w-full overflow-hidden text-[1.05em] leading-none">{emoji}</span>
  {:else}
    {initial}
  {/if}
</span>
