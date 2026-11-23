<script lang="ts">
  import Search from '@lucide/svelte/icons/search'
  import AgentChip from './AgentChip.svelte'
  import Button from './Button.svelte'
  import IdChip from './IdChip.svelte'
  import PriorityBadge from './PriorityBadge.svelte'
  import Select from './Select.svelte'
  import StatusCell from './StatusCell.svelte'
  import { issuePath, navigate } from '$lib/router.svelte'
  import { ANY_PRIORITY_OPTION, ANY_STATUS_OPTION, ISSUE_STATUS_OPTIONS, PRIORITY_OPTIONS } from '$lib/meta'
  import { inputClass } from '$lib/ui'
  import { cn } from '$lib/utils'
  import type { Epic, Issue } from '$lib/types'

  let {
    projectRef,
    issues,
    epics,
    loading,
    error,
    status = $bindable(''),
    priority = $bindable(''),
    search = $bindable(''),
    onreload,
    onclear,
  }: {
    projectRef: string
    issues: Issue[]
    epics: Epic[]
    loading: boolean
    error: string
    status?: string
    priority?: string
    search?: string
    onreload: () => void
    onclear: () => void
  } = $props()

  const epicMap = $derived(new Map(epics.map((epic) => [epic.id, epic.title])))
  const hasFilters = $derived(Boolean(status || priority || search.trim()))

  function epicTitle(id: string | null): string {
    if (!id) return '—'
    return epicMap.get(id) ?? 'Unknown epic'
  }
</script>

<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
  <div class="flex flex-wrap items-end gap-2">
    <label class="flex flex-col gap-1 text-[11px] text-[var(--text-3)]">
      Status
      <Select
        value={status}
        options={[ANY_STATUS_OPTION, ...ISSUE_STATUS_OPTIONS]}
        class="min-w-40"
        onchange={(value) => {
          status = value
          onreload()
        }}
      />
    </label>

    <label class="flex flex-col gap-1 text-[11px] text-[var(--text-3)]">
      Priority
      <Select
        value={priority}
        options={[ANY_PRIORITY_OPTION, ...PRIORITY_OPTIONS]}
        class="min-w-36"
        onchange={(value) => {
          priority = value
          onreload()
        }}
      />
    </label>

    <form
      class="flex items-end gap-2"
      onsubmit={(event) => {
        event.preventDefault()
        onreload()
      }}
    >
      <label class="flex flex-col gap-1 text-[11px] text-[var(--text-3)]">
        Search
        <input
          class={cn(inputClass, 'w-44 text-xs sm:w-56')}
          bind:value={search}
          placeholder="Title or description"
          autocomplete="off"
        />
      </label>
      <Button variant="secondary" type="submit" class="px-2.5 text-xs">
        <Search size={14} strokeWidth={1.6} aria-hidden="true" />
        Search
      </Button>
    </form>

    {#if hasFilters}
      <Button variant="ghost" class="px-2 text-xs" onclick={onclear}>Clear</Button>
    {/if}
  </div>

  <p class="text-[11px] text-[var(--text-3)] tnum">{issues.length} shown</p>
</div>

{#if error}
  <div class="mt-4 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
    {error}
  </div>
{/if}

{#if loading}
  <div class="mt-4 space-y-2" aria-hidden="true">
    {#each Array(4) as _}
      <div class="h-12 animate-pulse rounded-xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none"></div>
    {/each}
  </div>
{:else if issues.length === 0}
  <div class="mt-4 grid place-items-center rounded-2xl border border-border bg-[var(--surface-1)] px-6 py-12 text-center">
    <h3 class="text-sm font-semibold">{hasFilters ? 'No issues match these filters' : 'No issues yet'}</h3>
    <p class="mt-1 max-w-sm text-xs leading-relaxed text-[var(--text-2)]">
      {hasFilters ? 'Try clearing the filters or searching for something else.' : 'Create the first issue, or let an agent create one through the API.'}
    </p>
    {#if hasFilters}
      <Button variant="secondary" class="mt-4 text-xs" onclick={onclear}>Clear filters</Button>
    {/if}
  </div>
{:else}
  <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
  <div class="glass mt-4 hidden overflow-hidden rounded-2xl sm:block" role="region" tabindex="0" aria-label="Issue list">
    <table class="w-full text-left text-sm">
      <caption class="sr-only">Issues in this project</caption>
      <thead class="border-b border-border text-[11px] font-medium text-[var(--text-3)]">
        <tr>
          <th scope="col" class="px-3 py-2.5">ID</th>
          <th scope="col" class="w-36 px-3 py-2.5">Status</th>
          <th scope="col" class="w-28 px-3 py-2.5">Priority</th>
          <th scope="col" class="px-3 py-2.5">Issue</th>
          <th scope="col" class="hidden px-3 py-2.5 lg:table-cell">Epic</th>
          <th scope="col" class="hidden px-3 py-2.5 xl:table-cell">Agent</th>
        </tr>
      </thead>
      <tbody>
        {#each issues as issue (issue.id)}
          {@const href = issuePath(projectRef, issue.identifier)}
          <tr class="border-t border-border/60 transition-colors hover:bg-white/[.03]">
            <th scope="row" class="whitespace-nowrap px-3 py-2.5 text-left font-normal">
              <a
                {href}
                class="cursor-pointer rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
                onclick={(event) => {
                  event.preventDefault()
                  navigate(href)
                }}
              >
                <IdChip identifier={issue.identifier} />
              </a>
            </th>
            <td class="px-3 py-2.5"><StatusCell status={issue.status} /></td>
            <td class="whitespace-nowrap px-3 py-2.5"><PriorityBadge priority={issue.priority} /></td>
            <td class="max-w-[34rem] px-3 py-2.5">
              <a
                {href}
                class="block cursor-pointer truncate rounded text-sm font-medium text-foreground outline-none hover:text-primary focus-visible:ring-2 focus-visible:ring-ring"
                onclick={(event) => {
                  event.preventDefault()
                  navigate(href)
                }}
              >
                {issue.title}
              </a>
            </td>
            <td class="hidden max-w-48 truncate px-3 py-2.5 text-xs text-[var(--text-2)] lg:table-cell">{epicTitle(issue.epic_id)}</td>
            <td class="hidden whitespace-nowrap px-3 py-2.5 xl:table-cell">
              {#if issue.agent_id}
                <AgentChip agentId={issue.agent_id} />
              {:else}
                <span class="text-xs text-[var(--text-3)]">—</span>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <div class="mt-4 space-y-2 sm:hidden">
    {#each issues as issue (issue.id)}
      {@const href = issuePath(projectRef, issue.identifier)}
      <a
        {href}
        class="block cursor-pointer rounded-xl border border-border bg-[var(--surface-1)] p-3 outline-none transition-colors hover:bg-[var(--surface-2)] focus-visible:ring-2 focus-visible:ring-ring"
        onclick={(event) => {
          event.preventDefault()
          navigate(href)
        }}
      >
        <div class="flex flex-wrap items-center gap-1.5">
          <IdChip identifier={issue.identifier} />
          <StatusCell status={issue.status} class="w-auto" />
          <PriorityBadge priority={issue.priority} />
        </div>
        <h3 class="mt-2 text-sm font-medium leading-snug">{issue.title}</h3>
        <div class="mt-2 flex flex-wrap items-center gap-2 text-[11px] text-[var(--text-3)]">
          <span class="truncate">{epicTitle(issue.epic_id)}</span>
          {#if issue.agent_id}
            <AgentChip agentId={issue.agent_id} />
          {/if}
        </div>
      </a>
    {/each}
  </div>
{/if}
