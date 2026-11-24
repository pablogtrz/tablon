<script lang="ts">
  import { tick } from 'svelte'
  import BoardCard from './BoardCard.svelte'
  import Button from './Button.svelte'
  import {
    adjacentBoardStatus,
    boardStatusLabel,
    groupBoardIssues,
    hiddenBoardCounts,
    isBoardStatus,
  } from '$lib/board'
  import type { BoardColumn, HiddenBoardStatus } from '$lib/board'
  import { issuePath } from '$lib/router.svelte'
  import { cn } from '$lib/utils'
  import type { Epic, Issue, IssueStatus } from '$lib/types'

  let {
    projectRef,
    issues,
    epics,
    movingIds,
    truncated,
    onmove,
    onhidden,
    onopen,
  }: {
    projectRef: string
    issues: Issue[]
    epics: Epic[]
    movingIds: string[]
    truncated: boolean
    onmove: (issue: Issue, status: IssueStatus) => Promise<boolean>
    onhidden: (status: HiddenBoardStatus) => void
    onopen: (issue: Issue) => void
  } = $props()

  let board = $state<HTMLDivElement>()
  let draggingId = $state<string | null>(null)
  let dragOver = $state<IssueStatus | null>(null)
  let announcement = $state('')

  const grouped = $derived(groupBoardIssues(issues))
  const hidden = $derived(hiddenBoardCounts(issues))
  const epicMap = $derived(new Map(epics.map((epic) => [epic.id, epic.title])))
  const columns = $derived<Array<BoardColumn & { items: Issue[] }>>([
    { status: 'todo', label: 'Todo', items: grouped.todo },
    { status: 'in-progress', label: 'In Progress', items: grouped['in-progress'] },
    { status: 'in-review', label: 'In Review', items: grouped['in-review'] },
    { status: 'done', label: 'Done', items: grouped.done },
  ])

  async function announce(message: string) {
    announcement = ''
    await tick()
    announcement = message
  }

  function epicTitle(issue: Issue): string {
    if (!issue.epic_id) return ''
    return epicMap.get(issue.epic_id) ?? ''
  }

  function handleDragStart(issue: Issue, event: DragEvent) {
    draggingId = issue.id
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move'
      event.dataTransfer.setData('text/plain', issue.id)
    }
  }

  function handleDragOver(event: DragEvent, status: IssueStatus) {
    event.preventDefault()
    if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
    dragOver = status
  }

  function handleDragLeave(event: DragEvent) {
    const current = event.currentTarget as HTMLElement
    const next = event.relatedTarget as Node | null
    if (next && current.contains(next)) return
    dragOver = null
  }

  function handleDrop(event: DragEvent, status: IssueStatus) {
    event.preventDefault()
    const id = event.dataTransfer?.getData('text/plain') ?? ''
    const issue = issues.find((item) => item.id === id)
    dragOver = null
    draggingId = null
    if (!issue || issue.status === status) return
    void move(issue, status)
  }

  async function move(issue: Issue, status: IssueStatus) {
    const sourceStatus = issue.status
    const label = boardStatusLabel(status)
    const ok = await onmove(issue, status)
    await tick()

    if (ok) {
      await announce(`${issue.identifier} moved to ${label}`)
    } else {
      await announce(`Could not move ${issue.identifier}`)
    }

    if (isBoardStatus(status)) {
      board
        ?.querySelector<HTMLElement>(`[data-issue-id="${CSS.escape(issue.id)}"]`)
        ?.focus()
      return
    }

    const source = board?.querySelector<HTMLElement>(
      `[data-column-status="${sourceStatus}"]`,
    )
    const nextCard = source?.querySelector<HTMLElement>('[data-issue-id]')
    if (nextCard) {
      nextCard.focus()
    } else {
      source?.querySelector<HTMLElement>('h2')?.focus()
    }
  }

  function handleKeyMove(issue: Issue, direction: -1 | 1) {
    const next = adjacentBoardStatus(issue.status, direction)
    if (!next) return
    void move(issue, next)
  }

  function openIssue(issue: Issue) {
    onopen(issue)
  }

  function hrefFor(issue: Issue): string {
    return issuePath(projectRef, issue.identifier)
  }
</script>

<div>
  {#if hidden.backlog > 0 || hidden.canceled > 0}
    <div class="mb-3 flex flex-wrap items-center gap-2 text-[11px] text-[var(--text-3)]">
      <span class="tnum">
        {hidden.backlog} backlog · {hidden.canceled} canceled not shown on the board.
      </span>
      {#if hidden.backlog > 0}
        <Button variant="ghost" class="h-7 px-2 text-[11px]" onclick={() => onhidden('backlog')}>
          View backlog
        </Button>
      {/if}
      {#if hidden.canceled > 0}
        <Button variant="ghost" class="h-7 px-2 text-[11px]" onclick={() => onhidden('canceled')}>
          View canceled
        </Button>
      {/if}
    </div>
  {/if}

  {#if truncated}
    <p class="mb-3 rounded-lg border border-border bg-[var(--surface-1)] px-3 py-2 text-[11px] text-[var(--text-3)]">
      Showing the first 1,000 issues. Use the Issues view to filter a larger project.
    </p>
  {/if}

  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    bind:this={board}
    class="-mx-4 flex snap-x snap-mandatory gap-3 overflow-x-auto overscroll-x-contain px-4 pb-3 lg:mx-0 lg:grid lg:grid-cols-4 lg:gap-4 lg:overflow-visible lg:px-0"
    role="region"
    aria-label="Issue board"
  >
    {#each columns as column (column.status)}
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <section
        data-column-status={column.status}
        class={cn(
          'flex w-[min(82vw,300px)] shrink-0 snap-start flex-col rounded-2xl border bg-[var(--bg-2)]/40 transition-colors lg:w-auto',
          dragOver === column.status
            ? 'border-brand/60 bg-[var(--brand-soft)]'
            : 'border-border',
        )}
        aria-labelledby={`board-${column.status}`}
        ondragover={(event) => handleDragOver(event, column.status)}
        ondragleave={handleDragLeave}
        ondrop={(event) => handleDrop(event, column.status)}
      >
        <header class="flex items-center justify-between gap-2 border-b border-border px-3 py-2.5">
          <h2
            id={`board-${column.status}`}
            tabindex="-1"
            class="text-xs font-semibold tracking-[-.01em] outline-none"
          >
            {column.label}
          </h2>
          <span class="rounded-full bg-[var(--surface-2)] px-1.5 py-0.5 text-[10px] tnum text-[var(--text-3)]">
            {column.items.length}
          </span>
        </header>

        <ul class="flex min-h-40 flex-col gap-2 p-2" aria-label={`${column.label} issues`}>
          {#each column.items as issue (issue.id)}
            <BoardCard
              {issue}
              href={hrefFor(issue)}
              epicTitle={epicTitle(issue)}
              saving={movingIds.includes(issue.id)}
              menuId={`move-${issue.id}`}
              onopen={() => openIssue(issue)}
              onmove={(item, status) => void move(item, status)}
              onkeymove={handleKeyMove}
              ondragstart={handleDragStart}
              ondragend={() => {
                draggingId = null
                dragOver = null
              }}
            />
          {:else}
            <li
              class={cn(
                'grid min-h-24 place-items-center rounded-lg border border-dashed border-border px-3 text-center text-[11px] text-[var(--text-3)]',
                draggingId ? 'border-brand/40 text-[var(--text-2)]' : '',
              )}
            >
              No issues
            </li>
          {/each}
        </ul>
      </section>
    {/each}
  </div>

  <p class="sr-only" role="status" aria-live="polite">{announcement}</p>
</div>
