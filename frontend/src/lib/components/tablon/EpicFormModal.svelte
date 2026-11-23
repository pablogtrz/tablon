<script lang="ts">
  import Button from '$lib/components/tablon/Button.svelte'
  import Modal from '$lib/components/tablon/Modal.svelte'
  import Select from '$lib/components/tablon/Select.svelte'
  import { createEpic, updateEpic } from '$lib/api'
  import { messageOf } from '$lib/errors'
  import { inputClass, labelClass, textareaClass } from '$lib/ui'
  import { EPIC_STATUS_OPTIONS } from '$lib/meta'
  import type { Epic, EpicStatus } from '$lib/types'

  let {
    open = $bindable(false),
    projectRef,
    epic = null,
    onsaved,
  }: {
    open?: boolean
    projectRef: string
    epic?: Epic | null
    onsaved: (epic: Epic) => void
  } = $props()

  let title = $state('')
  let description = $state('')
  let status = $state<EpicStatus>('backlog')
  let saving = $state(false)
  let error = $state('')

  $effect(() => {
    if (!open) return
    title = epic?.title ?? ''
    description = epic?.description ?? ''
    status = epic?.status ?? 'backlog'
    error = ''
  })

  async function submit() {
    if (saving) return
    saving = true
    error = ''
    try {
      const saved = epic
        ? await updateEpic(epic.id, { title, description, status })
        : await createEpic(projectRef, { title, description, status })
      onsaved(saved)
      open = false
    } catch (err) {
      error = messageOf(err)
    } finally {
      saving = false
    }
  }
</script>

<Modal
  bind:open
  title={epic ? 'Edit epic' : 'New epic'}
  description={epic ? 'Update the epic.' : 'Epics group related issues inside this project.'}
>
  <form id="epic-form" class="space-y-4" onsubmit={(event) => { event.preventDefault(); void submit() }}>
    <div>
      <label class={labelClass} for="epic-title">Title</label>
      <input id="epic-title" class={inputClass} bind:value={title} required maxlength="200" placeholder="Authentication" autocomplete="off" />
    </div>

    <div>
      <label class={labelClass} for="epic-description">Description</label>
      <textarea id="epic-description" class={textareaClass} bind:value={description} rows="4"></textarea>
    </div>

    <div>
      <label class={labelClass} for="epic-status">Status</label>
      <Select id="epic-status" value={status} options={EPIC_STATUS_OPTIONS} onchange={(value) => (status = value as EpicStatus)} />
    </div>

    {#if error}
      <p class="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{error}</p>
    {/if}
  </form>

  {#snippet footer()}
    <Button variant="secondary" onclick={() => (open = false)}>Cancel</Button>
    <Button type="submit" form="epic-form" disabled={saving}>{saving ? 'Saving…' : epic ? 'Save epic' : 'Create epic'}</Button>
  {/snippet}
</Modal>
