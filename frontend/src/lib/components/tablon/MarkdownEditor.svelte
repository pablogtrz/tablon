<script lang="ts">
  import { onMount, tick } from 'svelte'
  import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
  import { Compartment, EditorState } from '@codemirror/state'
  import {
    drawSelection,
    EditorView,
    highlightSpecialChars,
    keymap,
    placeholder as editorPlaceholder,
  } from '@codemirror/view'
  import Button from './Button.svelte'
  import CodeBlockDialog from './CodeBlockDialog.svelte'
  import MarkdownRenderer from './MarkdownRenderer.svelte'
  import MarkdownToolbar from './MarkdownToolbar.svelte'
  import Modal from './Modal.svelte'
  import { markdownEditorTheme, markdownLineDecorations } from '$lib/editor-theme'
  import {
    buildCodeFence,
    insertText as insertTextAt,
    prefixLines as prefixLinesText,
    replaceText,
    wrapSelection as wrapSelectionText,
    type EditResult,
    type Selection,
  } from '$lib/markdown-editor'
  import { toggleTask, type MarkdownAction } from '$lib/markdown'
  import { inputClass, labelClass } from '$lib/ui'

  let {
    value = $bindable(''),
    dirty = false,
    saving = false,
    error = '',
    submitLabel = 'Save description',
    placeholder = 'Write with Markdown, or use the toolbar above.',
    compact = false,
    ariaLabel = 'Markdown editor',
    ariaDescribedBy = undefined,
    onsave,
    oncancel,
  }: {
    value?: string
    dirty?: boolean
    saving?: boolean
    error?: string
    submitLabel?: string
    placeholder?: string
    compact?: boolean
    ariaLabel?: string
    ariaDescribedBy?: string
    onsave: (value: string) => void
    oncancel: () => void
  } = $props()

  let mode = $state<'write' | 'preview'>('write')

  function onModeKeydown(event: KeyboardEvent) {
    if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
    event.preventDefault()
    const next: 'write' | 'preview' =
      event.key === 'Home'
        ? 'write'
        : event.key === 'End'
          ? 'preview'
          : mode === 'write'
            ? 'preview'
            : 'write'
    mode = next
    document.getElementById(`editor-tab-${next}`)?.focus()
  }
  let host = $state<HTMLDivElement>()
  let view: EditorView | null = null
  let syncing = false
  const readOnly = new Compartment()

  let linkOpen = $state(false)
  let linkText = $state('')
  let linkUrl = $state('')
  let linkError = $state('')
  let linkRange = $state<Selection | null>(null)
  let linkWasOpen = $state(false)

  let imageOpen = $state(false)
  let imageAlt = $state('')
  let imageUrl = $state('')
  let imageError = $state('')
  let imageRange = $state<Selection | null>(null)
  let imageWasOpen = $state(false)

  let codeBlockOpen = $state(false)
  let codeBlockRange = $state<Selection | null>(null)
  let codeBlockWasOpen = $state(false)

  function currentValue(): string {
    return view?.state.doc.toString() ?? value
  }

  function selection(): Selection {
    if (!view) return { from: 0, to: 0 }
    const range = view.state.selection.main
    return { from: range.from, to: range.to }
  }

  function applyEdit(result: EditResult) {
    if (!view) {
      value = result.value
      return
    }
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: result.value },
      selection: { anchor: result.from, head: result.to },
      scrollIntoView: true,
    })
    view.focus()
  }

  function replaceRange(
    from: number,
    to: number,
    text: string,
    selectFrom = from + text.length,
    selectTo = selectFrom,
  ) {
    if (!view) {
      value = replaceText(value, from, to, text, selectFrom, selectTo).value
      return
    }
    const result = replaceText(currentValue(), from, to, text, selectFrom, selectTo)
    view.dispatch({
      changes: { from, to, insert: text },
      selection: { anchor: result.from, head: result.to },
      scrollIntoView: true,
    })
    view.focus()
  }

  function wrapSelectionBeforeAfter(before: string, after: string, placeholderText: string) {
    const { from, to } = selection()
    applyEdit(wrapSelectionText(currentValue(), from, to, before, after, placeholderText))
  }

  function prefixSelectionLines(prefix: string | ((index: number) => string)) {
    const { from, to } = selection()
    applyEdit(prefixLinesText(currentValue(), from, to, prefix))
  }

  function insertText(text: string) {
    const { from, to } = selection()
    applyEdit(insertTextAt(currentValue(), from, to, text))
  }

  function openLink() {
    const range = selection()
    linkRange = range
    linkText = currentValue().slice(range.from, range.to)
    linkUrl = ''
    linkError = ''
    linkOpen = true
  }

  function openImage() {
    const range = selection()
    imageRange = range
    imageAlt = ''
    imageUrl = ''
    imageError = ''
    imageOpen = true
  }

  function openCodeBlock() {
    codeBlockRange = selection()
    codeBlockOpen = true
  }

  function insertCodeBlock(language: string) {
    const range = codeBlockRange ?? selection()
    const built = buildCodeFence(currentValue(), range.from, range.to, language)
    replaceRange(range.from, range.to, built.text, built.selectFrom, built.selectTo)
    codeBlockRange = null
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

  function insertLink() {
    if (!linkRange) return
    const url = normalizeUrl(linkUrl, false)
    if (!url) {
      linkError = 'Enter a valid URL.'
      return
    }
    const label = linkText.trim() || 'link text'
    replaceRange(linkRange.from, linkRange.to, `[${label}](${url})`)
    linkOpen = false
  }

  function insertImage() {
    if (!imageRange) return
    const url = normalizeUrl(imageUrl, true)
    if (!url) {
      imageError = 'Use an https:// image URL.'
      return
    }
    const alt = imageAlt.trim() || 'image'
    replaceRange(imageRange.from, imageRange.to, `![${alt}](${url})`)
    imageOpen = false
  }

  function applyAction(action: MarkdownAction) {
    switch (action) {
      case 'bold':
        wrapSelectionBeforeAfter('**', '**', 'bold text')
        break
      case 'italic':
        wrapSelectionBeforeAfter('*', '*', 'italic text')
        break
      case 'strikethrough':
        wrapSelectionBeforeAfter('~~', '~~', 'strikethrough')
        break
      case 'heading':
        prefixSelectionLines('## ')
        break
      case 'bullet':
        prefixSelectionLines('- ')
        break
      case 'numbered':
        prefixSelectionLines((index) => `${index + 1}. `)
        break
      case 'task':
        prefixSelectionLines('- [ ] ')
        break
      case 'quote':
        prefixSelectionLines('> ')
        break
      case 'code':
        wrapSelectionBeforeAfter('`', '`', 'code')
        break
      case 'codeblock':
        openCodeBlock()
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

  function handleTaskToggle(index: number) {
    value = toggleTask(currentValue(), index)
  }

  function save() {
    if (dirty && !saving) onsave(currentValue())
  }

  onMount(() => {
    if (!host) return

    const attributes: Record<string, string> = {
      'aria-label': ariaLabel,
      autocomplete: 'off',
      autocapitalize: 'off',
      autocorrect: 'off',
      spellcheck: 'false',
    }
    if (ariaDescribedBy) attributes['aria-describedby'] = ariaDescribedBy

    view = new EditorView({
      parent: host,
      state: EditorState.create({
        doc: value,
        extensions: [
          history(),
          drawSelection(),
          highlightSpecialChars(),
          EditorView.lineWrapping,
          editorPlaceholder(placeholder),
          EditorView.contentAttributes.of(attributes),
          markdownEditorTheme,
          markdownLineDecorations,
          readOnly.of(EditorState.readOnly.of(saving)),
          keymap.of([
            {
              key: 'Mod-b',
              run: () => {
                applyAction('bold')
                return true
              },
            },
            {
              key: 'Mod-i',
              run: () => {
                applyAction('italic')
                return true
              },
            },
            {
              key: 'Mod-k',
              run: () => {
                openLink()
                return true
              },
            },
            {
              key: 'Mod-Enter',
              run: () => {
                save()
                return true
              },
            },
            {
              key: 'Escape',
              run: () => {
                oncancel()
                return true
              },
            },
            ...defaultKeymap,
            ...historyKeymap,
          ]),
          EditorView.updateListener.of((update) => {
            if (update.docChanged && !syncing) {
              value = update.state.doc.toString()
            }
          }),
        ],
      }),
    })

    return () => {
      view?.destroy()
      view = null
    }
  })

  $effect(() => {
    const next = value
    const editor = view
    if (!editor || editor.state.doc.toString() === next) return
    syncing = true
    const length = next.length
    const range = editor.state.selection.main
    editor.dispatch({
      changes: { from: 0, to: editor.state.doc.length, insert: next },
      selection: {
        anchor: Math.min(range.anchor, length),
        head: Math.min(range.head, length),
      },
    })
    syncing = false
  })

  $effect(() => {
    const editor = view
    if (!editor) return
    editor.dispatch({
      effects: readOnly.reconfigure(EditorState.readOnly.of(saving)),
    })
  })

  $effect(() => {
    if (mode !== 'write') return
    void tick().then(() => view?.requestMeasure())
  })

  $effect(() => {
    if (codeBlockOpen) {
      codeBlockWasOpen = true
      return
    }
    if (!codeBlockWasOpen) return
    codeBlockWasOpen = false
    void tick().then(() => view?.focus())
  })

  $effect(() => {
    if (linkOpen) {
      linkWasOpen = true
      return
    }
    if (!linkWasOpen) return
    linkWasOpen = false
    void tick().then(() => view?.focus())
  })

  $effect(() => {
    if (imageOpen) {
      imageWasOpen = true
      return
    }
    if (!imageWasOpen) return
    imageWasOpen = false
    void tick().then(() => view?.focus())
  })
</script>

<div class="overflow-hidden rounded-[10px] border border-brand/60 bg-[var(--surface-1)] shadow-[0_0_0_3px_rgba(255,164,67,.10)]" style="--editor-min-h: {compact ? '8rem' : '16rem'}">
  <div class="flex flex-wrap items-center justify-between gap-2 border-b border-border px-2 py-1.5">
    <MarkdownToolbar onaction={applyAction} disabled={saving || mode === 'preview'} />
    <div class="flex items-center gap-0.5 rounded-lg border border-border bg-[var(--bg-2)] p-0.5" role="tablist" aria-label="Editor view">
      <button
        type="button"
        role="tab"
        id="editor-tab-write"
        aria-controls="editor-panel-write"
        onkeydown={onModeKeydown}
        aria-selected={mode === 'write'}
        tabindex={mode === 'write' ? 0 : -1}
        class="rounded-md px-2 py-1 text-[11px] transition-colors {mode === 'write' ? 'bg-[var(--surface-2)] text-foreground' : 'text-[var(--text-3)] hover:text-foreground'} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onclick={() => (mode = 'write')}
      >
        Write
      </button>
      <button
        type="button"
        role="tab"
        id="editor-tab-preview"
        aria-controls="editor-panel-preview"
        onkeydown={onModeKeydown}
        aria-selected={mode === 'preview'}
        tabindex={mode === 'preview' ? 0 : -1}
        class="rounded-md px-2 py-1 text-[11px] transition-colors {mode === 'preview' ? 'bg-[var(--surface-2)] text-foreground' : 'text-[var(--text-3)] hover:text-foreground'} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onclick={() => (mode = 'preview')}
      >
        Preview
      </button>
    </div>
  </div>

  <div
    id="editor-panel-write"
    role="tabpanel"
    aria-label="Markdown source"
    aria-labelledby="editor-tab-write"
    class={mode === 'write' ? 'block' : 'hidden'}
  >
    <div
      bind:this={host}
      data-testid="markdown-editor"
      class="markdown-editor"
    ></div>
  </div>

  <div
    id="editor-panel-preview"
    role="tabpanel"
    aria-label="Markdown preview"
    aria-labelledby="editor-tab-preview"
    class={mode === 'preview' ? 'block' : 'hidden'}
  >
    <div class="px-3 py-3" style="min-height: var(--editor-min-h)">
      <MarkdownRenderer source={value} interactive ontasktoggle={handleTaskToggle} />
    </div>
  </div>

  {#if error}
    <p class="mx-3 mb-2 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive" role="alert">{error}</p>
  {/if}

  <div class="flex flex-wrap items-center justify-between gap-3 border-t border-border px-3 py-2">
    <span class="text-[11px] text-[var(--text-3)]">Markdown · Cmd/Ctrl+Enter to save</span>
    <div class="flex gap-2">
      <Button variant="secondary" disabled={saving} onclick={oncancel}>Cancel</Button>
      <Button disabled={!dirty || saving} onclick={save}>{saving ? 'Saving…' : submitLabel}</Button>
    </div>
  </div>
</div>

<CodeBlockDialog bind:open={codeBlockOpen} oninsert={insertCodeBlock} />

<Modal bind:open={linkOpen} title="Insert link" description="Add a link to the description.">
  <form id="markdown-link" class="space-y-4" onsubmit={(event) => { event.preventDefault(); insertLink() }}>
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
  <form id="markdown-image" class="space-y-4" onsubmit={(event) => { event.preventDefault(); insertImage() }}>
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
