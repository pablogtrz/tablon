<script lang="ts">
  import { live } from '$lib/live.svelte'
  import type { LiveState } from '$lib/live.svelte'

  let { compact = false }: { compact?: boolean } = $props()

  const labels: Record<LiveState, string> = {
    idle: 'Live updates idle',
    connecting: 'Connecting',
    live: 'Live',
    offline: 'Offline',
    unauthorized: 'Token required',
  }

  const dots: Record<LiveState, string> = {
    idle: 'bg-[var(--text-3)]',
    connecting: 'bg-[var(--brand)]',
    live: 'bg-[var(--status-done)]',
    offline: 'bg-[var(--status-backlog)]',
    unauthorized: 'bg-[var(--danger)]',
  }

  const label = $derived(labels[live.state])
  const dot = $derived(dots[live.state])
</script>

{#if live.state !== 'idle'}
  {#if compact}
    <span
      class="grid size-6 place-items-center rounded-md text-[var(--text-3)]"
      role="status"
      aria-live="polite"
      aria-label={label}
      title={label}
    >
      <span class="size-1.5 rounded-full {dot}" aria-hidden="true"></span>
    </span>
  {:else}
    <span
      class="inline-flex items-center gap-1.5 text-[11px] text-[var(--text-3)]"
      role="status"
      aria-live="polite"
    >
      <span class="size-1.5 rounded-full {dot}" aria-hidden="true"></span>
      {label}
    </span>
  {/if}
{/if}
