<script lang="ts">
  import { navigate, route } from '$lib/router.svelte'
  import Brand from './Brand.svelte'
  import ProjectNav from './ProjectNav.svelte'
  import SettingsIcon from '@lucide/svelte/icons/settings'
  import X from '@lucide/svelte/icons/x'

  let { open = $bindable(false) }: { open?: boolean } = $props()

  let dialog: HTMLDialogElement

  $effect(() => {
    if (!dialog) return
    if (open && !dialog.open) {
      dialog.showModal()
    } else if (!open && dialog.open) {
      dialog.close()
    }
  })

  const settingsActive = $derived(route.path === '/settings')

  function go(to: string) {
    navigate(to)
    open = false
  }
</script>

<dialog
  bind:this={dialog}
  closedby="any"
  aria-label="Navigation"
  class="tablon-drawer m-0 h-dvh max-h-none w-[280px] max-w-[85vw] rounded-none border-0 bg-[var(--bg-2)] p-0 text-foreground shadow-2xl"
  onclose={() => (open = false)}
  onclick={(event) => {
    if (event.target === dialog) open = false
  }}
>
  <div class="flex h-full flex-col">
    <div class="flex items-center justify-between gap-3 border-b border-border px-4 py-4">
      <Brand onclick={() => (open = false)} />
      <button
        type="button"
        class="grid size-8 place-items-center rounded-lg text-[var(--text-3)] transition-colors hover:bg-[var(--surface-2)] hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        aria-label="Close navigation"
        onclick={() => (open = false)}
      >
        <X size={16} strokeWidth={1.6} aria-hidden="true" />
      </button>
    </div>

    <nav class="min-h-0 flex-1 space-y-1 overflow-y-auto px-3 py-3" aria-label="Main">
      <ProjectNav onNavigate={() => (open = false)} />
      <button
        type="button"
        class="group flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left text-sm transition-colors {settingsActive
          ? 'bg-[var(--surface-2)] text-foreground'
          : 'text-[var(--text-2)] hover:bg-[var(--surface-2)] hover:text-foreground'} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onclick={() => go('/settings')}
      >
        <SettingsIcon size={16} strokeWidth={1.6} class={settingsActive ? 'text-primary' : 'text-[var(--text-3)]'} aria-hidden="true" />
        Settings
      </button>
    </nav>

    <div class="border-t border-border px-4 py-4 text-[11px] text-[var(--text-3)]">
      Tablon v0.1 · local-first
    </div>
  </div>
</dialog>
