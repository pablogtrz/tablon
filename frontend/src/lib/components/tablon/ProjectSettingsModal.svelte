<script lang="ts">
  import { untrack } from 'svelte'
  import Button from '$lib/components/tablon/Button.svelte'
  import EmojiPicker from '$lib/components/tablon/EmojiPicker.svelte'
  import LogoPicker from '$lib/components/tablon/LogoPicker.svelte'
  import Modal from '$lib/components/tablon/Modal.svelte'
  import Select from '$lib/components/tablon/Select.svelte'
  import ProjectLogo from '$lib/components/tablon/ProjectLogo.svelte'
  import { updateProject } from '$lib/api'
  import { messageOf } from '$lib/errors'
  import { inputClass, labelClass, textareaClass } from '$lib/ui'
  import { PROJECT_STATUS_OPTIONS } from '$lib/meta'
  import type { Project, ProjectStatus } from '$lib/types'

  let {
    open = $bindable(false),
    project,
    onsaved,
  }: {
    open?: boolean
    project: Project
    onsaved: (project: Project) => void
  } = $props()

  let name = $state('')
  let description = $state('')
  let status = $state<ProjectStatus>('planning')
  let emoji = $state('')
  let gradient = $state(0)
  let gradientFrom = $state('')
  let gradientTo = $state('')
  let gradientAngle = $state(135)
  let saving = $state(false)
  let error = $state('')

  let wasOpen = false

  $effect(() => {
    if (open && !wasOpen) {
      untrack(() => {
        name = project.name
        description = project.description
        status = project.status
        emoji = project.logo_emoji
        gradient = project.logo_gradient
        gradientFrom = project.logo_gradient_from
        gradientTo = project.logo_gradient_to
        gradientAngle = project.logo_gradient_angle
        error = ''
      })
    }
    wasOpen = open
  })

  async function submit() {
    if (saving) return
    saving = true
    error = ''
    try {
      const customGradient = Boolean(gradientFrom && gradientTo)
      const hadCustomGradient = Boolean(project.logo_gradient_from && project.logo_gradient_to)
      const updated = await updateProject(project.identifier, {
        name,
        description,
        status,
        logo_gradient: gradient,
        logo_emoji: emoji,
        ...(customGradient || hadCustomGradient
          ? {
              logo_gradient_from: gradientFrom,
              logo_gradient_to: gradientTo,
              logo_gradient_angle: gradientAngle,
            }
          : {}),
      })
      onsaved(updated)
      open = false
    } catch (err) {
      error = messageOf(err)
    } finally {
      saving = false
    }
  }
</script>

<Modal bind:open title="Project settings" description="Update the project name, description, status and colour.">
  <form id="project-settings" class="space-y-4" onsubmit={(event) => { event.preventDefault(); void submit() }}>
    <div class="flex items-center gap-3 rounded-xl border border-border bg-[var(--surface-1)] p-3">
      <ProjectLogo
        gradient={gradient}
        from={gradientFrom}
        to={gradientTo}
        angle={gradientAngle}
        emoji={emoji}
        name={name || project.name}
        size="lg"
      />
      <div class="min-w-0">
        <p class="truncate text-sm font-medium">{name || project.name}</p>
        <p class="mt-0.5 font-mono text-[11px] text-[var(--text-3)]">{project.identifier}</p>
      </div>
    </div>

    <LogoPicker
      bind:value={gradient}
      bind:from={gradientFrom}
      bind:to={gradientTo}
      bind:angle={gradientAngle}
      name="project-settings-colour"
    />

    <div>
      <label class={labelClass} for="settings-name">Name</label>
      <input id="settings-name" class={inputClass} bind:value={name} required maxlength="120" />
    </div>

    <div>
      <label class={labelClass} for="settings-description">Description</label>
      <textarea id="settings-description" class={textareaClass} bind:value={description} rows="3"></textarea>
    </div>

    <div>
      <label class={labelClass} for="settings-status">Status</label>
      <Select id="settings-status" value={status} options={PROJECT_STATUS_OPTIONS} onchange={(value) => (status = value as ProjectStatus)} />
    </div>

    <EmojiPicker bind:value={emoji} />

    {#if error}
      <p class="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{error}</p>
    {/if}
  </form>

  {#snippet footer()}
    <Button variant="secondary" onclick={() => (open = false)}>Cancel</Button>
    <Button type="submit" form="project-settings" disabled={saving}>{saving ? 'Saving…' : 'Save changes'}</Button>
  {/snippet}
</Modal>
