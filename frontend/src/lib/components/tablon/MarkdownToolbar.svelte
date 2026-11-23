<script lang="ts">
  import Bold from '@lucide/svelte/icons/bold'
  import Code from '@lucide/svelte/icons/code'
  import CodeXml from '@lucide/svelte/icons/code-xml'
  import Heading2 from '@lucide/svelte/icons/heading-2'
  import Image from '@lucide/svelte/icons/image'
  import Italic from '@lucide/svelte/icons/italic'
  import Link2 from '@lucide/svelte/icons/link-2'
  import List from '@lucide/svelte/icons/list'
  import ListOrdered from '@lucide/svelte/icons/list-ordered'
  import ListTodo from '@lucide/svelte/icons/list-todo'
  import Quote from '@lucide/svelte/icons/quote'
  import Strikethrough from '@lucide/svelte/icons/strikethrough'
  import Table from '@lucide/svelte/icons/table'
  import type { Component } from 'svelte'
  import type { MarkdownAction } from '$lib/markdown'

  let {
    onaction,
    disabled = false,
  }: {
    onaction: (action: MarkdownAction) => void
    disabled?: boolean
  } = $props()

  interface ToolbarItem {
    action: MarkdownAction
    label: string
    icon: Component
    shortcut?: string
  }

  const inline: ToolbarItem[] = [
    { action: 'bold', label: 'Bold', icon: Bold, shortcut: 'Cmd/Ctrl+B' },
    { action: 'italic', label: 'Italic', icon: Italic, shortcut: 'Cmd/Ctrl+I' },
    { action: 'strikethrough', label: 'Strikethrough', icon: Strikethrough },
    { action: 'code', label: 'Inline code', icon: Code },
  ]

  const blocks: ToolbarItem[] = [
    { action: 'heading', label: 'Heading', icon: Heading2 },
    { action: 'bullet', label: 'Bullet list', icon: List },
    { action: 'numbered', label: 'Numbered list', icon: ListOrdered },
    { action: 'task', label: 'Task list', icon: ListTodo },
    { action: 'quote', label: 'Quote', icon: Quote },
    { action: 'codeblock', label: 'Code block', icon: CodeXml },
  ]

  const inserts: ToolbarItem[] = [
    { action: 'link', label: 'Link', icon: Link2, shortcut: 'Cmd/Ctrl+K' },
    { action: 'image', label: 'Image', icon: Image },
    { action: 'table', label: 'Table', icon: Table },
  ]

  const groups = [inline, blocks, inserts]

  const buttonClass =
    'grid size-8 place-items-center rounded-lg text-[var(--text-2)] transition-colors hover:bg-[var(--surface-2)] hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-40'
</script>

<div role="toolbar" aria-label="Markdown formatting" class="flex flex-wrap items-center gap-0.5">
  {#each groups as group, groupIndex}
    {#if groupIndex > 0}
      <span class="mx-1 h-5 w-px shrink-0 bg-border" aria-hidden="true"></span>
    {/if}
    {#each group as item}
      {@const Icon = item.icon}
      <button
        type="button"
        class={buttonClass}
        {disabled}
        aria-label={item.label}
        title={item.shortcut ? `${item.label} (${item.shortcut})` : item.label}
        onclick={() => onaction(item.action)}
      >
        <Icon size={15} strokeWidth={1.6} aria-hidden="true" />
      </button>
    {/each}
  {/each}
</div>
