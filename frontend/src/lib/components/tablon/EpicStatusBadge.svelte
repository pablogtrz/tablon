<script lang="ts">
  import { cn } from '$lib/utils'
  import { EPIC_STATUS_META } from '$lib/meta'
  import type { EpicStatus } from '$lib/types'

  let {
    status,
    class: className,
  }: { status: EpicStatus; class?: string } = $props()

  const meta = $derived(EPIC_STATUS_META[status])
  const Icon = $derived(meta.icon)
</script>

<span
  class={cn(
    'inline-flex h-7 items-center gap-1.5 rounded-full px-2.5 text-xs font-semibold whitespace-nowrap',
    className,
  )}
  style="background: {meta.soft}; color: {meta.tone}"
>
  <Icon size={14} strokeWidth={1.8} aria-hidden="true" />
  {meta.label}
</span>
