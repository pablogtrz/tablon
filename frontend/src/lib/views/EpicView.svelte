<script lang="ts">
  import { onMount } from 'svelte'
  import ArrowLeft from '@lucide/svelte/icons/arrow-left'
  import Pencil from '@lucide/svelte/icons/pencil'
  import Trash2 from '@lucide/svelte/icons/trash-2'
  import Button from '$lib/components/tablon/Button.svelte'
  import ConfirmModal from '$lib/components/tablon/ConfirmModal.svelte'
  import EpicFormModal from '$lib/components/tablon/EpicFormModal.svelte'
  import EpicStatusBadge from '$lib/components/tablon/EpicStatusBadge.svelte'
  import IssueStatusCounts from '$lib/components/tablon/IssueStatusCounts.svelte'
  import IssueTable from '$lib/components/tablon/IssueTable.svelte'
  import MarkdownRenderer from '$lib/components/tablon/MarkdownRenderer.svelte'
  import { deleteEpic, getEpic, getProject, listAllIssues } from '$lib/api'
  import { messageOf } from '$lib/errors'
  import { plural, relativeTime } from '$lib/format'
  import { navigate, projectPath } from '$lib/router.svelte'
  import { ISSUE_STATUSES } from '$lib/types'
  import type { Epic, Issue, IssueStatus, Project } from '$lib/types'

  let { projectRef, epicRef }: { projectRef: string; epicRef: string } = $props()

  let project = $state<Project | null>(null)
  let epic = $state<Epic | null>(null)
  let issues = $state<Issue[]>([])
  let loading = $state(true)
  let error = $state('')
  let truncated = $state(false)
  let heading = $state<HTMLHeadingElement>()
  let focusPending = true

  let editOpen = $state(false)
  let deleteOpen = $state(false)
  let deleting = $state(false)
  let deleteError = $state('')

  const counts = $derived.by(() => {
    const result = Object.fromEntries(ISSUE_STATUSES.map((status) => [status, 0])) as Record<IssueStatus, number>
    for (const issue of issues) result[issue.status] += 1
    return result
  })
  const total = $derived(ISSUE_STATUSES.reduce((sum, status) => sum + counts[status], 0))
  const done = $derived(counts.done)
  const progress = $derived(total === 0 ? 0 : Math.round((done / total) * 100))

  $effect(() => {
    if (!focusPending || !heading) return
    focusPending = false
    heading.focus({ preventScroll: true })
  })

  async function load() {
    loading = true
    error = ''
    try {
      const [loadedProject, loadedEpic] = await Promise.all([
        getProject(projectRef),
        getEpic(epicRef),
      ])
      const result = await listAllIssues(projectRef, { epic: loadedEpic.id })
      project = loadedProject
      epic = loadedEpic
      issues = result.items
      truncated = result.truncated
    } catch (err) {
      error = messageOf(err)
    } finally {
      loading = false
    }
  }

  function backToProject() {
    void navigate(projectPath(projectRef))
  }

  function onEpicSaved(saved: Epic) {
    epic = saved
  }

  function openDelete() {
    deleteError = ''
    deleteOpen = true
  }

  async function submitDelete() {
    const target = epic
    if (!target || deleting) return
    deleting = true
    deleteError = ''
    try {
      await deleteEpic(target.id)
      deleteOpen = false
      backToProject()
    } catch (err) {
      deleteError = messageOf(err)
    } finally {
      deleting = false
    }
  }

  onMount(() => {
    void load()
  })
</script>

<svelte:head>
  <title>{epic?.title ?? 'Epic'} · Tablon</title>
</svelte:head>

{#if loading}
  <div class="space-y-5" aria-hidden="true">
    <div class="h-4 w-24 animate-pulse rounded bg-white/[.06] motion-reduce:animate-none"></div>
    <div class="glass h-36 animate-pulse rounded-2xl motion-reduce:animate-none"></div>
    <div class="h-72 animate-pulse rounded-2xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none"></div>
  </div>
{:else if !epic}
  <div class="glass mx-auto mt-10 max-w-lg rounded-2xl p-6 text-center">
    <h1 class="text-base font-semibold">Could not load this epic</h1>
    <p class="mt-2 text-sm leading-relaxed text-[var(--text-2)]">{error || 'The epic does not exist or the API is unavailable.'}</p>
    <div class="mt-5 flex justify-center gap-2">
      <Button variant="secondary" onclick={backToProject}>Back to project</Button>
      <Button onclick={() => void load()}>Retry</Button>
    </div>
  </div>
{:else}
  <div class="space-y-5">
    <a
      href={projectPath(projectRef)}
      class="inline-flex items-center gap-1.5 rounded text-xs text-[var(--text-2)] outline-none transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
      onclick={(event) => {
        event.preventDefault()
        backToProject()
      }}
    >
      <ArrowLeft size={14} strokeWidth={1.6} aria-hidden="true" />
      {project?.name ?? 'Project'}
    </a>

    <header class="rounded-2xl border border-border bg-[var(--surface-1)] p-4 sm:p-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="min-w-0">
          <h1 bind:this={heading} tabindex="-1" data-route-focus class="text-xl font-semibold tracking-[-.02em] outline-none">
            {epic.title}
          </h1>
          <p class="mt-1 text-[11px] text-[var(--text-3)]">
            Updated {relativeTime(epic.updated_at)} · {plural(total, 'task')}
          </p>
        </div>
        <div class="flex flex-wrap items-center justify-end gap-2">
          <EpicStatusBadge status={epic.status} />
          <Button variant="secondary" class="h-8 px-2 text-xs" onclick={() => (editOpen = true)}>
            <Pencil size={13} strokeWidth={1.6} aria-hidden="true" />
            Edit
          </Button>
          <Button
            variant="ghost"
            class="h-8 px-2 text-xs text-destructive hover:text-destructive"
            onclick={openDelete}
          >
            <Trash2 size={13} strokeWidth={1.6} aria-hidden="true" />
            Delete
          </Button>
        </div>
      </div>

      <div class="mt-4">
        <div class="flex items-baseline justify-between gap-3">
          <p class="text-xs text-[var(--text-2)]">
            <span class="tnum text-sm font-semibold text-foreground">{done}/{total}</span>
            <span class="ml-0.5">done</span>
          </p>
          <span class="tnum text-[11px] text-[var(--text-3)]">{progress}%</span>
        </div>
        <progress
          class="tablon-progress mt-2.5"
          value={done}
          max={Math.max(total, 1)}
          style="--progress-fill: {epic.status === 'done' ? 'var(--status-done)' : 'var(--brand)'}"
          aria-label={`${done} of ${total} tasks done in ${epic.title}`}
        ></progress>
        <IssueStatusCounts {counts} label="Tasks by status" class="mt-4" />
      </div>

      <div class="mt-5 border-t border-border pt-4">
        <h2 class="text-xs font-medium text-[var(--text-3)]">Description</h2>
        {#if epic.description.trim()}
          <div class="mt-2">
            <MarkdownRenderer source={epic.description} />
          </div>
        {:else}
          <p class="mt-2 text-sm text-[var(--text-3)]">No description yet.</p>
        {/if}
      </div>
    </header>

    {#if truncated}
      <p class="rounded-lg border border-border bg-[var(--surface-1)] px-3 py-2 text-[11px] text-[var(--text-3)]">
        Showing the first 1,000 tasks in this epic.
      </p>
    {/if}

    {#if issues.length === 0}
      <div class="grid place-items-center rounded-2xl border border-border bg-[var(--surface-1)] px-6 py-12 text-center">
        <h2 class="text-sm font-semibold">No tasks in this epic yet</h2>
        <p class="mt-1 max-w-sm text-xs leading-relaxed text-[var(--text-2)]">
          Assign issues to this epic from the project list or the issue detail view.
        </p>
        <Button variant="secondary" class="mt-4 text-xs" onclick={backToProject}>Browse issues</Button>
      </div>
    {:else}
      <IssueTable
        {projectRef}
        {issues}
        epics={[epic]}
        caption={`Tasks in ${epic.title}`}
        showHeaders
        showEpicColumn={false}
      />
    {/if}
  </div>
{/if}

{#if epic}
  <EpicFormModal bind:open={editOpen} {projectRef} {epic} onsaved={onEpicSaved} />
  <ConfirmModal
    bind:open={deleteOpen}
    title="Delete epic"
    description={`Delete “${epic.title}”? Its issues will become unassigned.`}
    confirmLabel="Delete epic"
    busy={deleting}
    error={deleteError}
    onconfirm={() => void submitDelete()}
  />
{/if}
