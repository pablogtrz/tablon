<script lang="ts">
  import { cn } from '$lib/utils'
  import { ISSUE_STATUS_META } from '$lib/meta'
  import { ISSUE_STATUSES, ISSUE_STATUS_LABELS } from '$lib/types'
  import type { IssueStatus } from '$lib/types'

  let {
    counts,
    label = '',
    showLegend = true,
    class: className,
  }: {
    counts: Record<IssueStatus, number>
    label?: string
    showLegend?: boolean
    class?: string
  } = $props()

  const total = $derived(ISSUE_STATUSES.reduce((sum, status) => sum + counts[status], 0))
  const summary = $derived(
    ISSUE_STATUSES.filter((status) => counts[status] > 0)
      .map((status) => `${counts[status]} ${ISSUE_STATUS_LABELS[status]}`)
      .join(', '),
  )
</script>

<div class={cn('min-w-0', className)}>
  {#if label}
    <div class="mb-1.5 flex items-center justify-between gap-3 text-[10.5px] text-[var(--text-3)]">
      <span>{label}</span>
      <span class="tnum">{total} tasks</span>
    </div>
  {/if}
  <div
    class="flex h-2 gap-[2px] overflow-hidden rounded-full bg-white/[.06]"
    role="img"
    aria-label={total === 0 ? 'No tasks yet' : `Tasks by status: ${summary}`}
  >
    {#each ISSUE_STATUSES as status (status)}
      {#if counts[status] > 0}
        <span
          class="h-full rounded-full"
          style="width: {(counts[status] / total) * 100}%; background: {ISSUE_STATUS_META[status].tone}"
        ></span>
      {/if}
    {/each}
  </div>

  {#if showLegend && total > 0}
    <div class="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-[10.5px] text-[var(--text-3)]">
      {#each ISSUE_STATUSES as status (status)}
        {#if counts[status] > 0}
          <span class="inline-flex items-center gap-1.5">
            <span
              class="size-1.5 shrink-0 rounded-full"
              style="background: {ISSUE_STATUS_META[status].tone}"
              aria-hidden="true"
            ></span>
            <span class="tnum">{counts[status]}</span>
            {ISSUE_STATUS_LABELS[status]}
          </span>
        {/if}
      {/each}
    </div>
  {/if}
</div>
