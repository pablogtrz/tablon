<script lang="ts">
  import type { Component } from 'svelte'
  import { cn } from '$lib/utils'

  let {
    label,
    tone,
    icon: Icon,
    class: className,
  }: { label: string; tone: string; icon?: Component; class?: string } = $props()
</script>

<span
  class={cn(
    'inline-flex h-[22px] items-center gap-1.5 rounded-full bg-white/[.04] px-2 text-[10.5px] font-medium tracking-[-.005em] whitespace-nowrap text-muted-foreground',
    className,
  )}
  style="--tone: {tone}"
>
  {#if Icon}
    <Icon size={12} strokeWidth={1.6} class="shrink-0 text-[var(--tone)]" aria-hidden="true" />
  {:else}
    <span class="size-1.5 shrink-0 rounded-full bg-[var(--tone)]" aria-hidden="true"></span>
  {/if}
  {label}
</span>
