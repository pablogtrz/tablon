<script lang="ts">
  import Check from '@lucide/svelte/icons/check'
  import GripVertical from '@lucide/svelte/icons/grip-vertical'
  import LoaderCircle from '@lucide/svelte/icons/loader-circle'
  import AgentChip from './AgentChip.svelte'
  import IssueId from './IssueId.svelte'
  import PriorityBadge from './PriorityBadge.svelte'
  import { ISSUE_STATUS_OPTIONS } from '$lib/meta'
  import { cn } from '$lib/utils'
  import type { Issue, IssueStatus } from '$lib/types'

  let {
    issue,
    href,
    epicTitle,
    saving,
    menuId,
    onopen,
    onmove,
    onkeymove,
    ondragstart,
    ondragend,
  }: {
    issue: Issue
    href: string
    epicTitle: string
    saving: boolean
    menuId: string
    onopen: () => void
    onmove: (issue: Issue, status: IssueStatus) => void
    onkeymove: (issue: Issue, direction: -1 | 1) => void
    ondragstart: (issue: Issue, event: DragEvent) => void
    ondragend: () => void
  } = $props()

  let menuOpen = $state(false)
  let summary = $state<HTMLElement>()

  function choose(status: IssueStatus) {
    menuOpen = false
    if (status !== issue.status) onmove(issue, status)
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.target !== event.currentTarget) return
    if (event.altKey && event.key === 'ArrowLeft') {
      event.preventDefault()
      onkeymove(issue, -1)
      return
    }
    if (event.altKey && event.key === 'ArrowRight') {
      event.preventDefault()
      onkeymove(issue, 1)
      return
    }
    if (event.key === 'Enter') {
      event.preventDefault()
      onopen()
    }
  }

  function handleMenuKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && menuOpen) {
      event.preventDefault()
      menuOpen = false
      summary?.focus()
    }
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<li
  data-issue-id={issue.id}
  tabindex="0"
  draggable={!saving}
  aria-busy={saving}
  aria-describedby={`${menuId}-instructions`}
  class={cn(
    'group/card relative list-none rounded-xl border border-border bg-[var(--surface-1)] p-3 outline-none transition-colors',
    'hover:border-[var(--border-strong)] hover:bg-[var(--surface-2)] focus-visible:ring-2 focus-visible:ring-ring',
    saving ? 'cursor-wait opacity-70' : 'cursor-grab active:cursor-grabbing',
  )}
  ondragstart={(event) => ondragstart(issue, event)}
  ondragend={ondragend}
  onkeydown={handleKeydown}
>
  <div class="flex items-start justify-between gap-2">
    <IssueId identifier={issue.identifier} />
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <details
      bind:open={menuOpen}
      class="relative"
      onkeydown={handleMenuKeydown}
    >
      <summary
        bind:this={summary}
        aria-label={`Move ${issue.identifier}`}
        class="grid size-10 cursor-pointer list-none place-items-center rounded-lg md:size-8 text-[var(--text-3)] outline-none transition-opacity hover:bg-[var(--surface-3)] hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring group-hover/card:opacity-100 group-focus-within/card:opacity-100 md:opacity-0 [@media(pointer:coarse)]:opacity-100"
      >
        {#if saving}
          <LoaderCircle size={15} strokeWidth={1.6} class="animate-spin motion-reduce:animate-none" aria-hidden="true" />
        {:else}
          <GripVertical size={15} strokeWidth={1.6} aria-hidden="true" />
        {/if}
      </summary>
      <div
        class="absolute right-0 z-20 mt-1 w-48 rounded-xl border border-[var(--border-strong)] bg-[var(--bg-2)] p-1 shadow-[var(--shadow-lg)]"
        role="group"
        aria-label={`Status for ${issue.identifier}`}
      >
        {#each ISSUE_STATUS_OPTIONS as option (option.value)}
          {@const Icon = option.icon}
          <button
            type="button"
            class="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-xs text-[var(--text-2)] transition-colors hover:bg-[var(--surface-2)] hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            onclick={() => choose(option.value)}
          >
            {#if Icon}
              <Icon size={14} strokeWidth={1.6} style="color: {option.tone}" aria-hidden="true" />
            {/if}
            <span class="min-w-0 flex-1 truncate">{option.label}</span>
            {#if option.value === issue.status}
              <Check size={14} strokeWidth={1.8} class="text-primary" aria-hidden="true" />
              <span class="sr-only">Current status</span>
            {/if}
          </button>
        {/each}
      </div>
    </details>
  </div>

  <a
    {href}
    class="mt-2 block cursor-pointer rounded text-sm font-medium leading-snug text-foreground outline-none hover:text-primary focus-visible:ring-2 focus-visible:ring-ring"
    onclick={(event) => {
      event.preventDefault()
      onopen()
    }}
  >
    {issue.title}
  </a>

  <div class="mt-2.5 flex flex-wrap items-center gap-x-2 gap-y-1">
    <PriorityBadge priority={issue.priority} />
    {#if issue.agent_id}
      <AgentChip agentId={issue.agent_id} />
    {:else if epicTitle}
      <span class="truncate text-[11px] text-[var(--text-3)]">{epicTitle}</span>
    {/if}
  </div>

  <p id={`${menuId}-instructions`} class="sr-only">
    Press Alt plus left or right arrow to move across board columns. Use the move menu for
    any status.
  </p>
</li>
