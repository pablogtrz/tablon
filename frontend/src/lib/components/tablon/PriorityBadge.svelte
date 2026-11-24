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
    'inline-flex h-[26px] items-center gap-1.5 rounded-full bg-[var(--soft)] px-2.5 text-[11.5px] font-semibold whitespace-nowrap text-[var(--tone)]',
    className,
  )}
  style="--tone: {meta.tone}; --soft: {meta.soft}"
>
  <Icon size={14} strokeWidth={1.8} class="shrink-0" aria-hidden="true" />
  {meta.label}
</span>
