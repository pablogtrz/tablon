<script lang="ts">
  import { onMount, tick } from 'svelte'
  import MessageSquare from '@lucide/svelte/icons/message-square'
  import Button from './Button.svelte'
  import CommentItem from './CommentItem.svelte'
  import ConfirmModal from './ConfirmModal.svelte'
  import MarkdownEditor from './MarkdownEditor.svelte'
  import UnsavedChangesDialog from './UnsavedChangesDialog.svelte'
  import {
    createComment,
    deleteComment,
    listAllComments,
    listComments,
    updateComment,
  } from '$lib/api'
  import { messageOf } from '$lib/errors'
  import { newIdempotencyKey } from '$lib/idempotency'
  import { plural } from '$lib/format'
  import type { Comment } from '$lib/types'

  let { issueRef, refreshKey = 0 }: { issueRef: string; refreshKey?: number } = $props()

  let comments = $state<Comment[]>([])
  let loading = $state(true)
  let error = $state('')
  let loadMoreCursor = $state<string | undefined>()
  let loadingMore = $state(false)

  let draft = $state('')
  let submitKey = $state('')
  let submitting = $state(false)
  let submitError = $state('')

  let editingId = $state<string | null>(null)
  let editDraft = $state('')
  let editSaving = $state(false)
  let editError = $state('')
  let cancelEditOpen = $state(false)

  let deleteTarget = $state<Comment | null>(null)
  let deleteOpen = $state(false)
  let deleting = $state(false)
  let deleteError = $state('')

  let announcement = $state('')
  let composer = $state<HTMLElement>()

  const editing = $derived(comments.find((comment) => comment.id === editingId) ?? null)

  async function announce(message: string) {
    announcement = ''
    await tick()
    announcement = message
  }

  async function load(initial = true) {
    if (initial) loading = true
    error = ''
    try {
      const result = await listAllComments(issueRef)
      comments = result.items
      loadMoreCursor = result.truncated ? result.nextCursor : undefined
    } catch (err) {
      error = messageOf(err)
    } finally {
      if (initial) loading = false
    }
  }

  async function loadMore() {
    if (!loadMoreCursor || loadingMore) return
    loadingMore = true
    error = ''
    try {
      const page = await listComments(issueRef, { limit: 200, cursor: loadMoreCursor })
      const known = new Set(comments.map((comment) => comment.id))
      comments = [...comments, ...page.items.filter((comment) => !known.has(comment.id))]
      loadMoreCursor = page.next_cursor
    } catch (err) {
      error = messageOf(err)
    } finally {
      loadingMore = false
    }
  }

  $effect(() => {
    void draft
    submitKey = ''
  })

  async function submit() {
    const content = draft.trim()
    if (!content || submitting) return
    submitting = true
    submitError = ''
    submitKey ||= newIdempotencyKey()
    try {
      const created = await createComment(issueRef, { content: draft }, submitKey)
      comments = [...comments, created]
      draft = ''
      submitKey = ''
      await announce('Comment added')
      await tick()
      document.getElementById(`comment-${created.id}`)?.focus()
    } catch (err) {
      submitError = messageOf(err)
    } finally {
      submitting = false
    }
  }

  function startEdit(comment: Comment) {
    editingId = comment.id
    editDraft = comment.content
    editError = ''
  }

  async function saveEdit(comment: Comment) {
    if (editSaving) return
    if (!editDraft.trim()) {
      editError = 'Comment cannot be empty.'
      return
    }
    if (editDraft === comment.content) {
      editingId = null
      return
    }
    editSaving = true
    editError = ''
    try {
      const updated = await updateComment(comment.id, { content: editDraft })
      comments = comments.map((item) => (item.id === updated.id ? updated : item))
      editingId = null
      await announce('Comment updated')
      await tick()
      document.getElementById(`comment-${comment.id}-edit`)?.focus()
    } catch (err) {
      editError = messageOf(err)
    } finally {
      editSaving = false
    }
  }

  function cancelEdit(comment: Comment) {
    if (editDraft !== comment.content) {
      cancelEditOpen = true
      return
    }
    editingId = null
  }

  function discardEdit() {
    editingId = null
    editDraft = ''
    editError = ''
    cancelEditOpen = false
  }

  function askDelete(comment: Comment) {
    deleteTarget = comment
    deleteError = ''
    deleteOpen = true
  }

  async function confirmDelete() {
    if (!deleteTarget || deleting) return
    const index = comments.findIndex((comment) => comment.id === deleteTarget?.id)
    deleting = true
    deleteError = ''
    try {
      await deleteComment(deleteTarget.id)
      comments = comments.filter((comment) => comment.id !== deleteTarget?.id)
      deleteOpen = false
      deleteTarget = null
      await announce('Comment deleted')
      await tick()
      const next = comments[index] ?? comments[index - 1]
      if (next) {
        document.getElementById(`comment-${next.id}`)?.focus()
      } else {
        composer?.focus()
      }
    } catch (err) {
      deleteError = messageOf(err)
    } finally {
      deleting = false
    }
  }

  let seenRefresh = $state<number | null>(null)

  $effect(() => {
    if (seenRefresh === null) {
      seenRefresh = refreshKey
      return
    }
    if (refreshKey === seenRefresh) return
    seenRefresh = refreshKey
    void load(false)
  })

  onMount(() => {
    void load()
  })
</script>

<section class="mt-5" aria-labelledby="comments-heading">
  <div class="flex flex-wrap items-center justify-between gap-2">
    <h2 id="comments-heading" class="flex items-center gap-2 text-sm font-semibold">
      <MessageSquare size={15} strokeWidth={1.6} class="text-[var(--text-3)]" aria-hidden="true" />
      Comments
      {#if !loading && comments.length > 0}
        <span class="rounded-full bg-[var(--surface-2)] px-1.5 py-0.5 text-[10px] tnum text-[var(--text-3)]">
          {plural(comments.length, 'comment')}
        </span>
      {/if}
    </h2>
  </div>

  {#if loading}
    <div class="mt-3 space-y-2" aria-hidden="true">
      {#each Array(3) as _}
        <div class="h-20 animate-pulse rounded-xl border border-border bg-[var(--surface-1)] motion-reduce:animate-none"></div>
      {/each}
    </div>
  {:else if error && comments.length === 0}
    <div class="mt-3 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
      {error}
      <Button variant="secondary" class="ml-2 h-7 px-2 text-xs" onclick={() => void load()}>Retry</Button>
    </div>
  {:else}
    {#if comments.length === 0}
      <p class="mt-3 rounded-xl border border-dashed border-border px-4 py-6 text-center text-xs leading-relaxed text-[var(--text-3)]">
        No comments yet. Add the first update.
      </p>
    {:else}
      <ol class="mt-3 space-y-2">
        {#each comments as comment (comment.id)}
          <li class="list-none">
            {#if editingId === comment.id}
              <div class="rounded-[10px]">
                <MarkdownEditor
                  bind:value={editDraft}
                  ariaLabel="Edit comment"
                  dirty={editDraft !== comment.content}
                  saving={editSaving}
                  error={editError}
                  submitLabel="Save comment"
                  compact
                  onsave={() => void saveEdit(comment)}
                  oncancel={() => cancelEdit(comment)}
                />
              </div>
            {:else}
              <CommentItem
                {comment}
                onedit={() => startEdit(comment)}
                ondelete={() => askDelete(comment)}
              />
            {/if}
          </li>
        {/each}
      </ol>
    {/if}

    {#if loadMoreCursor}
      <div class="mt-2 flex justify-center">
        <Button variant="secondary" class="text-xs" disabled={loadingMore} onclick={() => void loadMore()}>
          {loadingMore ? 'Loading…' : 'Load more comments'}
        </Button>
      </div>
    {/if}

    {#if error}
      <p class="mt-2 text-xs text-destructive" role="alert">{error}</p>
    {/if}

    <div id="comment-composer" bind:this={composer} tabindex="-1" class="mt-4 outline-none">
      <MarkdownEditor
        bind:value={draft}
        ariaLabel="Add comment"
        dirty={draft.trim().length > 0}
        saving={submitting}
        error={submitError}
        submitLabel="Add comment"
        placeholder="Add a comment with Markdown…"
        compact
        onsave={() => void submit()}
        oncancel={() => {
          draft = ''
          submitError = ''
        }}
      />
    </div>
  {/if}

  <p class="sr-only" role="status" aria-live="polite">{announcement}</p>
</section>

{#if editing}
  <UnsavedChangesDialog
    bind:open={cancelEditOpen}
    title="Discard comment changes?"
    description="Your comment edits have not been saved."
    ondiscard={discardEdit}
    onstay={() => (cancelEditOpen = false)}
  />
{/if}

<ConfirmModal
  bind:open={deleteOpen}
  title="Delete comment"
  description="Delete this comment? This cannot be undone."
  confirmLabel="Delete comment"
  busy={deleting}
  error={deleteError}
  onconfirm={() => void confirmDelete()}
/>
