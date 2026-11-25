<script lang="ts">
  import type { Snippet } from 'svelte'
  import X from '@lucide/svelte/icons/x'
  import { cn } from '$lib/utils'

  let {
    open = $bindable(false),
    title,
    description = '',
    wide = false,
    children,
    footer,
    onclose,
  }: {
    open?: boolean
    title: string
    description?: string
    wide?: boolean
    children: Snippet
    footer?: Snippet
    onclose?: () => void
  } = $props()

  const dialogId = 'dialog-' + Math.random().toString(36).slice(2)
  const descriptionId = dialogId + '-description'
  let dialog: HTMLDialogElement

  $effect(() => {
    if (!dialog) return
    if (open && !dialog.open) {
      dialog.showModal()
    } else if (!open && dialog.open) {
      dialog.close()
    }
  })

  function close() {
    open = false
  }
</script>

<dialog
  bind:this={dialog}
  closedby="any"
  aria-labelledby={dialogId}
  aria-describedby={description ? descriptionId : undefined}
  class={cn(
    'tablon-dialog glass-strong m-auto max-h-[85dvh] w-[min(92vw,32rem)] overflow-y-auto overscroll-contain rounded-2xl p-0 text-foreground',
    wide && 'w-[min(92vw,44rem)]',
  )}
  onclose={() => {
    open = false
    onclose?.()
  }}
  onclick={(event) => {
    if (event.target === dialog) close()
  }}
>
  <div class="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
    <div>
      <h2 id={dialogId} class="text-base font-semibold tracking-[-.01em]">{title}</h2>
      {#if description}
        <p id={descriptionId} class="mt-1 text-xs leading-relaxed text-[var(--text-3)]">{description}</p>
      {/if}
    </div>
    <button
      type="button"
      class="-mr-1 grid size-8 shrink-0 place-items-center rounded-lg text-[var(--text-3)] transition-colors hover:bg-[var(--surface-2)] hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      aria-label="Close dialog"
      onclick={close}
    >
      <X size={16} strokeWidth={1.6} aria-hidden="true" />
    </button>
  </div>

  <div class="px-5 py-4">
    {@render children()}
  </div>

  {#if footer}
    <div class="flex flex-wrap justify-end gap-2 border-t border-border px-5 py-3.5">
      {@render footer()}
    </div>
  {/if}
</dialog>
