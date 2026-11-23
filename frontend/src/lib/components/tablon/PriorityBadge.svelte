<script lang="ts">
  import { PRIORITY_META } from '$lib/meta'
  import { cn } from '$lib/utils'
  import type { IssuePriority } from '$lib/types'

  let {
    priority,
    class: className,
  }: { priority: IssuePriority; class?: string } = $props()

  const meta = $derived(PRIORITY_META[priority])
  const Icon = $derived(meta.icon)
</script>

<span
  class={cn(
    'inline-flex h-[22px] items-center gap-1.5 rounded-full bg-white/[.035] px-2 text-[10.5px] font-medium whitespace-nowrap text-muted-foreground',
    className,
  )}
  style="--tone: {meta.tone}"
>
  <Icon size={12} strokeWidth={1.6} class="shrink-0 text-[var(--tone)]" aria-hidden="true" />
  {meta.label}
</span>
