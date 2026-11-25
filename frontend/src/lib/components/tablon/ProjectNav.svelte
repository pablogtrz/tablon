<script lang="ts">
  import { untrack } from 'svelte'
  import Archive from '@lucide/svelte/icons/archive'
  import ChevronDown from '@lucide/svelte/icons/chevron-down'
  import LayoutGrid from '@lucide/svelte/icons/layout-grid'
  import ProjectLogo from './ProjectLogo.svelte'
  import { projectNav, refreshProjectNav } from '$lib/project-nav.svelte'
  import { navigate, parseRoute, projectPath, route } from '$lib/router.svelte'
  import { cn } from '$lib/utils'

  let { onNavigate }: { onNavigate?: () => void } = $props()

  const uid = $props.id()
  const listId = `project-nav-${uid}`
  const page = $derived(parseRoute(route.path))
  const projectsActive = $derived(route.path === '/' || route.path.startsWith('/projects'))
  const dashboardActive = $derived(page.name === 'dashboard')
  const activeProjectRef = $derived(
    page.name === 'project' || page.name === 'epic' || page.name === 'issue' ? page.projectRef : '',
  )

  let open = $state(untrack(() => projectsActive))
  let lastActive = untrack(() => projectsActive)

  $effect(() => {
    if (projectsActive === lastActive) return
    open = projectsActive
    lastActive = projectsActive
  })

  $effect(() => {
    if (!open || projectNav.loaded || projectNav.loading || projectNav.error) return
    void refreshProjectNav()
  })

  function isCurrent(project: { id: string; identifier: string }): boolean {
    if (!activeProjectRef) return false
    const ref = activeProjectRef.toLowerCase()
    return ref === project.identifier.toLowerCase() || ref === project.id.toLowerCase()
  }

  function go(event: MouseEvent, href: string) {
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    void navigate(href).then((ok) => {
      if (ok) onNavigate?.()
    })
  }
</script>

<div class="min-w-0">
  <div
    class={cn(
      'group flex min-h-10 items-center rounded-lg transition-colors lg:min-h-9',
      projectsActive ? 'bg-[var(--surface-2)]' : 'hover:bg-[var(--surface-2)]',
    )}
  >
    <a
      href="/"
      aria-current={dashboardActive ? 'page' : undefined}
      class={cn(
        'flex min-w-0 flex-1 items-center gap-3 rounded-lg px-3 py-2.5 text-sm outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring lg:py-2',
        projectsActive ? 'text-foreground' : 'text-[var(--text-2)] group-hover:text-foreground',
      )}
      onclick={(event) => go(event, '/')}
    >
      <LayoutGrid
        size={16}
        strokeWidth={1.6}
        class={projectsActive ? 'text-primary' : 'text-[var(--text-3)]'}
        aria-hidden="true"
      />
      <span class="truncate">Projects</span>
    </a>
    <button
      type="button"
      aria-expanded={open}
      aria-controls={listId}
      aria-label={open ? 'Collapse projects' : 'Expand projects'}
      class="mr-1 grid size-10 shrink-0 cursor-pointer place-items-center rounded-lg text-[var(--text-3)] outline-none transition-colors hover:bg-[var(--surface-3)] hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring lg:size-9"
      onclick={() => (open = !open)}
    >
      <ChevronDown
        size={15}
        strokeWidth={1.8}
        class={cn(
          'transition-transform duration-[120ms] motion-reduce:transition-none',
          !open && '-rotate-90',
        )}
        aria-hidden="true"
      />
    </button>
  </div>

  <div id={listId} hidden={!open}>
    {#if !projectNav.loaded}
      {#if projectNav.error}
        <div class="mt-1 px-2.5 py-1.5 text-[11px] text-[var(--text-3)]">
          {projectNav.error}
          <button
            type="button"
            class="ml-1 cursor-pointer rounded text-primary underline underline-offset-2 hover:text-brand-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            onclick={() => void refreshProjectNav()}
          >
            Retry
          </button>
        </div>
      {:else}
        <div class="mt-1 space-y-1 pl-2" aria-hidden="true">
          {#each Array(3) as _}
            <div class="h-10 animate-pulse rounded-lg bg-white/[.05] motion-reduce:animate-none lg:h-9"></div>
          {/each}
        </div>
      {/if}
    {:else if projectNav.error && projectNav.items.length === 0}
      <div class="mt-1 px-2.5 py-1.5 text-[11px] text-[var(--text-3)]">
        {projectNav.error}
        <button
          type="button"
          class="ml-1 cursor-pointer rounded text-primary underline underline-offset-2 hover:text-brand-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          onclick={() => void refreshProjectNav()}
        >
          Retry
        </button>
      </div>
    {:else}
      <ul class="mt-1 space-y-0.5 pl-2" aria-label="Projects">
        {#each projectNav.items as project (project.id)}
          {@const current = isCurrent(project)}
          <li>
            <a
              href={projectPath(project.identifier)}
              aria-current={current ? 'page' : undefined}
              title={project.name}
              class={cn(
                'flex min-h-10 cursor-pointer items-center gap-2 rounded-lg px-2.5 py-1.5 text-xs outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring lg:min-h-9',
                current
                  ? 'bg-[var(--surface-2)] font-medium text-foreground'
                  : 'text-[var(--text-2)] hover:bg-[var(--surface-2)] hover:text-foreground',
              )}
              onclick={(event) => go(event, projectPath(project.identifier))}
            >
              <ProjectLogo
                gradient={project.logo_gradient}
                from={project.logo_gradient_from}
                to={project.logo_gradient_to}
                angle={project.logo_gradient_angle}
                emoji={project.logo_emoji}
                name={project.name}
                size="xs"
              />
              <span class="min-w-0 flex-1 truncate">{project.name}</span>
              {#if project.status === 'archived'}
                <Archive size={12} strokeWidth={1.6} class="shrink-0 text-[var(--text-3)]" aria-hidden="true" />
                <span class="sr-only">Archived</span>
              {/if}
            </a>
          </li>
        {:else}
          <li class="px-2.5 py-1.5 text-[11px] text-[var(--text-3)]">No projects yet</li>
        {/each}
      </ul>
    {/if}
  </div>
</div>
