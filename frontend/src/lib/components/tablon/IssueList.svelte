<script lang="ts">
  import { onMount } from 'svelte'
  import Layers from '@lucide/svelte/icons/layers'
  import Search from '@lucide/svelte/icons/search'
  import AgentChip from './AgentChip.svelte'
  import Button from './Button.svelte'
  import IssueTable from './IssueTable.svelte'
  import Select from './Select.svelte'
  import { ANY_PRIORITY_OPTION, ANY_STATUS_OPTION, ISSUE_STATUS_META, ISSUE_STATUS_OPTIONS, PRIORITY_OPTIONS } from '$lib/meta'
  import { groupIssues, isListGroupBy, LIST_GROUP_BY_LABELS, LIST_GROUP_BY_OPTIONS } from '$lib/issue-grouping'
  import type { ListGroupBy } from '$lib/issue-grouping'
  import { readPreference, writePreference } from '$lib/preferences'
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

  let groupBy = $state<ListGroupBy>('none')

  const hasFilters = $derived(Boolean(status || priority || search.trim()))
  const groups = $derived(groupIssues(issues, groupBy, epics))
  const grouped = $derived(groupBy !== 'none')
  const showEpicColumn = $derived(groupBy !== 'epic')
  const showAgentColumn = $derived(groupBy !== 'agent')
  const groupOptions = LIST_GROUP_BY_OPTIONS.map((value) => ({
    value,
    label: LIST_GROUP_BY_LABELS[value],
  }))

  function changeGroup(value: string) {
    if (!isListGroupBy(value)) return
    groupBy = value
    writePreference('list.groupBy', value)
  }

  onMount(() => {
    const saved = readPreference('list.groupBy')
    if (isListGroupBy(saved)) groupBy = saved
  })
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

    <label class="flex flex-col gap-1 text-[11px] text-[var(--text-3)]">
      Group by
      <Select
        value={groupBy}
        options={groupOptions}
        class="min-w-32"
        onchange={changeGroup}
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
      <div class="h-11 animate-pulse rounded-lg bg-[var(--surface-1)] motion-reduce:animate-none"></div>
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
  <div class="mt-4">
    {#if !grouped}
      <IssueTable
        {projectRef}
        {issues}
        {epics}
        caption="Issues in this project"
        showHeaders
        {showEpicColumn}
        {showAgentColumn}
      />
    {:else}
      <div class="space-y-6">
        {#each groups as group, index (group.key)}
          {@const statusMeta = group.kind === 'status' ? ISSUE_STATUS_META[group.key as keyof typeof ISSUE_STATUS_META] : null}
          <section aria-labelledby={`issue-group-${index}`}>
            <div class="flex items-center gap-2 border-b border-border py-2">
              {#if statusMeta}
                {@const Icon = statusMeta.icon}
                <Icon size={15} strokeWidth={1.7} style="color: {statusMeta.tone}" aria-hidden="true" />
              {:else if group.kind === 'epic'}
                <Layers size={15} strokeWidth={1.7} class="text-[var(--text-3)]" aria-hidden="true" />
              {:else if group.kind === 'agent'}
                <span class="size-1.5 rounded-full bg-agent" aria-hidden="true"></span>
              {/if}
              <h3 id={`issue-group-${index}`} class="text-xs font-semibold text-foreground">
                {#if group.kind === 'agent'}
                  <AgentChip agentId={group.key} class="h-6 text-[11px]" />
                {:else}
                  {group.label}
                {/if}
              </h3>
              <span class="tnum rounded-full bg-white/[.06] px-1.5 py-0.5 text-[10px] text-[var(--text-3)]">
                {group.issues.length}
              </span>
            </div>
            <IssueTable
              {projectRef}
              issues={group.issues}
              {epics}
              caption={`${group.label} issues`}
              showHeaders={false}
              {showEpicColumn}
              {showAgentColumn}
            />
          </section>
        {/each}
      </div>
    {/if}
  </div>
{/if}
