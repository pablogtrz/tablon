<script lang="ts">
  import Button from '$lib/components/tablon/Button.svelte'
  import Modal from '$lib/components/tablon/Modal.svelte'
  import Select from '$lib/components/tablon/Select.svelte'
  import Layers from '@lucide/svelte/icons/layers'
  import { createIssue } from '$lib/api'
  import { messageOf } from '$lib/errors'
  import { newIdempotencyKey } from '$lib/idempotency'
  import { inputClass, labelClass, textareaClass } from '$lib/ui'
  import { ISSUE_STATUS_OPTIONS, PRIORITY_OPTIONS } from '$lib/meta'
  import type { Epic, Issue, IssuePriority, IssueStatus } from '$lib/types'

  let {
    open = $bindable(false),
    projectRef,
    epics,
    oncreated,
  }: {
    open?: boolean
    projectRef: string
    epics: Epic[]
    oncreated: (issue: Issue) => void
  } = $props()

  let title = $state('')
  let description = $state('')
  let status = $state<IssueStatus>('backlog')
  let priority = $state<IssuePriority>('none')
  let epicID = $state('')
  let saving = $state(false)
  let error = $state('')
  let idempotencyKey = $state('')

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
    if (!open) return
    title = ''
    description = ''
    status = 'backlog'
    priority = 'none'
    epicID = ''
    error = ''
    idempotencyKey = ''
  })

  $effect(() => {
    // A retry with the same draft must reuse the key; an edited draft gets a new one.
    void title
    void description
    void status
    void priority
    void epicID
    idempotencyKey = ''
  })

  async function submit() {
    if (saving) return
    saving = true
    error = ''
    idempotencyKey ||= newIdempotencyKey()
    try {
      const issue = await createIssue(
        projectRef,
        {
          title,
          description,
          status,
          priority,
          epic_id: epicID || null,
        },
        idempotencyKey,
      )
      oncreated(issue)
      open = false
    } catch (err) {
      error = messageOf(err)
    } finally {
      saving = false
    }
  }
</script>

<Modal bind:open title="New issue" description="Issues are the unit of work. Agents can create and claim them too." wide>
  <form id="create-issue" class="space-y-4" onsubmit={(event) => { event.preventDefault(); void submit() }}>
    <div>
      <label class={labelClass} for="issue-title">Title</label>
      <input id="issue-title" class={inputClass} bind:value={title} required maxlength="200" placeholder="Implement authentication" autocomplete="off" />
    </div>

    <div>
      <label class={labelClass} for="issue-description">Description</label>
      <textarea id="issue-description" class={textareaClass} bind:value={description} rows="5" placeholder="Context, acceptance criteria, links…"></textarea>
    </div>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div>
        <label class={labelClass} for="issue-status">Status</label>
        <Select id="issue-status" value={status} options={ISSUE_STATUS_OPTIONS} onchange={(value) => (status = value as IssueStatus)} />
      </div>
      <div>
        <label class={labelClass} for="issue-priority">Priority</label>
        <Select id="issue-priority" value={priority} options={PRIORITY_OPTIONS} onchange={(value) => (priority = value as IssuePriority)} />
      </div>
    </div>

    <div>
      <label class={labelClass} for="issue-epic">Epic</label>
      <Select id="issue-epic" value={epicID} options={epicOptions} onchange={(value) => (epicID = value)} />
    </div>

    {#if error}
      <p class="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{error}</p>
    {/if}
  </form>

  {#snippet footer()}
    <Button variant="secondary" onclick={() => (open = false)}>Cancel</Button>
    <Button type="submit" form="create-issue" disabled={saving}>{saving ? 'Creating…' : 'Create issue'}</Button>
  {/snippet}
</Modal>
