<script lang="ts">
  import { onMount, tick } from 'svelte'
  import Pencil from '@lucide/svelte/icons/pencil'
  import Plus from '@lucide/svelte/icons/plus'
  import Trash2 from '@lucide/svelte/icons/trash-2'
  import { fade } from 'svelte/transition'
  import { motion } from '$lib/motion.svelte'
  import Button from '$lib/components/tablon/Button.svelte'
  import ConfirmModal from '$lib/components/tablon/ConfirmModal.svelte'
  import EpicFormModal from '$lib/components/tablon/EpicFormModal.svelte'
  import EpicList from '$lib/components/tablon/EpicList.svelte'
  import IdChip from '$lib/components/tablon/IdChip.svelte'
  import IssueBoard from '$lib/components/tablon/IssueBoard.svelte'
  import IssueCreateModal from '$lib/components/tablon/IssueCreateModal.svelte'
  import IssueList from '$lib/components/tablon/IssueList.svelte'
  import ProjectLogo from '$lib/components/tablon/ProjectLogo.svelte'
  import ProjectSettingsModal from '$lib/components/tablon/ProjectSettingsModal.svelte'
  import ProjectStatusBadge from '$lib/components/tablon/ProjectStatusBadge.svelte'
  import { deleteEpic, deleteProject, getProject, listAllIssues, listEpics, listIssues, updateIssueStatus } from '$lib/api'
  import { applyBoardMove } from '$lib/board'
  import { burstConfetti } from '$lib/confetti'
  import { subscribeProjectEvents } from '$lib/events'
  import type { ProjectEvent } from '$lib/events'
  import type { HiddenBoardStatus } from '$lib/board'
  import { messageOf } from '$lib/errors'
  import { relativeTime } from '$lib/format'
  import { issuePath, epicPath, navigate } from '$lib/router.svelte'
  import type { Epic, Issue, IssueStatus, Project } from '$lib/types'

  let { projectRef }: { projectRef: string } = $props()

  let project = $state<Project | null>(null)
  let epics = $state<Epic[]>([])
  let issues = $state<Issue[]>([])
  let loading = $state(true)
  let error = $state('')
  let heading = $state<HTMLHeadingElement>()
  let focusPending = true

  let tab = $state<'issues' | 'board' | 'epics'>('issues')
  let issuesLoading = $state(false)
  let issuesError = $state('')

  let boardIssues = $state<Issue[]>([])
  let boardLoading = $state(false)
  let boardError = $state('')
  let boardTruncated = $state(false)
  let movingIssueIds = $state<string[]>([])
  let liveMessage = $state('')
  let eventTimer: ReturnType<typeof setTimeout> | null = null

  let epicIssues = $state<Issue[]>([])
  let epicIssuesLoading = $state(false)
  let epicIssuesError = $state('')
  let epicIssuesTruncated = $state(false)

  let statusFilter = $state('')
  let priorityFilter = $state('')
  let search = $state('')

  let projectSettingsOpen = $state(false)
  let issueOpen = $state(false)
  let epicOpen = $state(false)
  let editingEpic = $state<Epic | null>(null)

  let deleteProjectOpen = $state(false)
  let deleteEpicOpen = $state(false)
  let deletingEpic = $state<Epic | null>(null)
  let deleting = $state(false)
  let deleteError = $state('')

  const apiRef = $derived(project?.identifier ?? projectRef)

  $effect(() => {
    if (!focusPending || !heading) return
    focusPending = false
    heading.focus({ preventScroll: true })
  })

  async function loadAll() {
    loading = true
    error = ''
    try {
      const [loadedProject, loadedEpics, issuePage] = await Promise.all([
        getProject(projectRef),
        listEpics(projectRef),
        listIssues(projectRef, { limit: 100, sort: 'created_at', order: 'desc' }),
      ])
      project = loadedProject
      epics = loadedEpics
      issues = issuePage.items
    } catch (err) {
      error = messageOf(err)
    } finally {
      loading = false
    }
  }

  async function loadIssues() {
    issuesLoading = true
    issuesError = ''
    try {
      const page = await listIssues(apiRef, {
        status: statusFilter || undefined,
        priority: priorityFilter || undefined,
        q: search.trim() || undefined,
        limit: 100,
        sort: 'created_at',
        order: 'desc',
      })
      issues = page.items
    } catch (err) {
      issuesError = messageOf(err)
    } finally {
      issuesLoading = false
    }
  }

  async function loadBoard() {
    boardLoading = true
    boardError = ''
    try {
      const result = await listAllIssues(apiRef)
      boardIssues = result.items
      boardTruncated = result.truncated
    } catch (err) {
      boardError = messageOf(err)
    } finally {
      boardLoading = false
    }
  }

  async function loadEpicIssues() {
    epicIssuesLoading = true
    epicIssuesError = ''
    try {
      const result = await listAllIssues(apiRef)
      epicIssues = result.items
      epicIssuesTruncated = result.truncated
    } catch (err) {
      epicIssuesError = messageOf(err)
    } finally {
      epicIssuesLoading = false
    }
  }

  async function moveBoardIssue(issue: Issue, status: IssueStatus): Promise<boolean> {
    if (movingIssueIds.includes(issue.id)) return false
    const previous = boardIssues
    const previousIssue = boardIssues.find((item) => item.id === issue.id) ?? issue
    const optimistic = applyBoardMove(boardIssues, issue.id, status)
    if (optimistic === previous) return true

    boardIssues = optimistic
    boardError = ''
    movingIssueIds = [...movingIssueIds, issue.id]
    try {
      const updated = await updateIssueStatus(issue.identifier, status)
      boardIssues = boardIssues.map((item) => (item.id === updated.id ? updated : item))
      if (status === 'done' && previousIssue.status !== 'done') {
        burstConfetti()
      }
      return true
    } catch (err) {
      boardIssues = boardIssues.map((item) => (item.id === issue.id ? previousIssue : item))
      boardError = messageOf(err)
      return false
    } finally {
      movingIssueIds = movingIssueIds.filter((id) => id !== issue.id)
    }
  }

  function selectTab(next: 'issues' | 'board' | 'epics') {
    tab = next
    if (next === 'board') {
      void loadBoard()
    } else if (next === 'issues') {
      void loadIssues()
    } else {
      void loadEpicIssues()
    }
  }

  function showHidden(status: HiddenBoardStatus) {
    statusFilter = status
    tab = 'issues'
    void loadIssues()
  }

  function openBoardIssue(issue: Issue) {
    navigate(issuePath(apiRef, issue.identifier))
  }

  function scheduleRefresh() {
    if (eventTimer) clearTimeout(eventTimer)
    eventTimer = setTimeout(() => {
      eventTimer = null
      if (tab === 'board') {
        void loadBoard()
      } else if (tab === 'epics') {
        void loadEpicIssues()
      } else {
        void loadIssues()
      }
    }, 150)
  }

  async function announceEvent(event: ProjectEvent) {
    if (typeof document !== 'undefined' && document.hidden) return
    const label = event.identifier ?? 'issue'
    let message = ''
    switch (event.type) {
      case 'issue.created':
        message = `Issue ${label} created`
        break
      case 'issue.updated':
        message = `Issue ${label} updated`
        break
      case 'issue.deleted':
        message = `Issue ${label} deleted`
        break
      case 'issue.claimed':
        message = `Issue ${label} claimed`
        break
      case 'issue.released':
        message = `Issue ${label} released`
        break
      case 'comment.created':
        message = `New comment on ${label}`
        break
      case 'comment.updated':
        message = `Comment updated on ${label}`
        break
      case 'comment.deleted':
        message = `Comment deleted on ${label}`
        break
    }
    if (!message) return
    liveMessage = ''
    await tick()
    liveMessage = message
  }

  $effect(() => {
    if (!project) return
    const dispose = subscribeProjectEvents(project.identifier, {
      onsync: scheduleRefresh,
      onEvent: (event) => {
        void announceEvent(event)
        scheduleRefresh()
      },
    })
    return () => {
      dispose()
      if (eventTimer) {
        clearTimeout(eventTimer)
        eventTimer = null
      }
    }
  })

  async function reloadEpics() {
    try {
      epics = await listEpics(apiRef)
    } catch (err) {
      error = messageOf(err)
    }
  }

  function openNewIssue() {
    issueOpen = true
  }

  function openNewEpic() {
    editingEpic = null
    epicOpen = true
  }

  function openEditEpic(epic: Epic) {
    editingEpic = epic
    epicOpen = true
  }

  function openDeleteProject() {
    deleteError = ''
    deleteProjectOpen = true
  }

  function openDeleteEpic(epic: Epic) {
    deleteError = ''
    deletingEpic = epic
    deleteEpicOpen = true
  }

  function clearFilters() {
    statusFilter = ''
    priorityFilter = ''
    search = ''
    void loadIssues()
  }

  async function submitDeleteProject() {
    if (!project || deleting) return
    deleting = true
    deleteError = ''
    try {
      await deleteProject(project.identifier)
      navigate('/')
    } catch (err) {
      deleteError = messageOf(err)
    } finally {
      deleting = false
    }
  }

  async function submitDeleteEpic() {
    if (!deletingEpic || deleting) return
    deleting = true
    deleteError = ''
    try {
      await deleteEpic(deletingEpic.id)
      deleteEpicOpen = false
      deletingEpic = null
      await Promise.all([reloadEpics(), loadIssues(), loadEpicIssues()])
    } catch (err) {
      deleteError = messageOf(err)
    } finally {
      deleting = false
    }
  }

  function onIssueCreated(issue: Issue) {
    navigate(issuePath(apiRef, issue.identifier))
  }

  function onProjectSaved(updated: Project) {
    project = updated
  }

  function onEpicSaved() {
    void reloadEpics()
    if (tab === 'epics') void loadEpicIssues()
  }

  function openEpic(epic: Epic) {
    void navigate(epicPath(apiRef, epic.id))
  }

  onMount(() => {
    void loadAll()
  })
</script>

<svelte:head>
  <title>{project?.name ?? 'Project'} · Tablon</title>
</svelte:head>

{#if loading && !project}
  <div class="space-y-5" aria-hidden="true">
    <div class="h-4 w-24 animate-pulse rounded bg-white/[.06] motion-reduce:animate-none"></div>
    <div class="flex gap-4">
      <div class="size-12 animate-pulse rounded-xl bg-white/[.06] motion-reduce:animate-none"></div>
      <div class="flex-1 space-y-2 py-1">
        <div class="h-5 w-56 animate-pulse rounded bg-white/[.06] motion-reduce:animate-none"></div>
        <div class="h-3 w-80 animate-pulse rounded bg-white/[.04] motion-reduce:animate-none"></div>
      </div>
    </div>
    <div class="h-10 w-64 animate-pulse rounded-xl bg-white/[.04] motion-reduce:animate-none"></div>
    <div class="h-72 animate-pulse rounded-2xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none"></div>
  </div>
{:else if !project}
  <div class="glass mx-auto mt-10 max-w-lg rounded-2xl p-6 text-center">
    <h1 class="text-base font-semibold">Could not load this project</h1>
    <p class="mt-2 text-sm leading-relaxed text-[var(--text-2)]">{error || 'The project does not exist or the API is unavailable.'}</p>
    <div class="mt-5 flex justify-center gap-2">
      <Button variant="secondary" onclick={() => navigate('/')}>All projects</Button>
      <Button onclick={() => void loadAll()}>Retry</Button>
    </div>
  </div>
{:else}
  <div class="space-y-5">
    <a
      href="/"
      class="inline-flex items-center gap-1.5 rounded text-xs text-[var(--text-2)] outline-none transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
      onclick={(event) => {
        event.preventDefault()
        navigate('/')
      }}
    >
      <svg viewBox="0 0 20 20" class="size-3.5" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
        <path d="M11.5 5 6.5 10l5 5" stroke-linecap="round" stroke-linejoin="round"></path>
      </svg>
      All projects
    </a>

    <header class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
      <div class="flex min-w-0 gap-3.5 lg:flex-1">
        <ProjectLogo gradient={project.logo_gradient} emoji={project.logo_emoji} name={project.name} size="lg" />
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h1 bind:this={heading} class="truncate text-xl font-semibold tracking-[-.02em] outline-none" tabindex="-1" data-route-focus>{project.name}</h1>
            <IdChip identifier={project.identifier} />
            <ProjectStatusBadge status={project.status} />
          </div>
          <p class="mt-1 max-w-2xl text-sm leading-relaxed text-[var(--text-2)]">
            {project.description || 'No description yet.'}
          </p>
          <p class="mt-1.5 text-[11px] text-[var(--text-3)]">Updated {relativeTime(project.updated_at)}</p>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2 lg:flex-nowrap lg:shrink-0">
        <Button variant="secondary" onclick={() => (projectSettingsOpen = true)}>
          <Pencil size={15} strokeWidth={1.6} aria-hidden="true" />
          Edit project
        </Button>
        <Button variant="ghost" class="text-destructive hover:text-destructive" onclick={openDeleteProject}>
          <Trash2 size={15} strokeWidth={1.6} aria-hidden="true" />
          Delete
        </Button>
        <Button onclick={tab === 'epics' ? openNewEpic : openNewIssue}>
          <Plus size={16} strokeWidth={1.6} aria-hidden="true" />
          {tab === 'epics' ? 'New epic' : 'New issue'}
        </Button>
      </div>
    </header>

    {#if error}
      <div class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
        {error}
      </div>
    {/if}

    <div class="inline-flex w-fit items-center gap-1 rounded-[10px] border border-border bg-[var(--surface-1)] p-1" role="tablist" aria-label="Project views">
      <button
        type="button"
        role="tab"
        aria-selected={tab === 'issues'}
        class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm transition-colors {tab === 'issues'
          ? 'bg-[var(--surface-2)] text-foreground'
          : 'text-[var(--text-2)] hover:bg-white/[.03] hover:text-foreground'} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onclick={() => selectTab('issues')}
      >
        <span class="size-1.5 rounded-full {tab === 'issues' ? 'bg-primary' : 'bg-transparent'}" aria-hidden="true"></span>
        Issues
        <span class="rounded-full bg-white/[.06] px-1.5 py-0.5 text-[10px] tnum text-[var(--text-3)]">{issues.length}</span>
      </button>
      <button
        type="button"
        role="tab"
        aria-selected={tab === 'board'}
        class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm transition-colors {tab === 'board'
          ? 'bg-[var(--surface-2)] text-foreground'
          : 'text-[var(--text-2)] hover:bg-white/[.03] hover:text-foreground'} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onclick={() => selectTab('board')}
      >
        <span class="size-1.5 rounded-full {tab === 'board' ? 'bg-primary' : 'bg-transparent'}" aria-hidden="true"></span>
        Board
        {#if boardIssues.length > 0}
          <span class="rounded-full bg-white/[.06] px-1.5 py-0.5 text-[10px] tnum text-[var(--text-3)]">{boardIssues.length}</span>
        {/if}
      </button>
      <button
        type="button"
        role="tab"
        aria-selected={tab === 'epics'}
        class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm transition-colors {tab === 'epics'
          ? 'bg-[var(--surface-2)] text-foreground'
          : 'text-[var(--text-2)] hover:bg-white/[.03] hover:text-foreground'} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onclick={() => selectTab('epics')}
      >
        <span class="size-1.5 rounded-full {tab === 'epics' ? 'bg-primary' : 'bg-transparent'}" aria-hidden="true"></span>
        Epics
        <span class="rounded-full bg-white/[.06] px-1.5 py-0.5 text-[10px] tnum text-[var(--text-3)]">{epics.length}</span>
      </button>
    </div>

    {#key tab}
      <div in:fade={{ duration: motion.reduced ? 0 : 120 }}>
        {#if tab === 'issues'}
          <div role="tabpanel" aria-label="Issues">
            <IssueList
              projectRef={apiRef}
              {issues}
              {epics}
              loading={issuesLoading}
              error={issuesError}
              bind:status={statusFilter}
              bind:priority={priorityFilter}
              bind:search={search}
              onreload={() => void loadIssues()}
              onclear={clearFilters}
            />
          </div>
        {:else if tab === 'board'}
          <div role="tabpanel" aria-label="Board">
            {#if boardError}
              <div class="mb-3 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
                {boardError}
              </div>
            {/if}
            {#if boardLoading && boardIssues.length === 0}
              <div class="grid gap-3 lg:grid-cols-4" aria-hidden="true">
                {#each Array(4) as _}
                  <div class="h-72 animate-pulse rounded-2xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none"></div>
                {/each}
              </div>
            {:else}
              <IssueBoard
                projectRef={apiRef}
                issues={boardIssues}
                {epics}
                movingIds={movingIssueIds}
                truncated={boardTruncated}
                onmove={moveBoardIssue}
                onhidden={showHidden}
                onopen={openBoardIssue}
              />
            {/if}
          </div>
        {:else}
          <div role="tabpanel" aria-label="Epics">
            {#if epicIssuesError}
              <div class="mb-3 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
                {epicIssuesError}
              </div>
            {/if}
            {#if epicIssuesTruncated}
              <p class="mb-3 rounded-lg border border-border bg-[var(--surface-1)] px-3 py-2 text-[11px] text-[var(--text-3)]">
                Showing task counts from the first 1,000 issues.
              </p>
            {/if}
            <EpicList
              projectRef={apiRef}
              {epics}
              issues={epicIssues}
              loading={false}
              statsLoading={epicIssuesLoading}
              onopen={openEpic}
              onedit={openEditEpic}
              ondelete={openDeleteEpic}
            />
          </div>
        {/if}
      </div>
    {/key}
  </div>

  {#if project}
    <ProjectSettingsModal bind:open={projectSettingsOpen} {project} onsaved={onProjectSaved} />
    <IssueCreateModal bind:open={issueOpen} projectRef={apiRef} {epics} oncreated={onIssueCreated} />
    <EpicFormModal bind:open={epicOpen} projectRef={apiRef} epic={editingEpic} onsaved={onEpicSaved} />

    <ConfirmModal
      bind:open={deleteProjectOpen}
      title="Delete project"
      description="This permanently deletes the project, its epics, issues and comments. The project identifier cannot be reused."
      confirmLabel="Delete project"
      requiredText={project.identifier}
      busy={deleting}
      error={deleteError}
      onconfirm={() => void submitDeleteProject()}
    />
    <ConfirmModal
      bind:open={deleteEpicOpen}
      title="Delete epic"
      description={deletingEpic ? `Delete “${deletingEpic.title}”? Its issues will become unassigned.` : 'Delete this epic? Its issues will become unassigned.'}
      confirmLabel="Delete epic"
      busy={deleting}
      error={deleteError}
      onconfirm={() => void submitDeleteEpic()}
    />
  {/if}
{/if}
