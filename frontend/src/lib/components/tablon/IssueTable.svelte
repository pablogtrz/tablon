<script lang="ts">
  import AgentChip from './AgentChip.svelte'
  import IssueId from './IssueId.svelte'
  import PriorityBadge from './PriorityBadge.svelte'
  import StatusCell from './StatusCell.svelte'
  import { issuePath, navigate } from '$lib/router.svelte'
  import type { Epic, Issue } from '$lib/types'

  let {
    projectRef,
    issues,
    epics,
    caption,
    showHeaders = true,
    showEpicColumn = true,
    showAgentColumn = true,
  }: {
    projectRef: string
    issues: Issue[]
    epics: Epic[]
    caption: string
    showHeaders?: boolean
    showEpicColumn?: boolean
    showAgentColumn?: boolean
  } = $props()

  const epicMap = $derived(new Map(epics.map((epic) => [epic.id, epic.title])))

  function epicTitle(id: string | null): string {
    if (!id) return '—'
    return epicMap.get(id) ?? 'Unknown epic'
  }

  function open(event: MouseEvent, href: string) {
    event.preventDefault()
    void navigate(href)
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div class="hidden sm:block" role="region" tabindex="0" aria-label={caption}>
  <table class="w-full border-collapse text-left text-sm">
    <caption class="sr-only">{caption}</caption>
    {#if showHeaders}
      <thead class="border-b border-border text-[11px] font-medium text-[var(--text-3)]">
        <tr>
          <th scope="col" class="w-28 px-3 py-2.5">ID</th>
          <th scope="col" class="w-36 px-3 py-2.5">Status</th>
          <th scope="col" class="w-32 px-3 py-2.5">Priority</th>
          <th scope="col" class="px-3 py-2.5">Issue</th>
          {#if showEpicColumn}
            <th scope="col" class="hidden px-3 py-2.5 lg:table-cell">Epic</th>
          {/if}
          {#if showAgentColumn}
            <th scope="col" class="hidden px-3 py-2.5 xl:table-cell">Agent</th>
          {/if}
        </tr>
      </thead>
    {/if}
    <tbody>
      {#each issues as issue (issue.id)}
        {@const href = issuePath(projectRef, issue.identifier)}
        <tr class="border-t border-white/[.045] transition-colors hover:bg-white/[.035]">
          <th scope="row" class="whitespace-nowrap px-3 py-2.5 text-left font-normal">
            <a
              {href}
              class="cursor-pointer rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
              onclick={(event) => open(event, href)}
            >
              <IssueId identifier={issue.identifier} />
            </a>
          </th>
          <td class="px-3 py-2.5"><StatusCell status={issue.status} /></td>
          <td class="whitespace-nowrap px-3 py-2.5"><PriorityBadge priority={issue.priority} /></td>
          <td class="max-w-[34rem] px-3 py-2.5">
            <a
              {href}
              class="block cursor-pointer truncate rounded text-sm font-medium text-foreground outline-none hover:text-primary focus-visible:ring-2 focus-visible:ring-ring"
              onclick={(event) => open(event, href)}
            >
              {issue.title}
            </a>
          </td>
          {#if showEpicColumn}
            <td class="hidden max-w-48 truncate px-3 py-2.5 text-xs text-[var(--text-2)] lg:table-cell">
              {epicTitle(issue.epic_id)}
            </td>
          {/if}
          {#if showAgentColumn}
            <td class="hidden whitespace-nowrap px-3 py-2.5 xl:table-cell">
              {#if issue.agent_id}
                <AgentChip agentId={issue.agent_id} />
              {:else}
                <span class="text-xs text-[var(--text-3)]">—</span>
              {/if}
            </td>
          {/if}
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<ul class="divide-y divide-white/[.045] sm:hidden" aria-label={caption}>
  {#each issues as issue (issue.id)}
    {@const href = issuePath(projectRef, issue.identifier)}
    <li>
      <a
        {href}
        class="block cursor-pointer px-1 py-3 outline-none transition-colors hover:bg-white/[.03] focus-visible:ring-2 focus-visible:ring-ring"
        onclick={(event) => open(event, href)}
      >
        <div class="flex flex-wrap items-center gap-1.5">
          <IssueId identifier={issue.identifier} />
          <StatusCell status={issue.status} class="w-auto" />
          <PriorityBadge priority={issue.priority} />
        </div>
        <h3 class="mt-2 text-sm font-medium leading-snug">{issue.title}</h3>
        {#if (showEpicColumn && issue.epic_id) || (showAgentColumn && issue.agent_id)}
          <div class="mt-2 flex flex-wrap items-center gap-2 text-[11px] text-[var(--text-3)]">
            {#if showEpicColumn && issue.epic_id}
              <span class="truncate">{epicTitle(issue.epic_id)}</span>
            {/if}
            {#if showAgentColumn && issue.agent_id}
              <AgentChip agentId={issue.agent_id} />
            {/if}
          </div>
        {/if}
      </a>
    </li>
  {/each}
</ul>
