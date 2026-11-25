<script lang="ts">
  import { onMount, tick, type Component } from 'svelte'
  import ChevronDown from '@lucide/svelte/icons/chevron-down'
  import Clock from '@lucide/svelte/icons/clock'
  import Dumbbell from '@lucide/svelte/icons/dumbbell'
  import Flag from '@lucide/svelte/icons/flag'
  import Heart from '@lucide/svelte/icons/heart'
  import Leaf from '@lucide/svelte/icons/leaf'
  import Lightbulb from '@lucide/svelte/icons/lightbulb'
  import Plane from '@lucide/svelte/icons/plane'
  import Search from '@lucide/svelte/icons/search'
  import Smile from '@lucide/svelte/icons/smile'
  import SmilePlus from '@lucide/svelte/icons/smile-plus'
  import Utensils from '@lucide/svelte/icons/utensils'
  import X from '@lucide/svelte/icons/x'
  import Button from './Button.svelte'
  import { EMOJI_CATEGORIES, EMOJI_ITEMS } from '$lib/emoji'
  import type { EmojiItem } from '$lib/emoji'
  import { readPreference, writePreference } from '$lib/preferences'
  import { inputClass, labelClass } from '$lib/ui'
  import { cn } from '$lib/utils'

  const uid = $props.id()
  const panelId = `emoji-panel-${uid}`
  const customInputId = `emoji-custom-${uid}`
  const anchorName = `--emoji-anchor-${uid}`
  const supportsPopover =
    typeof HTMLElement !== 'undefined' && Object.hasOwn(HTMLElement.prototype, 'popover')

  const icons: Record<string, Component> = {
    smileys: Smile,
    nature: Leaf,
    food: Utensils,
    activities: Dumbbell,
    travel: Plane,
    objects: Lightbulb,
    symbols: Heart,
    flags: Flag,
  }
  const categories = EMOJI_CATEGORIES.map((category) => ({
    ...category,
    icon: icons[category.key] ?? Smile,
  }))
  const allCategories = [
    { key: 'recent', label: 'Recent', icon: Clock, items: [] as EmojiItem[] },
    ...categories,
  ]
  const labelByEmoji = new Map(EMOJI_ITEMS.map((item) => [item.emoji, item.label]))

  let { value = $bindable('') } = $props()

  let panel = $state<HTMLDivElement>()
  let searchInput = $state<HTMLInputElement>()
  let customInput = $state<HTMLInputElement>()
  let open = $state(false)
  let fallbackOpen = $state(false)
  let query = $state('')
  let activeCategory = $state('smileys')
  let customMode = $state(false)
  let customDraft = $state('')
  let recent = $state<string[]>(readRecent())

  const expanded = $derived(supportsPopover ? open : fallbackOpen)
  const queryValue = $derived(query.trim().toLowerCase())
  const searchResults = $derived(
    queryValue
      ? EMOJI_ITEMS.filter(
          (item) => item.label.toLowerCase().includes(queryValue) || item.emoji === queryValue,
        )
      : [],
  )
  const activeMeta = $derived(
    categories.find((category) => category.key === activeCategory) ?? categories[0],
  )
  const recentItems = $derived(
    recent.map((emoji) => ({ emoji, label: labelByEmoji.get(emoji) ?? emoji })),
  )

  function readRecent(): string[] {
    const raw = readPreference('emoji.recent')
    if (!raw) return []
    try {
      const parsed: unknown = JSON.parse(raw)
      if (!Array.isArray(parsed)) return []
      return parsed.filter((item): item is string => typeof item === 'string').slice(0, 12)
    } catch {
      return []
    }
  }

  function remember(emoji: string) {
    recent = [emoji, ...recent.filter((item) => item !== emoji)].slice(0, 12)
    writePreference('emoji.recent', JSON.stringify(recent))
  }

  function closePanel() {
    if (supportsPopover) {
      panel?.hidePopover?.()
    } else {
      fallbackOpen = false
    }
  }

  function select(emoji: string) {
    value = emoji
    remember(emoji)
    closePanel()
  }

  function toggleFallback() {
    if (supportsPopover) return
    fallbackOpen = !fallbackOpen
    if (fallbackOpen) void tick().then(() => searchInput?.focus())
  }

  function openCustom() {
    customDraft = value
    customMode = true
    void tick().then(() => customInput?.focus())
  }

  function applyCustom() {
    const next = customDraft.trim()
    if (!next) return
    value = next
    customMode = false
    closePanel()
  }

  function onCustomKeydown(event: KeyboardEvent) {
    if (event.key !== 'Enter' || event.isComposing) return
    event.preventDefault()
    applyCustom()
  }

  function onSearchKeydown(event: KeyboardEvent) {
    if (event.key !== 'Enter' || event.isComposing) return
    event.preventDefault()
  }

  function onToggle(event: Event) {
    open = (event as ToggleEvent).newState === 'open'
    if (open) {
      void tick().then(() => searchInput?.focus())
    } else {
      customMode = false
      query = ''
    }
  }

  onMount(() => {
    if (supportsPopover) return
    const onKeydown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') fallbackOpen = false
    }
    window.addEventListener('keydown', onKeydown)
    return () => window.removeEventListener('keydown', onKeydown)
  })
</script>

{#snippet emojiGrid(items: EmojiItem[])}
  <div class="grid grid-cols-6 gap-1 sm:grid-cols-8">
    {#each items as item (item.emoji)}
      <button
        type="button"
        class={cn(
          'grid size-10 place-items-center rounded-lg text-xl leading-none transition-colors hover:bg-[var(--surface-2)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:size-9',
          value === item.emoji && 'bg-[var(--surface-2)] ring-1 ring-[var(--ring)]',
        )}
        aria-label={item.label}
        aria-pressed={value === item.emoji}
        onclick={() => select(item.emoji)}
      >
        {item.emoji}
      </button>
    {/each}
  </div>
{/snippet}

<div>
  <span class={labelClass}>Emoji</span>
  <div class="flex items-center gap-2">
    <button
      type="button"
      popovertarget={supportsPopover ? panelId : undefined}
      aria-controls={panelId}
      aria-expanded={expanded}
      aria-haspopup="dialog"
      style="anchor-name: {anchorName}"
      class="flex h-10 min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-[10px] border border-border bg-[var(--surface-1)] px-3 text-left text-sm text-foreground transition-colors hover:bg-[var(--surface-2)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:h-9"
      onclick={toggleFallback}
    >
      {#if value}
        <span class="text-lg leading-none">{value}</span>
        <span>Change emoji</span>
      {:else}
        <Smile size={16} strokeWidth={1.6} class="text-[var(--text-3)]" aria-hidden="true" />
        <span class="text-[var(--text-2)]">Choose emoji</span>
      {/if}
      <ChevronDown size={14} strokeWidth={1.6} class="ml-auto text-[var(--text-3)]" aria-hidden="true" />
    </button>

    {#if value}
      <button
        type="button"
        class="grid size-10 shrink-0 cursor-pointer place-items-center rounded-[10px] text-[var(--text-3)] transition-colors hover:bg-[var(--surface-2)] hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:size-9"
        aria-label="Remove emoji"
        onclick={() => (value = '')}
      >
        <X size={16} strokeWidth={1.6} aria-hidden="true" />
      </button>
    {/if}
  </div>
</div>

<div
  bind:this={panel}
  id={panelId}
  popover={supportsPopover ? 'auto' : undefined}
  hidden={!supportsPopover && !fallbackOpen}
  role="dialog"
  aria-label="Emoji picker"
  class="emoji-popover overflow-hidden rounded-2xl border border-[var(--border-strong)] bg-[var(--bg-2)] p-0 text-foreground shadow-[var(--shadow-lg)]"
  style="position-anchor: {anchorName};"
  ontoggle={onToggle}
>
  <div class="flex max-h-[min(26rem,70dvh)] flex-col">
    <header class="shrink-0 border-b border-border">
      <div class="flex items-center gap-1 px-4 pt-3">
        <span class="flex items-center gap-1.5 border-b-2 border-brand px-2 pb-2 text-sm font-medium text-foreground">
          <Smile size={16} strokeWidth={1.6} class="text-primary" aria-hidden="true" />
          Emojis
        </span>
      </div>

      <div class="px-4 py-3">
        <label class="relative block">
          <span class="sr-only">Search emoji</span>
          <Search
            size={15}
            strokeWidth={1.6}
            class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-3)]"
            aria-hidden="true"
          />
          <input
            bind:this={searchInput}
            type="search"
            class={cn(inputClass, 'h-10 pl-9 text-sm')}
            bind:value={query}
            placeholder="Search emoji"
            autocomplete="off"
            onkeydown={onSearchKeydown}
          />
        </label>
      </div>

      <div class="flex gap-1 overflow-x-auto px-3 pb-2" role="group" aria-label="Emoji categories">
        {#each allCategories as category (category.key)}
          {@const Icon = category.icon}
          <button
            type="button"
            class={cn(
              'grid size-9 shrink-0 cursor-pointer place-items-center rounded-lg text-[var(--text-3)] transition-colors hover:bg-[var(--surface-2)] hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
              !query.trim() &&
                activeCategory === category.key &&
                'bg-[var(--surface-2)] text-foreground',
            )}
            aria-label={category.label}
            title={category.label}
            aria-pressed={!query.trim() && activeCategory === category.key}
            onclick={() => {
              query = ''
              activeCategory = category.key
            }}
          >
            <Icon size={16} strokeWidth={1.7} aria-hidden="true" />
          </button>
        {/each}
      </div>
    </header>

    <div class="min-h-0 flex-1 overflow-y-auto px-4 pb-3">
      {#if queryValue}
        <section class="pt-2" aria-labelledby={`${panelId}-search`}>
          <h3 id={`${panelId}-search`} class="py-2 text-[11px] font-medium text-[var(--text-3)]">
            Search results
          </h3>
          {#if searchResults.length}
            {@render emojiGrid(searchResults)}
          {:else}
            <p class="py-6 text-center text-xs text-[var(--text-3)]">No emoji found.</p>
          {/if}
        </section>
      {:else if activeCategory === 'recent'}
        <section class="pt-2" aria-labelledby={`${panelId}-recent`}>
          <h3 id={`${panelId}-recent`} class="py-2 text-[11px] font-medium text-[var(--text-3)]">
            Recent
          </h3>
          {#if recentItems.length}
            {@render emojiGrid(recentItems)}
          {:else}
            <p class="py-6 text-center text-xs text-[var(--text-3)]">No recent emoji yet.</p>
          {/if}
        </section>
      {:else}
        {#if recentItems.length}
          <section class="pt-2" aria-labelledby={`${panelId}-recent-top`}>
            <h3 id={`${panelId}-recent-top`} class="py-2 text-[11px] font-medium text-[var(--text-3)]">
              Recent
            </h3>
            {@render emojiGrid(recentItems)}
          </section>
        {/if}
        <section class="pt-2" aria-labelledby={`${panelId}-active`}>
          <h3 id={`${panelId}-active`} class="py-2 text-[11px] font-medium text-[var(--text-3)]">
            {activeMeta.label}
          </h3>
          {@render emojiGrid(activeMeta.items)}
        </section>
      {/if}
    </div>

    <footer class="shrink-0 border-t border-border px-4 py-3">
      {#if customMode}
        <label class="sr-only" for={customInputId}>Custom emoji</label>
        <div class="flex items-center gap-2">
          <input
            bind:this={customInput}
            id={customInputId}
            class={cn(inputClass, 'h-9 min-w-0 flex-1 text-xs')}
            bind:value={customDraft}
            maxlength="16"
            placeholder="Paste or type an emoji"
            autocomplete="off"
            onkeydown={onCustomKeydown}
          />
          <Button class="h-9 px-3 text-xs" disabled={!customDraft.trim()} onclick={applyCustom}>
            Use
          </Button>
        </div>
      {:else}
        <div class="flex items-center justify-between gap-3">
          <span class="text-xs text-[var(--text-2)]">Select an emoji</span>
          <Button variant="ghost" class="h-8 px-2 text-xs" onclick={openCustom}>
            <SmilePlus size={14} strokeWidth={1.6} aria-hidden="true" />
            Custom emoji
          </Button>
        </div>
      {/if}
    </footer>
  </div>
</div>
