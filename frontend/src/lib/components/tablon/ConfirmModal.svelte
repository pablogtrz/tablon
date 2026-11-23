<script lang="ts">
  import Button from './Button.svelte'
  import Modal from './Modal.svelte'
  import { inputClass, labelClass } from '$lib/ui'

  let {
    open = $bindable(false),
    title,
    description,
    confirmLabel = 'Delete',
    requiredText = '',
    busy = false,
    error = '',
    onconfirm,
  }: {
    open?: boolean
    title: string
    description: string
    confirmLabel?: string
    requiredText?: string
    busy?: boolean
    error?: string
    onconfirm: () => void
  } = $props()

  const formId = 'confirm-' + Math.random().toString(36).slice(2)
  let typed = $state('')

  $effect(() => {
    if (open) typed = ''
  })

  const canConfirm = $derived(!requiredText || typed.trim() === requiredText)

  function submit() {
    if (canConfirm && !busy) onconfirm()
  }
</script>

<Modal bind:open {title} {description}>
  <form id={formId} onsubmit={(event) => { event.preventDefault(); submit() }}>
    {#if requiredText}
      <label class={labelClass} for="{formId}-text">
        Type <span class="font-mono text-foreground">{requiredText}</span> to confirm
      </label>
      <input id="{formId}-text" class={inputClass} bind:value={typed} autocomplete="off" />
    {/if}
    {#if error}
      <p class="mt-3 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{error}</p>
    {/if}
  </form>

  {#snippet footer()}
    <Button variant="secondary" onclick={() => (open = false)}>Cancel</Button>
    <Button variant="danger" type="submit" form={formId} disabled={!canConfirm || busy}>
      {busy ? 'Deleting…' : confirmLabel}
    </Button>
  {/snippet}
</Modal>
