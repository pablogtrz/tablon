<script lang="ts">
  import Brand from '$lib/components/tablon/Brand.svelte'
  import LiveIndicator from '$lib/components/tablon/LiveIndicator.svelte'
  import MobileNav from '$lib/components/tablon/MobileNav.svelte'
  import ProjectNav from '$lib/components/tablon/ProjectNav.svelte'
  import Dashboard from '$lib/views/Dashboard.svelte'
  import EpicView from '$lib/views/EpicView.svelte'
  import NotFound from '$lib/views/NotFound.svelte'
  import ProjectView from '$lib/views/ProjectView.svelte'
  import SettingsView from '$lib/views/SettingsView.svelte'
  import { navigate, parseRoute, route } from '$lib/router.svelte'
  import { projectNav, refreshProjectNav } from '$lib/project-nav.svelte'
  import { motion } from '$lib/motion.svelte'
  import LayoutGrid from '@lucide/svelte/icons/layout-grid'
  import Menu from '@lucide/svelte/icons/menu'
  import SettingsIcon from '@lucide/svelte/icons/settings'
  import { tick, untrack } from 'svelte'
  import { cubicOut } from 'svelte/easing'
  import { fly } from 'svelte/transition'

  let navOpen = $state(false)
  let issueViewPromise: Promise<typeof import('$lib/views/IssueView.svelte')> | null = null

  function loadIssueView() {
    issueViewPromise ??= import('$lib/views/IssueView.svelte')
    return issueViewPromise
  }
  const page = $derived(parseRoute(route.path))
  const projectsActive = $derived(route.path === '/' || route.path.startsWith('/projects'))
  const settingsActive = $derived(route.path === '/settings')

  $effect(() => {
    const currentPath = route.path
    if (!currentPath) return
    void tick().then(() => {
      window.scrollTo({ top: 0, behavior: 'auto' })
      document.querySelector<HTMLElement>('main [data-route-focus]')?.focus({ preventScroll: true })
    })
  })

  $effect(() => {
    if (!projectsActive) return
    untrack(() => {
      if (projectNav.loading) return
      void refreshProjectNav()
    })
  })

  const navClass = (active: boolean) =>
    `group flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring ${
      active
        ? 'bg-[var(--surface-2)] text-foreground'
        : 'text-[var(--text-2)] hover:bg-[var(--surface-2)] hover:text-foreground'
    }`
</script>

<div class="min-h-dvh md:grid md:grid-cols-[64px_minmax(0,1fr)] lg:grid-cols-[248px_minmax(0,1fr)]">
  <aside class="hidden border-r border-border bg-[var(--bg-2)]/70 lg:sticky lg:top-0 lg:flex lg:h-dvh lg:flex-col">
    <div class="px-4 py-5">
      <Brand />
    </div>

    <nav class="min-h-0 flex-1 space-y-1 overflow-y-auto px-3 py-2" aria-label="Main">
      <ProjectNav />
      <a
        href="/settings"
        aria-current={settingsActive ? 'page' : undefined}
        class={navClass(settingsActive)}
        onclick={(event) => {
          event.preventDefault()
          navigate('/settings')
        }}
      >
        <SettingsIcon size={16} strokeWidth={1.6} class={settingsActive ? 'text-primary' : 'text-[var(--text-3)]'} aria-hidden="true" />
        Settings
      </a>
    </nav>

    <div class="border-t border-border px-4 py-4 text-[11px] leading-relaxed text-[var(--text-3)]">
      <div class="mb-2 empty:hidden">
        <LiveIndicator />
      </div>
      Tablon v0.1<br />local-first, agent-ready
    </div>
  </aside>

  <aside class="hidden border-r border-border bg-[var(--bg-2)]/70 md:sticky md:top-0 md:flex md:h-dvh md:flex-col md:items-center lg:hidden" aria-label="Tablet navigation">
    <div class="py-5">
      <Brand markOnly />
    </div>
    <nav class="flex flex-1 flex-col items-center gap-1.5 px-2" aria-label="Main">
      <a
        href="/"
        aria-label="Projects"
        title="Projects"
        aria-current={projectsActive ? 'page' : undefined}
        class="grid size-10 place-items-center rounded-xl transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring {projectsActive ? 'bg-[var(--surface-2)] text-primary' : 'text-[var(--text-2)] hover:bg-[var(--surface-2)] hover:text-foreground'}"
        onclick={(event) => {
          event.preventDefault()
          navigate('/')
        }}
      >
        <LayoutGrid size={18} strokeWidth={1.6} aria-hidden="true" />
      </a>
      <a
        href="/settings"
        aria-label="Settings"
        title="Settings"
        aria-current={settingsActive ? 'page' : undefined}
        class="grid size-10 place-items-center rounded-xl transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring {settingsActive ? 'bg-[var(--surface-2)] text-primary' : 'text-[var(--text-2)] hover:bg-[var(--surface-2)] hover:text-foreground'}"
        onclick={(event) => {
          event.preventDefault()
          navigate('/settings')
        }}
      >
        <SettingsIcon size={18} strokeWidth={1.6} aria-hidden="true" />
      </a>
    </nav>
    <div class="pb-4 empty:hidden">
      <LiveIndicator compact />
    </div>
  </aside>

  <div class="min-w-0">
    <header class="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-border bg-[var(--bg)]/80 px-4 backdrop-blur md:hidden">
      <Brand compact />
      <div class="flex items-center gap-1">
        <LiveIndicator />
        <button
          type="button"
          class="grid size-10 place-items-center rounded-xl text-[var(--text-2)] transition-colors hover:bg-[var(--surface-2)] hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          aria-label="Open navigation"
          aria-expanded={navOpen}
          onclick={() => (navOpen = true)}
        >
          <Menu size={20} strokeWidth={1.6} aria-hidden="true" />
        </button>
      </div>
    </header>

    <main class="mx-auto w-full max-w-[1360px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      {#key route.path}
        <div
          in:fly={{ y: motion.reduced ? 0 : 6, duration: motion.reduced ? 0 : 180, easing: cubicOut }}
          class="min-w-0"
        >
          {#if page.name === 'dashboard'}
            <Dashboard />
          {:else if page.name === 'project'}
            <ProjectView projectRef={page.projectRef} />
          {:else if page.name === 'epic'}
            <EpicView projectRef={page.projectRef} epicRef={page.epicRef} />
          {:else if page.name === 'issue'}
            {#await loadIssueView()}
              <div class="h-72 animate-pulse rounded-2xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none" aria-hidden="true"></div>
            {:then { default: IssueView }}
              <IssueView projectRef={page.projectRef} issueRef={page.issueRef} />
            {:catch}
              <div class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
                Could not load the issue view.
              </div>
            {/await}
          {:else if page.name === 'settings'}
            <SettingsView />
          {:else}
            <NotFound />
          {/if}
        </div>
      {/key}
    </main>
  </div>
</div>

<MobileNav bind:open={navOpen} />
