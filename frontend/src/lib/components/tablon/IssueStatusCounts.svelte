<script lang="ts">
  import { ISSUE_STATUS_META } from '$lib/meta'
  import { ISSUE_STATUSES, ISSUE_STATUS_LABELS } from '$lib/types'
  import type { IssueStatus } from '$lib/types'
  import { cn } from '$lib/utils'

  let {
    counts,
    label = '',
    muted = false,
    class: className,
  }: {
    counts: Record<IssueStatus, number>
    label?: string
    muted?: boolean
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
    <div class="mb-2 flex items-center justify-between gap-3 text-[10.5px] text-[var(--text-3)]">
      <span>{label}</span>
      <span class="tnum">{total} {total === 1 ? 'task' : 'tasks'}</span>
    </div>
  {/if}

  {#if total === 0}
    <p class="text-[11px] text-[var(--text-3)]">No tasks yet</p>
  {:else}
    <dl class="m-0 flex flex-wrap gap-x-5 gap-y-2" aria-label={`Tasks by status: ${summary}`}>
      {#each ISSUE_STATUSES as status (status)}
        {#if counts[status] > 0}
          <div class="flex min-w-10 flex-col gap-0.5">
            <dt class="flex items-center gap-1.5 text-[10.5px] leading-none text-[var(--text-3)]">
              <span
                class="size-1.5 shrink-0 rounded-full"
                style="background: {ISSUE_STATUS_META[status].tone}"
                aria-hidden="true"
              ></span>
              {ISSUE_STATUS_LABELS[status]}
            </dt>
            <dd class={cn('m-0 text-sm font-semibold leading-tight tnum', muted ? 'text-[var(--text-2)]' : 'text-foreground')}>
              {counts[status]}
            </dd>
          </div>
        {/if}
      {/each}
    </dl>
  {/if}
</div>
