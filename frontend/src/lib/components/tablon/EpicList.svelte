<script lang="ts">
  import Pencil from '@lucide/svelte/icons/pencil'
  import Trash2 from '@lucide/svelte/icons/trash-2'
  import Button from './Button.svelte'
  import EpicStatusBadge from './EpicStatusBadge.svelte'
  import EpicStatusBar from './EpicStatusBar.svelte'
  import { plural, relativeTime } from '$lib/format'
  import { epicPath } from '$lib/router.svelte'
  import { ISSUE_STATUSES } from '$lib/types'
  import { cn } from '$lib/utils'
  import type { Epic, Issue, IssueStatus } from '$lib/types'

  let {
    projectRef,
    epics,
    issues,
    loading,
    statsLoading = false,
    onopen,
    onedit,
    ondelete,
  }: {
    projectRef: string
    epics: Epic[]
    issues: Issue[]
    loading: boolean
    statsLoading?: boolean
    onopen: (epic: Epic) => void
    onedit: (epic: Epic) => void
    ondelete: (epic: Epic) => void
  } = $props()

  const countsByEpic = $derived.by(() => {
    const map = new Map<string, Record<IssueStatus, number>>()
    for (const epic of epics) {
      map.set(
        epic.id,
        Object.fromEntries(ISSUE_STATUSES.map((status) => [status, 0])) as Record<IssueStatus, number>,
      )
    }
    for (const issue of issues) {
      if (!issue.epic_id) continue
      const counts = map.get(issue.epic_id)
      if (!counts) continue
      counts[issue.status] += 1
    }
    return map
  })

  function countsFor(epicID: string): Record<IssueStatus, number> {
    return (
      countsByEpic.get(epicID) ??
      (Object.fromEntries(ISSUE_STATUSES.map((status) => [status, 0])) as Record<IssueStatus, number>)
    )
  }

  function totalOf(counts: Record<IssueStatus, number>): number {
    return ISSUE_STATUSES.reduce((sum, status) => sum + counts[status], 0)
  }

  function excerpt(markdown: string): string {
    return markdown
      .replace(/```[\s\S]*?```/g, ' ')
      .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')
      .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
      .replace(/^#{1,6}\s+/gm, '')
      .replace(/^\s*(?:[-+]|\d+\.)\s+/gm, '')
      .replace(/[*_~`>#]+/g, ' ')
      .replace(/\s+/g, ' ')
      .trim()
  }

  function open(event: MouseEvent, epic: Epic) {
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    onopen(epic)
  }
</script>

{#if loading}
  <div class="mt-4 grid grid-cols-1 gap-3 xl:grid-cols-2" aria-hidden="true">
    {#each Array(2) as _}
      <div class="h-52 animate-pulse rounded-2xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none"></div>
    {/each}
  </div>
{:else if epics.length === 0}
  <div class="mt-4 grid place-items-center rounded-2xl border border-border bg-[var(--surface-1)] px-6 py-12 text-center">
    <h3 class="text-sm font-semibold">No epics yet</h3>
    <p class="mt-1 max-w-sm text-xs leading-relaxed text-[var(--text-2)]">
      Group related issues into an epic to keep the board readable.
    </p>
  </div>
{:else}
  <div class="mt-4 grid grid-cols-1 gap-3 xl:grid-cols-2">
    {#each epics as epic (epic.id)}
      {@const counts = countsFor(epic.id)}
      {@const total = totalOf(counts)}
      {@const done = counts.done}
      {@const progress = total === 0 ? 0 : Math.round((done / total) * 100)}
      {@const isDone = epic.status === 'done'}
      {@const href = epicPath(projectRef, epic.id)}
      <article
        class={cn(
          'group relative flex flex-col rounded-2xl border p-4 transition-colors',
          'has-[a:focus-visible]:ring-2 has-[a:focus-visible]:ring-ring has-[a:focus-visible]:ring-offset-2 has-[a:focus-visible]:ring-offset-background',
          isDone
            ? 'border-[var(--epic-done-border)] bg-[var(--epic-done-bg)] hover:border-[var(--status-done)]'
            : 'border-border bg-[var(--surface-1)] hover:border-[var(--border-strong)] hover:bg-[var(--surface-2)]',
        )}
      >
        <header class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <h3 class={cn('truncate text-sm font-semibold tracking-[-.01em]', isDone ? 'text-[var(--text-2)]' : 'text-foreground')}>
              <a
                {href}
                class="card-link cursor-pointer rounded outline-none after:absolute after:inset-0 after:rounded-2xl after:content-[''] focus-visible:outline-none"
                onclick={(event) => open(event, epic)}
              >
                {epic.title}
              </a>
            </h3>
            <p class="mt-0.5 text-[11px] text-[var(--text-3)]">
              Updated {relativeTime(epic.updated_at)} · {plural(total, 'task')}
            </p>
          </div>
          <EpicStatusBadge status={epic.status} />
        </header>

        <p class={cn('mt-3 line-clamp-2 min-h-10 text-xs leading-relaxed', isDone ? 'text-[var(--text-3)]' : 'text-[var(--text-2)]')}>
          {excerpt(epic.description) || 'No description yet.'}
        </p>

        <div class="mt-auto pt-4">
          {#if statsLoading}
            <div class="h-2 animate-pulse rounded-full bg-white/[.06] motion-reduce:animate-none" aria-hidden="true"></div>
            <p class="mt-2 text-[10.5px] text-[var(--text-3)]">Counting tasks…</p>
          {:else}
            <div class="flex items-center justify-between gap-3 text-[11px] text-[var(--text-3)]">
              <span class="tnum">{done}/{total} done</span>
              <span class="tnum">{progress}%</span>
            </div>
            <progress
              class="epic-progress mt-2"
              value={done}
              max={Math.max(total, 1)}
              style="--progress-fill: {isDone ? 'var(--status-done)' : 'var(--brand)'}"
              aria-label={`${done} of ${total} tasks done in ${epic.title}`}
            ></progress>
            <EpicStatusBar {counts} class="mt-4" />
          {/if}
        </div>

        <footer class="relative z-10 mt-3 flex justify-end gap-1.5 opacity-100 transition-opacity md:opacity-0 md:group-hover:opacity-100 md:group-focus-within:opacity-100">
          <Button variant="ghost" class="h-8 px-2 text-xs" onclick={() => onedit(epic)}>
            <Pencil size={13} strokeWidth={1.6} aria-hidden="true" />
            Edit
          </Button>
          <Button variant="ghost" class="h-8 px-2 text-xs text-destructive hover:text-destructive" onclick={() => ondelete(epic)}>
            <Trash2 size={13} strokeWidth={1.6} aria-hidden="true" />
            Delete
          </Button>
        </footer>
      </article>
    {/each}
  </div>
{/if}
