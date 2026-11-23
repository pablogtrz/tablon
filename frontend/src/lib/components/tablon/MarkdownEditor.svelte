<script lang="ts">
  import { tick } from 'svelte'
  import Button from './Button.svelte'
  import MarkdownRenderer from './MarkdownRenderer.svelte'
  import MarkdownToolbar from './MarkdownToolbar.svelte'
  import Modal from './Modal.svelte'
  import { toggleTask, type MarkdownAction } from '$lib/markdown'
  import { inputClass, labelClass } from '$lib/ui'

  let {
    value = $bindable(''),
    dirty = false,
    saving = false,
    error = '',
    onsave,
    oncancel,
  }: {
    value?: string
    dirty?: boolean
    saving?: boolean
    error?: string
    onsave: (value: string) => void
    oncancel: () => void
  } = $props()

  let mode = $state<'write' | 'preview'>('write')
  let editor = $state<HTMLTextAreaElement>()

  let linkOpen = $state(false)
  let linkText = $state('')
  let linkUrl = $state('')
  let linkError = $state('')
  let linkRange = $state<{ start: number; end: number } | null>(null)

  let imageOpen = $state(false)
  let imageAlt = $state('')
  let imageUrl = $state('')
  let imageError = $state('')
  let imageRange = $state<{ start: number; end: number } | null>(null)

  async function replaceRange(start: number, end: number, text: string, selectStart?: number, selectEnd?: number) {
    value = value.slice(0, start) + text + value.slice(end)
    await tick()
    editor?.focus()
    const from = selectStart ?? start + text.length
    const to = selectEnd ?? from
    editor?.setSelectionRange(from, to)
  }

  function wrapSelection(before: string, after: string, placeholder: string) {
    if (!editor) return
    const start = editor.selectionStart
    const end = editor.selectionEnd
    const selected = value.slice(start, end)
    const content = selected || placeholder
    void replaceRange(start, end, before + content + after, start + before.length, start + before.length + content.length)
  }

  function prefixLines(prefix: string | ((index: number) => string)) {
    if (!editor) return
    const start = editor.selectionStart
    const end = editor.selectionEnd
    const lineStart = value.lastIndexOf('\n', start - 1) + 1
    const rawLineEnd = value.indexOf('\n', end)
    const lineEnd = rawLineEnd === -1 ? value.length : rawLineEnd
    const lines = value.slice(lineStart, lineEnd).split('\n')
    const next = lines
      .map((line, index) => (typeof prefix === 'function' ? prefix(index) : prefix) + line)
      .join('\n')
    void replaceRange(lineStart, lineEnd, next, lineStart, lineStart + next.length)
  }

  function insertText(text: string) {
    if (!editor) return
    void replaceRange(editor.selectionStart, editor.selectionEnd, text)
  }

  function openLink() {
    if (!editor) return
    linkRange = { start: editor.selectionStart, end: editor.selectionEnd }
    linkText = value.slice(linkRange.start, linkRange.end)
    linkUrl = ''
    linkError = ''
    linkOpen = true
  }

  function openImage() {
    if (!editor) return
    imageRange = { start: editor.selectionStart, end: editor.selectionEnd }
    imageAlt = ''
    imageUrl = ''
    imageError = ''
    imageOpen = true
  }

  function normalizeUrl(raw: string, image: boolean): string {
    const url = raw.trim()
    if (!url) return ''
    if (image) {
      return /^https:\/\//i.test(url) ? url : ''
    }
    if (/^(https?:\/\/|mailto:)/i.test(url) || url.startsWith('#')) return url
    return `https://${url}`
  }

  async function insertLink() {
    if (!linkRange) return
    const url = normalizeUrl(linkUrl, false)
    if (!url) {
      linkError = 'Enter a valid URL.'
      return
    }
    const label = linkText.trim() || 'link text'
    const markdown = `[${label}](${url})`
    await replaceRange(linkRange.start, linkRange.end, markdown)
    linkOpen = false
  }

  async function insertImage() {
    if (!imageRange) return
    const url = normalizeUrl(imageUrl, true)
    if (!url) {
      imageError = 'Use an https:// image URL.'
      return
    }
    const alt = imageAlt.trim() || 'image'
    const markdown = `![${alt}](${url})`
    await replaceRange(imageRange.start, imageRange.end, markdown)
    imageOpen = false
  }

  function applyAction(action: MarkdownAction) {
    switch (action) {
      case 'bold':
        wrapSelection('**', '**', 'bold text')
        break
      case 'italic':
        wrapSelection('*', '*', 'italic text')
        break
      case 'strikethrough':
        wrapSelection('~~', '~~', 'strikethrough')
        break
      case 'heading':
        prefixLines('## ')
        break
      case 'bullet':
        prefixLines('- ')
        break
      case 'numbered':
        prefixLines((index) => `${index + 1}. `)
        break
      case 'task':
        prefixLines('- [ ] ')
        break
      case 'quote':
        prefixLines('> ')
        break
      case 'code':
        wrapSelection('`', '`', 'code')
        break
      case 'codeblock':
        wrapSelection('```\n', '\n```', 'code')
        break
      case 'link':
        openLink()
        break
      case 'image':
        openImage()
        break
      case 'table':
        insertText('\n| Column 1 | Column 2 |\n| --- | --- |\n|  |  |\n')
        break
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.isComposing) return
    if (event.key === 'Escape') {
      event.preventDefault()
      oncancel()
      return
    }
    const mod = event.metaKey || event.ctrlKey
    if (mod && event.key === 'Enter') {
      event.preventDefault()
      if (dirty && !saving) onsave(value)
      return
    }
    if (!mod) return
    const key = event.key.toLowerCase()
    if (key === 'b') {
      event.preventDefault()
      applyAction('bold')
    } else if (key === 'i') {
      event.preventDefault()
      applyAction('italic')
    } else if (key === 'k') {
      event.preventDefault()
      applyAction('link')
    }
  }

  function handleTaskToggle(index: number) {
    value = toggleTask(value, index)
  }

  function save() {
    if (dirty && !saving) onsave(value)
  }
</script>

<div class="overflow-hidden rounded-[10px] border border-brand/60 bg-[var(--surface-1)] shadow-[0_0_0_3px_rgba(255,164,67,.10)]">
  <div class="flex flex-wrap items-center justify-between gap-2 border-b border-border px-2 py-1.5">
    <MarkdownToolbar onaction={applyAction} disabled={saving || mode === 'preview'} />
    <div class="flex items-center gap-0.5 rounded-lg border border-border bg-[var(--bg-2)] p-0.5" role="tablist" aria-label="Editor view">
      <button
        type="button"
        role="tab"
        aria-selected={mode === 'write'}
        class="rounded-md px-2 py-1 text-[11px] transition-colors {mode === 'write' ? 'bg-[var(--surface-2)] text-foreground' : 'text-[var(--text-3)] hover:text-foreground'} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onclick={() => (mode = 'write')}
      >
        Write
      </button>
      <button
        type="button"
        role="tab"
        aria-selected={mode === 'preview'}
        class="rounded-md px-2 py-1 text-[11px] transition-colors {mode === 'preview' ? 'bg-[var(--surface-2)] text-foreground' : 'text-[var(--text-3)] hover:text-foreground'} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onclick={() => (mode = 'preview')}
      >
        Preview
      </button>
    </div>
  </div>

  {#if mode === 'write'}
    <div role="tabpanel" aria-label="Markdown source">
      <textarea
        bind:this={editor}
        bind:value
        readonly={saving}
        class="min-h-64 w-full resize-y border-0 bg-transparent px-3 py-3 font-mono text-[13px] leading-relaxed text-foreground placeholder:text-[var(--text-3)] focus:outline-none"
        placeholder="Write with Markdown, or use the toolbar above."
        onkeydown={handleKeydown}
      ></textarea>
    </div>
  {:else}
    <div role="tabpanel" aria-label="Markdown preview" class="min-h-64 px-3 py-3">
      <MarkdownRenderer source={value} interactive ontasktoggle={handleTaskToggle} />
    </div>
  {/if}

  {#if error}
    <p class="mx-3 mb-2 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{error}</p>
  {/if}

  <div class="flex flex-wrap items-center justify-between gap-3 border-t border-border px-3 py-2">
    <span class="text-[11px] text-[var(--text-3)]">Markdown · Cmd/Ctrl+Enter to save</span>
    <div class="flex gap-2">
      <Button variant="secondary" disabled={saving} onclick={oncancel}>Cancel</Button>
      <Button disabled={!dirty || saving} onclick={save}>{saving ? 'Saving…' : 'Save description'}</Button>
    </div>
  </div>
</div>

<Modal bind:open={linkOpen} title="Insert link" description="Add a link to the description.">
  <form id="markdown-link" class="space-y-4" onsubmit={(event) => { event.preventDefault(); void insertLink() }}>
    <div>
      <label class={labelClass} for="markdown-link-text">Text</label>
      <input id="markdown-link-text" class={inputClass} bind:value={linkText} placeholder="Link text" />
    </div>
    <div>
      <label class={labelClass} for="markdown-link-url">URL</label>
      <input id="markdown-link-url" class={inputClass} bind:value={linkUrl} placeholder="https://example.com" autocomplete="off" />
    </div>
    {#if linkError}
      <p class="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{linkError}</p>
    {/if}
  </form>
  {#snippet footer()}
    <Button variant="secondary" onclick={() => (linkOpen = false)}>Cancel</Button>
    <Button type="submit" form="markdown-link">Insert link</Button>
  {/snippet}
</Modal>

<Modal bind:open={imageOpen} title="Insert image" description="Use an external https:// image URL.">
  <form id="markdown-image" class="space-y-4" onsubmit={(event) => { event.preventDefault(); void insertImage() }}>
    <div>
      <label class={labelClass} for="markdown-image-alt">Alt text</label>
      <input id="markdown-image-alt" class={inputClass} bind:value={imageAlt} placeholder="Describe the image" />
    </div>
    <div>
      <label class={labelClass} for="markdown-image-url">Image URL</label>
      <input id="markdown-image-url" class={inputClass} bind:value={imageUrl} placeholder="https://example.com/image.png" autocomplete="off" />
    </div>
    {#if imageError}
      <p class="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{imageError}</p>
    {/if}
  </form>
  {#snippet footer()}
    <Button variant="secondary" onclick={() => (imageOpen = false)}>Cancel</Button>
    <Button type="submit" form="markdown-image">Insert image</Button>
  {/snippet}
</Modal>
