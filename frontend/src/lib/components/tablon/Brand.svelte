<script lang="ts">
  import { navigate } from '$lib/router.svelte'

  let { compact = false, markOnly = false, onclick }: { compact?: boolean; markOnly?: boolean; onclick?: () => void } = $props()
</script>

<a
  href="/"
  class="flex min-w-0 items-center gap-3 rounded-xl outline-none focus-visible:ring-2 focus-visible:ring-ring {markOnly ? 'justify-center' : ''}"
  aria-label="Tablon home"
  onclick={(event) => {
    event.preventDefault()
    navigate('/')
    onclick?.()
  }}
>
  <span
    class="grid size-9 shrink-0 place-items-center rounded-[10px] bg-primary text-primary-foreground shadow-[inset_0_1px_0_rgba(255,255,255,.2)]"
    aria-hidden="true"
  >
    <svg viewBox="0 0 24 24" class="size-5" fill="currentColor">
      <rect x="4" y="4" width="6" height="16" rx="2" opacity=".95"></rect>
      <rect x="12" y="4" width="6" height="9" rx="2" opacity=".7"></rect>
      <rect x="12" y="15" width="6" height="5" rx="2" opacity=".45"></rect>
    </svg>
  </span>
  {#if !markOnly}
    <span class="min-w-0">
      <span class="block truncate text-sm font-semibold tracking-[-.02em]">Tablon</span>
      {#if !compact}
        <span class="block truncate text-[11px] text-[var(--text-3)]">Lightweight issue tracker</span>
      {/if}
    </span>
  {/if}
</a>
