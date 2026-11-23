<script lang="ts">
  import Pencil from '@lucide/svelte/icons/pencil'
  import Trash2 from '@lucide/svelte/icons/trash-2'
  import Button from './Button.svelte'
  import EpicStatusBadge from './EpicStatusBadge.svelte'
  import { relativeTime } from '$lib/format'
  import type { Epic } from '$lib/types'

  let {
    epics,
    loading,
    onedit,
    ondelete,
  }: {
    epics: Epic[]
    loading: boolean
    onedit: (epic: Epic) => void
    ondelete: (epic: Epic) => void
  } = $props()
</script>

{#if loading}
  <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2" aria-hidden="true">
    {#each Array(2) as _}
      <div class="h-36 animate-pulse rounded-2xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none"></div>
    {/each}
  </div>
{:else if epics.length === 0}
  <div class="mt-4 grid place-items-center rounded-2xl border border-border bg-[var(--surface-1)] px-6 py-12 text-center">
    <h3 class="text-sm font-semibold">No epics yet</h3>
    <p class="mt-1 max-w-sm text-xs leading-relaxed text-[var(--text-2)]">
      Group related issues into an epic to keep the board readable.
    </p>
  </div>
{:else}
  <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2">
    {#each epics as epic (epic.id)}
      <article class="flex flex-col rounded-2xl border border-border bg-[var(--surface-1)] p-4">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <h3 class="truncate text-sm font-semibold">{epic.title}</h3>
            <p class="mt-0.5 text-[11px] text-[var(--text-3)]">Updated {relativeTime(epic.updated_at)}</p>
          </div>
          <EpicStatusBadge status={epic.status} />
        </div>

        <p class="mt-3 line-clamp-3 min-h-10 text-xs leading-relaxed text-[var(--text-2)]">
          {epic.description || 'No description yet.'}
        </p>

        <div class="mt-auto flex justify-end gap-1.5 pt-4">
          <Button variant="ghost" class="px-2.5 text-xs" onclick={() => onedit(epic)}>
            <Pencil size={13} strokeWidth={1.6} aria-hidden="true" />
            Edit
          </Button>
          <Button variant="ghost" class="px-2.5 text-xs text-destructive hover:text-destructive" onclick={() => ondelete(epic)}>
            <Trash2 size={13} strokeWidth={1.6} aria-hidden="true" />
            Delete
          </Button>
        </div>
      </article>
    {/each}
  </div>
{/if}
