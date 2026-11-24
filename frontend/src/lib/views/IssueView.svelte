<script lang="ts">
  import { onMount, tick } from 'svelte'
  import Layers from '@lucide/svelte/icons/layers'
  import Pencil from '@lucide/svelte/icons/pencil'
  import Trash2 from '@lucide/svelte/icons/trash-2'
  import UserMinus from '@lucide/svelte/icons/user-minus'
  import AgentChip from '$lib/components/tablon/AgentChip.svelte'
  import Button from '$lib/components/tablon/Button.svelte'
  import CommentsSection from '$lib/components/tablon/CommentsSection.svelte'
  import ConfirmModal from '$lib/components/tablon/ConfirmModal.svelte'
  import IdChip from '$lib/components/tablon/IdChip.svelte'
  import MarkdownEditor from '$lib/components/tablon/MarkdownEditor.svelte'
  import MarkdownRenderer from '$lib/components/tablon/MarkdownRenderer.svelte'
  import Select from '$lib/components/tablon/Select.svelte'
  import UnsavedChangesDialog from '$lib/components/tablon/UnsavedChangesDialog.svelte'
  import { deleteIssue, getIssue, getProject, listEpics, releaseIssue, updateIssue } from '$lib/api'
  import type { UpdateIssueInput } from '$lib/api'
  import { messageOf } from '$lib/errors'
  import { subscribeProjectEvents } from '$lib/events'
  import type { ProjectEvent } from '$lib/events'
  import { formatDate, relativeTime } from '$lib/format'
  import { toggleTask } from '$lib/markdown'
  import { ISSUE_STATUS_OPTIONS, PRIORITY_OPTIONS } from '$lib/meta'
  import { navigate, projectPath, setNavigationGuard } from '$lib/router.svelte'
  import { cn } from '$lib/utils'
  import { inputClass, labelClass } from '$lib/ui'
  import type {
    Epic,
    Issue,
    IssuePriority,
    IssueStatus,
    Project,
  } from '$lib/types'

  let { projectRef, issueRef }: { projectRef: string; issueRef: string } = $props()

  let issue = $state<Issue | null>(null)
  let project = $state<Project | null>(null)
  let epics = $state<Epic[]>([])
  let loading = $state(true)
  let error = $state('')
  let pageFocus = $state<HTMLDivElement>()
  let focusPending = true

  let titleDraft = $state('')
  let titleSaving = $state(false)
  let titleError = $state('')

  let descriptionMode = $state<'read' | 'edit'>('read')
  let descriptionDraft = $state('')
  let descriptionSaving = $state(false)
  let descriptionError = $state('')
  let taskSaving = $state(false)
  let taskError = $state('')

  let propertySaving = $state<string[]>([])
  let propertyError = $state('')

  let deleteOpen = $state(false)
  let deleting = $state(false)
  let deleteError = $state('')

  let releasing = $state(false)
  let releaseError = $state('')

  let cancelDialogOpen = $state(false)
  let commentsRefresh = $state(0)
  let liveMessage = $state('')
  let eventTimer: ReturnType<typeof setTimeout> | null = null
  let pendingIssue = false
  let pendingComments = false
  let guardOpen = $state(false)
  let guardResolve: ((value: boolean) => void) | null = null
  let guardPromise: Promise<boolean> | null = null

  const epicOptions = $derived([
    { value: '', label: 'No epic', icon: Layers },
    ...epics.map((epic) => ({
      value: epic.id,
      label: epic.title,
      icon: Layers,
      tone: 'var(--text-3)',
    })),
  ])

  $effect(() => {
    if (!focusPending || !pageFocus) return
    focusPending = false
    pageFocus.focus({ preventScroll: true })
  })

  const titleDirty = $derived(Boolean(issue && titleDraft !== issue.title))
  const descriptionDirty = $derived(
    Boolean(issue && descriptionMode === 'edit' && descriptionDraft !== issue.description),
  )

  function isDirty(): boolean {
    return titleDirty || descriptionDirty || titleSaving || descriptionSaving
  }

  async function load() {
    loading = true
    error = ''
    try {
      const [loadedIssue, loadedProject, loadedEpics] = await Promise.all([
        getIssue(issueRef),
        getProject(projectRef),
        listEpics(projectRef),
      ])
      issue = loadedIssue
      project = loadedProject
      epics = loadedEpics
      titleDraft = loadedIssue.title
      descriptionDraft = loadedIssue.description
    } catch (err) {
      error = messageOf(err)
    } finally {
      loading = false
    }
  }

  async function saveTitle() {
    if (!issue || titleSaving) return
    const next = titleDraft.trim()
    if (!next) {
      titleDraft = issue.title
      titleError = 'Title is required.'
      return
    }
    if (next === issue.title) {
      titleError = ''
      return
    }

    titleSaving = true
    titleError = ''
    try {
      const updated = await updateIssue(issue.identifier, { title: next })
      issue = { ...issue, title: updated.title, updated_at: updated.updated_at }
      titleDraft = updated.title
    } catch (err) {
      titleError = messageOf(err)
    } finally {
      titleSaving = false
    }
  }

  function handleTitleKeydown(event: KeyboardEvent) {
    if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return
    event.preventDefault()
    void saveTitle()
  }

  function startDescriptionEdit() {
    if (!issue) return
    descriptionDraft = issue.description
    descriptionError = ''
    descriptionMode = 'edit'
  }

  async function saveDescription(next: string) {
    if (!issue || descriptionSaving) return
    descriptionSaving = true
    descriptionError = ''
    try {
      const updated = await updateIssue(issue.identifier, { description: next })
      issue = { ...issue, description: updated.description, updated_at: updated.updated_at }
      descriptionDraft = updated.description
      descriptionMode = 'read'
    } catch (err) {
      descriptionError = messageOf(err)
    } finally {
      descriptionSaving = false
    }
  }

  function cancelDescription() {
    if (descriptionDirty) {
      cancelDialogOpen = true
      return
    }
    descriptionMode = 'read'
  }

  function discardDescription() {
    if (!issue) return
    descriptionDraft = issue.description
    descriptionMode = 'read'
    descriptionError = ''
  }

  async function handleTaskToggle(index: number) {
    if (!issue || taskSaving) return
    const previous = issue.description
    const next = toggleTask(previous, index)
    if (next === previous) return

    taskSaving = true
    taskError = ''
    issue = { ...issue, description: next }
    try {
      const updated = await updateIssue(issue.identifier, { description: next })
      issue = { ...issue, description: updated.description, updated_at: updated.updated_at }
    } catch (err) {
      issue = { ...issue, description: previous }
      taskError = messageOf(err)
    } finally {
      taskSaving = false
    }
  }

  async function updateProperty(field: 'status' | 'priority' | 'epic_id', value: string | null) {
    if (!issue || propertySaving.length > 0) return

    const previous = issue
    let optimistic: Issue
    let payload: UpdateIssueInput

    if (field === 'status') {
      optimistic = { ...issue, status: value as IssueStatus }
      payload = { status: value as IssueStatus }
    } else if (field === 'priority') {
      optimistic = { ...issue, priority: value as IssuePriority }
      payload = { priority: value as IssuePriority }
    } else {
      optimistic = { ...issue, epic_id: value }
      payload = { epic_id: value }
    }

    issue = optimistic
    propertySaving = [...propertySaving, field]
    propertyError = ''
    try {
      const updated = await updateIssue(previous.identifier, payload)
      if (field === 'status') {
        issue = { ...issue, status: updated.status, updated_at: updated.updated_at }
      } else if (field === 'priority') {
        issue = { ...issue, priority: updated.priority, updated_at: updated.updated_at }
      } else {
        issue = { ...issue, epic_id: updated.epic_id, updated_at: updated.updated_at }
      }
    } catch (err) {
      issue = previous
      propertyError = messageOf(err)
    } finally {
      propertySaving = propertySaving.filter((item) => item !== field)
    }
  }

  async function releaseAgent() {
    if (!issue?.agent_id || releasing) return
    releasing = true
    releaseError = ''
    try {
      issue = await releaseIssue(issue.identifier, issue.agent_id)
    } catch (err) {
      releaseError = messageOf(err)
    } finally {
      releasing = false
    }
  }

  async function submitDelete() {
    if (!issue || deleting) return
    deleting = true
    deleteError = ''
    try {
      await deleteIssue(issue.identifier)
      descriptionMode = 'read'
      descriptionDraft = issue.description
      titleDraft = issue.title
      titleError = ''
      navigate(projectPath(projectRef))
    } catch (err) {
      deleteError = messageOf(err)
    } finally {
      deleting = false
    }
  }

  async function announce(message: string) {
    if (typeof document !== 'undefined' && document.hidden) return
    liveMessage = ''
    await tick()
    liveMessage = message
  }

  function queueRefresh(issueChanged: boolean, commentsChanged: boolean) {
    pendingIssue ||= issueChanged
    pendingComments ||= commentsChanged
    if (eventTimer) clearTimeout(eventTimer)
    eventTimer = setTimeout(() => {
      eventTimer = null
      if (pendingIssue) void refreshFromServer()
      if (pendingComments) commentsRefresh += 1
      pendingIssue = false
      pendingComments = false
    }, 150)
  }

  async function refreshFromServer() {
    if (!issue) return
    const keepTitle = titleDirty
    const keepDescription = descriptionDirty
    try {
      const updated = await getIssue(issue.identifier)
      issue = {
        ...updated,
        title: keepTitle ? issue.title : updated.title,
        description: keepDescription ? issue.description : updated.description,
      }
      if (!keepTitle) titleDraft = updated.title
      if (!keepDescription && descriptionMode === 'read') descriptionDraft = updated.description
    } catch {
      // The stream sync will retry; keep the current view usable.
    }
  }

  function eventMatches(event: ProjectEvent): boolean {
    if (!issue) return false
    return event.issue_id === issue.id || event.identifier === issue.identifier
  }

  $effect(() => {
    if (!project) return
    const dispose = subscribeProjectEvents(projectRef, {
      onsync: () => queueRefresh(true, true),
      onEvent: (event) => {
        if (!eventMatches(event)) return
        if (event.type === 'issue.deleted') {
          void announce(`Issue ${event.identifier ?? issue?.identifier ?? ''} deleted`)
          void navigate(projectPath(projectRef))
          return
        }
        if (event.type.startsWith('comment.')) {
          const label = event.identifier ?? issue?.identifier ?? 'issue'
          const messages: Record<string, string> = {
            'comment.created': `New comment on ${label}`,
            'comment.updated': `Comment updated on ${label}`,
            'comment.deleted': `Comment deleted on ${label}`,
          }
          void announce(messages[event.type] ?? 'Comments updated')
          queueRefresh(false, true)
          return
        }
        if (event.type.startsWith('issue.')) {
          void announce(`Issue ${event.identifier ?? issue?.identifier ?? ''} updated`)
          queueRefresh(true, false)
        }
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

  function askLeave(): Promise<boolean> {
    if (guardPromise) return guardPromise
    guardOpen = true
    guardPromise = new Promise<boolean>((resolve) => {
      guardResolve = resolve
    })
    return guardPromise
  }

  function resolveGuard(value: boolean) {
    guardOpen = false
    guardResolve?.(value)
    guardResolve = null
    guardPromise = null
  }

  onMount(() => {
    void load()

    const dispose = setNavigationGuard(async () => {
      if (!isDirty()) return true
      return askLeave()
    })

    const beforeUnload = (event: BeforeUnloadEvent) => {
      if (!isDirty()) return
      event.preventDefault()
      event.returnValue = ''
    }
    window.addEventListener('beforeunload', beforeUnload)

    return () => {
      dispose()
      window.removeEventListener('beforeunload', beforeUnload)
    }
  })
</script>

<svelte:head>
  <title>{issue?.identifier ?? 'Issue'} · Tablon</title>
</svelte:head>

{#if loading}
  <div class="space-y-5" aria-hidden="true">
    <div class="h-4 w-32 animate-pulse rounded bg-white/[.06] motion-reduce:animate-none"></div>
    <div class="grid grid-cols-1 gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div class="h-96 animate-pulse rounded-2xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none"></div>
      <div class="h-72 animate-pulse rounded-2xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none"></div>
    </div>
  </div>
{:else if !issue}
  <div class="glass mx-auto mt-10 max-w-lg rounded-2xl p-6 text-center">
    <h1 class="text-base font-semibold">Could not load this issue</h1>
    <p class="mt-2 text-sm leading-relaxed text-[var(--text-2)]">{error || 'The issue does not exist or the API is unavailable.'}</p>
    <div class="mt-5 flex justify-center gap-2">
      <Button variant="secondary" onclick={() => navigate(projectPath(projectRef))}>Back to project</Button>
      <Button onclick={() => void load()}>Retry</Button>
    </div>
  </div>
{:else}
  <div bind:this={pageFocus} tabindex="-1" class="space-y-5 outline-none" aria-label="Issue detail">
    <a
      href={projectPath(projectRef)}
      class="inline-flex items-center gap-1.5 rounded text-xs text-[var(--text-2)] outline-none transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
      onclick={(event) => {
        event.preventDefault()
        void navigate(projectPath(projectRef))
      }}
    >
      <svg viewBox="0 0 20 20" class="size-3.5" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
        <path d="M11.5 5 6.5 10l5 5" stroke-linecap="round" stroke-linejoin="round"></path>
      </svg>
      {project?.name ?? 'Project'}
    </a>

    <div class="flex flex-wrap items-center gap-2">
      <IdChip identifier={issue.identifier} />
      {#if issue.agent_id}
        <AgentChip agentId={issue.agent_id} />
      {/if}
    </div>

    <div class="grid grid-cols-1 gap-5 lg:grid-cols-[minmax(0,1fr)_320px] lg:items-start">
      <section class="glass rounded-2xl p-4 sm:p-5">
        <div class="group relative">
          <label class="sr-only" for="issue-title">Issue title</label>
          <input
            id="issue-title"
            class={cn(
              inputClass,
              'h-11 border-transparent bg-transparent px-1 pr-8 text-lg font-semibold tracking-[-.01em] focus:bg-[var(--surface-1)]',
            )}
            bind:value={titleDraft}
            readonly={titleSaving}
            maxlength="200"
            autocomplete="off"
            aria-invalid={Boolean(titleError)}
            aria-describedby={titleError ? 'issue-title-error' : undefined}
            oninput={() => { if (titleError) titleError = '' }}
            onblur={() => void saveTitle()}
            onkeydown={handleTitleKeydown}
          />
          <Pencil
            size={15}
            strokeWidth={1.6}
            class="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-[var(--text-3)] opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100"
            aria-hidden="true"
          />
        </div>
        {#if titleError}
          <p id="issue-title-error" class="mt-1 text-xs text-destructive" role="alert">{titleError}</p>
        {/if}

        <div class="mt-5 border-t border-border pt-4">
          <div class="flex items-center justify-between gap-3">
            <h2 class="text-xs font-medium text-[var(--text-3)]">Description</h2>
            {#if descriptionMode === 'read'}
              <Button variant="ghost" class="h-8 px-2 text-xs" onclick={startDescriptionEdit}>
                <Pencil size={14} strokeWidth={1.6} aria-hidden="true" />
                Edit
              </Button>
            {/if}
          </div>

          {#if descriptionMode === 'read'}
            {#if issue.description.trim()}
              <div class="mt-2">
                <MarkdownRenderer
                  source={issue.description}
                  interactive
                  disabled={taskSaving}
                  ontasktoggle={(index) => void handleTaskToggle(index)}
                />
              </div>
            {:else}
              <p class="mt-2 text-sm leading-relaxed text-[var(--text-3)]">
                No description yet.
                <button
                  type="button"
                  class="cursor-pointer text-primary underline underline-offset-2 hover:text-brand-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  onclick={startDescriptionEdit}
                >
                  Add one
                </button>
              </p>
            {/if}
          {:else}
            <div class="mt-2">
              <MarkdownEditor
                bind:value={descriptionDraft}
                ariaLabel="Issue description"
                dirty={descriptionDirty}
                saving={descriptionSaving}
                error={descriptionError}
                onsave={saveDescription}
                oncancel={cancelDescription}
              />
            </div>
          {/if}

          {#if taskError}
            <p class="mt-3 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{taskError}</p>
          {/if}
        </div>
      </section>

      <aside class="space-y-4">
        <section class="rounded-2xl border border-border bg-[var(--surface-1)] p-4">
          <h2 class="text-xs font-medium text-[var(--text-3)]">Properties</h2>

          <div class="mt-3 space-y-3">
            <div>
              <label class={labelClass} for="issue-status">Status</label>
              <Select
                id="issue-status"
                value={issue.status}
                options={ISSUE_STATUS_OPTIONS}
                disabled={propertySaving.length > 0}
                onchange={(value) => void updateProperty('status', value)}
              />
            </div>

            <div>
              <label class={labelClass} for="issue-priority">Priority</label>
              <Select
                id="issue-priority"
                value={issue.priority}
                options={PRIORITY_OPTIONS}
                disabled={propertySaving.length > 0}
                onchange={(value) => void updateProperty('priority', value)}
              />
            </div>

            <div>
              <label class={labelClass} for="issue-epic">Epic</label>
              <Select
                id="issue-epic"
                value={issue.epic_id ?? ''}
                options={epicOptions}
                disabled={propertySaving.length > 0}
                onchange={(value) => void updateProperty('epic_id', value || null)}
              />
            </div>
          </div>

          {#if propertyError}
            <p class="mt-3 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{propertyError}</p>
          {/if}
        </section>

        <section class="rounded-2xl border border-border bg-[var(--surface-1)] p-4">
          <h2 class="text-xs font-medium text-[var(--text-3)]">Agent</h2>
          <div class="mt-3">
            {#if issue.agent_id}
              <AgentChip agentId={issue.agent_id} />
              <Button
                variant="secondary"
                class="mt-3 w-full text-xs"
                disabled={releasing}
                onclick={() => void releaseAgent()}
              >
                <UserMinus size={14} strokeWidth={1.6} aria-hidden="true" />
                {releasing ? 'Releasing…' : 'Release agent'}
              </Button>
            {:else}
              <p class="text-xs leading-relaxed text-[var(--text-2)]">Unassigned. Agents claim issues through the CLI or API.</p>
            {/if}
            {#if releaseError}
              <p class="mt-2 text-xs text-destructive" role="alert">{releaseError}</p>
            {/if}
          </div>
        </section>

        <section class="rounded-2xl border border-border bg-[var(--surface-1)] p-4">
          <h2 class="text-xs font-medium text-[var(--text-3)]">Dates</h2>
          <dl class="mt-3 space-y-2 text-xs">
            <div class="flex items-center justify-between gap-3">
              <dt class="text-[var(--text-3)]">Created</dt>
              <dd class="text-[var(--text-2)]" title={formatDate(issue.created_at)}>{relativeTime(issue.created_at)}</dd>
            </div>
            <div class="flex items-center justify-between gap-3">
              <dt class="text-[var(--text-3)]">Updated</dt>
              <dd class="text-[var(--text-2)]" title={formatDate(issue.updated_at)}>{relativeTime(issue.updated_at)}</dd>
            </div>
          </dl>
        </section>

        <Button
          variant="ghost"
          class="w-full justify-start text-destructive hover:text-destructive"
          onclick={() => {
            deleteError = ''
            deleteOpen = true
          }}
        >
          <Trash2 size={14} strokeWidth={1.6} aria-hidden="true" />
          Delete issue
        </Button>
      </aside>
    </div>

    <CommentsSection issueRef={issue.identifier} refreshKey={commentsRefresh} />
    <p class="sr-only" role="status" aria-live="polite">{liveMessage}</p>
  </div>

  <ConfirmModal
    bind:open={deleteOpen}
    title="Delete issue"
    description={issue ? `Delete “${issue.title}”? This cannot be undone. The identifier ${issue.identifier} will not be reused.` : 'Delete this issue?'}
    confirmLabel="Delete issue"
    busy={deleting}
    error={deleteError}
    onconfirm={() => void submitDelete()}
  />

  <UnsavedChangesDialog
    bind:open={cancelDialogOpen}
    title="Discard description changes?"
    description="Your description edits have not been saved."
    ondiscard={discardDescription}
  />
  <UnsavedChangesDialog
    bind:open={guardOpen}
    title="Leave without saving?"
    description="You have unsaved changes on this issue."
    ondiscard={() => resolveGuard(true)}
    onstay={() => resolveGuard(false)}
  />
{/if}
