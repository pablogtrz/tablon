<script lang="ts">
  import { cn } from '$lib/utils'
  import { ISSUE_STATUS_META } from '$lib/meta'
  import type { IssueStatus } from '$lib/types'

  let {
    status,
    class: className,
  }: { status: IssueStatus; class?: string } = $props()

  const meta = $derived(ISSUE_STATUS_META[status])
  const Icon = $derived(meta.icon)
</script>

<span
  class={cn(
    'inline-flex h-[26px] w-full items-center gap-2 rounded-[6px] px-2 text-[11px] font-medium text-foreground',
    className,
  )}
  style="background: {meta.soft}; --tone: {meta.tone}"
>
  <Icon size={13} strokeWidth={1.6} class="shrink-0 text-[var(--tone)]" aria-hidden="true" />
  <span class="truncate">{meta.label}</span>
</span>
