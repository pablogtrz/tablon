<script lang="ts">
  import Button from './Button.svelte'
  import Modal from './Modal.svelte'

  let {
    open = $bindable(false),
    title = 'Leave without saving?',
    description = 'You have unsaved changes. If you leave now, those changes will be lost.',
    discardLabel = 'Discard changes',
    onstay,
    ondiscard,
  }: {
    open?: boolean
    title?: string
    description?: string
    discardLabel?: string
    onstay?: () => void
    ondiscard: () => void
  } = $props()

  let settled = $state(false)

  $effect(() => {
    if (open) settled = false
  })

  function stay() {
    if (settled) return
    settled = true
    open = false
    onstay?.()
  }

  function discard() {
    if (settled) return
    settled = true
    open = false
    ondiscard()
  }
</script>

<Modal bind:open {title} {description} onclose={stay}>
  <p class="text-sm leading-relaxed text-[var(--text-2)]">
    Save your changes first, or discard them to continue.
  </p>

  {#snippet footer()}
    <Button variant="secondary" onclick={stay}>Stay</Button>
    <Button variant="danger" onclick={discard}>{discardLabel}</Button>
  {/snippet}
</Modal>
