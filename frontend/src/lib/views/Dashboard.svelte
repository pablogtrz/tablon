<script lang="ts">
  import { onMount } from 'svelte'
  import Plus from '@lucide/svelte/icons/plus'
  import Button from '$lib/components/tablon/Button.svelte'
  import Modal from '$lib/components/tablon/Modal.svelte'
  import ProjectCard from '$lib/components/tablon/ProjectCard.svelte'
  import ProjectLogo from '$lib/components/tablon/ProjectLogo.svelte'
  import { createProject, listProjects } from '$lib/api'
  import { messageOf } from '$lib/errors'
  import { navigate, projectPath } from '$lib/router.svelte'
  import { inputClass, labelClass, textareaClass } from '$lib/ui'
  import type { ProjectListItem } from '$lib/types'

  let projects = $state<ProjectListItem[]>([])
  let loading = $state(true)
  let error = $state('')
  let showArchived = $state(false)

  let createOpen = $state(false)
  let name = $state('')
  let identifier = $state('')
  let description = $state('')
  let emoji = $state('')
  let submitting = $state(false)
  let formError = $state('')

  async function load() {
    loading = true
    error = ''
    try {
      projects = await listProjects(showArchived)
    } catch (err) {
      error = messageOf(err)
    } finally {
      loading = false
    }
  }

  function toggleArchived() {
    showArchived = !showArchived
    void load()
  }

  function openCreate() {
    name = ''
    identifier = ''
    description = ''
    emoji = ''
    formError = ''
    createOpen = true
  }

  async function submitCreate() {
    if (submitting) return
    submitting = true
    formError = ''
    try {
      const project = await createProject({
        name,
        identifier: identifier.trim() || undefined,
        description,
        logo_emoji: emoji.trim() || undefined,
      })
      createOpen = false
      navigate(projectPath(project.identifier))
    } catch (err) {
      formError = messageOf(err)
    } finally {
      submitting = false
    }
  }

  onMount(() => {
    void load()
  })
</script>

<svelte:head>
  <title>Projects · Tablon</title>
</svelte:head>

<div class="flex flex-wrap items-end justify-between gap-4">
  <div>
    <p class="text-xs font-medium text-[var(--text-3)]">Workspace</p>
    <h1 class="mt-1 text-xl font-semibold tracking-[-.02em] outline-none" tabindex="-1" data-route-focus>Projects</h1>
    <p class="mt-1 text-sm text-[var(--text-2)]">Everything you are building, in one calm place.</p>
  </div>
  <div class="flex items-center gap-2">
    <Button variant="ghost" onclick={toggleArchived} class="px-2.5 text-xs">
      {showArchived ? 'Hide archived' : 'Show archived'}
    </Button>
    <Button onclick={openCreate}>
      <Plus size={16} strokeWidth={1.6} aria-hidden="true" />
      New project
    </Button>
  </div>
</div>

{#if error}
  <div class="mt-5 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
    {error}
  </div>
{/if}

{#if loading}
  <div class="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3" aria-hidden="true">
    {#each Array(3) as _}
      <div class="glass min-h-[188px] animate-pulse rounded-2xl p-4 motion-reduce:animate-none">
        <div class="flex gap-3">
          <div class="size-10 rounded-[10px] bg-white/[.06]"></div>
          <div class="flex-1 space-y-2 py-1">
            <div class="h-3 w-2/3 rounded bg-white/[.06]"></div>
            <div class="h-2.5 w-1/3 rounded bg-white/[.04]"></div>
          </div>
        </div>
        <div class="mt-5 h-2.5 w-full rounded bg-white/[.04]"></div>
        <div class="mt-2 h-2.5 w-4/5 rounded bg-white/[.04]"></div>
      </div>
    {/each}
  </div>
{:else if projects.length === 0}
  <div class="glass mt-6 grid place-items-center rounded-2xl px-6 py-16 text-center">
    <div class="grid size-12 place-items-center rounded-2xl border border-border bg-[var(--surface-1)] text-[var(--text-3)]">
      <svg viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
        <path d="M4 7.5A2.5 2.5 0 0 1 6.5 5h4l2 2h5A2.5 2.5 0 0 1 20 9.5v7A2.5 2.5 0 0 1 17.5 19h-11A2.5 2.5 0 0 1 4 16.5v-9Z" stroke-linejoin="round"></path>
      </svg>
    </div>
    <h2 class="mt-4 text-base font-semibold">No projects yet</h2>
    <p class="mt-1 max-w-sm text-sm leading-relaxed text-[var(--text-2)]">
      {showArchived ? 'No archived projects either.' : 'Create your first project and give your agents a place to work.'}
    </p>
    {#if !showArchived}
      <Button class="mt-5" onclick={openCreate}>
        <Plus size={16} strokeWidth={1.6} aria-hidden="true" />
        Create project
      </Button>
    {/if}
  </div>
{:else}
  <div class="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
    {#each projects as project (project.id)}
      <ProjectCard {project} />
    {/each}
  </div>
{/if}

<Modal
  bind:open={createOpen}
  title="New project"
  description="A project groups epics and issues. The identifier is used in issue IDs like PROJ-12."
>
  <form id="create-project" class="space-y-4" onsubmit={(event) => { event.preventDefault(); void submitCreate() }}>
    <div class="flex items-center gap-3 rounded-xl border border-border bg-[var(--surface-1)] p-3">
      <ProjectLogo gradient={0} emoji={emoji} name={name || 'Project'} size="md" />
      <div>
        <p class="text-sm font-medium">Project identity</p>
        <p class="mt-0.5 text-xs text-[var(--text-3)]">A gradient is chosen automatically. Add an emoji to override it.</p>
      </div>
    </div>

    <div>
      <label class={labelClass} for="project-name">Name</label>
      <input id="project-name" class={inputClass} bind:value={name} required maxlength="120" placeholder="Tablon" autocomplete="off" />
    </div>

    <div>
      <label class={labelClass} for="project-identifier">Identifier</label>
      <input id="project-identifier" class={inputClass} bind:value={identifier} maxlength="10" placeholder="Auto from name" autocomplete="off" />
      <p class="mt-1 text-[11px] text-[var(--text-3)]">2-10 uppercase letters or digits. Immutable once created.</p>
    </div>

    <div>
      <label class={labelClass} for="project-description">Description</label>
      <textarea id="project-description" class={textareaClass} bind:value={description} rows="3" placeholder="What is this project about?"></textarea>
    </div>

    <div>
      <label class={labelClass} for="project-emoji">Emoji <span class="font-normal text-[var(--text-3)]">(optional)</span></label>
      <input id="project-emoji" class={inputClass} bind:value={emoji} maxlength="16" placeholder="🚀" autocomplete="off" />
    </div>

    {#if formError}
      <p class="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{formError}</p>
    {/if}
  </form>

  {#snippet footer()}
    <Button variant="secondary" onclick={() => (createOpen = false)}>Cancel</Button>
    <Button type="submit" form="create-project" disabled={submitting}>
      {submitting ? 'Creating…' : 'Create project'}
    </Button>
  {/snippet}
</Modal>
