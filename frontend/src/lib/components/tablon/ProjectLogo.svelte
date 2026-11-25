<script lang="ts">
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
    from && to
      ? `linear-gradient(${angle}deg, ${from} 0%, ${to} 100%)`
      : `var(--logo-${index})`,
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
    'inline-grid shrink-0 place-items-center font-semibold text-white/95 shadow-[inset_0_1px_0_rgba(255,255,255,.18)]',
    sizes[size],
    className,
  )}
  style="background-image: {background}"
  aria-hidden="true"
>
  {#if emoji}
    <span class="text-[1.05em] leading-none">{emoji}</span>
  {:else}
    {initial}
  {/if}
</span>
