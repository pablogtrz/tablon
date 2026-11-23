<script lang="ts">
  import { cn } from '$lib/utils'

  let {
    gradient = 0,
    emoji = '',
    name,
    size = 'md',
    class: className,
  }: {
    gradient?: number
    emoji?: string
    name: string
    size?: 'sm' | 'md' | 'lg'
    class?: string
  } = $props()

  const index = $derived(Math.abs(Math.trunc(gradient)) % 8)
  const initial = $derived(name.trim().charAt(0).toUpperCase() || 'T')
  const sizes = {
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
  style="background-image: var(--logo-{index})"
  aria-hidden="true"
>
  {#if emoji}
    <span class="text-[1.05em] leading-none">{emoji}</span>
  {:else}
    {initial}
  {/if}
</span>
