<script lang="ts">
  import Pencil from '@lucide/svelte/icons/pencil'
  import Trash2 from '@lucide/svelte/icons/trash-2'
  import UserRound from '@lucide/svelte/icons/user-round'
  import AgentChip from './AgentChip.svelte'
  import MarkdownRenderer from './MarkdownRenderer.svelte'
  import { formatDate, relativeTime } from '$lib/format'
  import type { Comment } from '$lib/types'

  let {
    comment,
    onedit,
    ondelete,
  }: {
    comment: Comment
    onedit: () => void
    ondelete: () => void
  } = $props()

  const edited = $derived(comment.updated_at !== comment.created_at)
</script>

<article
  id={`comment-${comment.id}`}
  tabindex="-1"
  class="rounded-xl border border-border bg-[var(--surface-1)] p-3 outline-none focus-visible:ring-2 focus-visible:ring-ring sm:p-4"
  aria-label={`Comment by ${comment.agent_id ?? 'local user'}`}
>
  <header class="flex flex-wrap items-center justify-between gap-2">
    <div class="flex min-w-0 flex-wrap items-center gap-2">
      {#if comment.agent_id}
        <AgentChip agentId={comment.agent_id} />
      {:else}
        <span class="inline-flex items-center gap-1 rounded-full bg-[var(--surface-2)] px-2 py-0.5 text-[11px] text-[var(--text-2)]">
          <UserRound size={12} strokeWidth={1.6} aria-hidden="true" />
          Local
        </span>
      {/if}
      <time datetime={comment.created_at} title={formatDate(comment.created_at)} class="text-[11px] text-[var(--text-3)]">
        {relativeTime(comment.created_at)}
      </time>
      {#if edited}
        <span class="text-[11px] text-[var(--text-3)]" title={formatDate(comment.updated_at)}>· edited</span>
      {/if}
    </div>

    <div class="flex items-center gap-0.5">
      <button
        id={`comment-${comment.id}-edit`}
        type="button"
        class="grid size-8 place-items-center rounded-lg text-[var(--text-3)] transition-colors hover:bg-[var(--surface-2)] hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        aria-label={`Edit comment from ${comment.agent_id ?? 'local user'}`}
        title="Edit comment"
        onclick={onedit}
      >
        <Pencil size={14} strokeWidth={1.6} aria-hidden="true" />
      </button>
      <button
        type="button"
        class="grid size-8 place-items-center rounded-lg text-[var(--text-3)] transition-colors hover:bg-destructive/10 hover:text-destructive focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        aria-label={`Delete comment from ${comment.agent_id ?? 'local user'}`}
        title="Delete comment"
        onclick={ondelete}
      >
        <Trash2 size={14} strokeWidth={1.6} aria-hidden="true" />
      </button>
    </div>
  </header>

  <div class="mt-2">
    <MarkdownRenderer source={comment.content} />
  </div>
</article>
