<script lang="ts">
  import { cn } from '$lib/utils'
  import type { SelectOption } from '$lib/meta'

  let {
    value,
    options,
    id = undefined,
    ariaLabel = undefined,
    disabled = false,
    class: className,
    onchange,
  }: {
    value: string
    options: SelectOption[]
    id?: string
    ariaLabel?: string
    disabled?: boolean
    class?: string
    onchange?: (value: string) => void
  } = $props()

  const selectedTone = $derived(
    options.find((option) => option.value === value)?.tone ?? 'var(--text-3)',
  )

  function handleChange(event: Event) {
    const next = (event.currentTarget as HTMLSelectElement).value
    onchange?.(next)
  }
</script>

<select
  {id}
  class={cn('rich-select', className)}
  {disabled}
  value={value}
  aria-label={ariaLabel}
  style="--tone: {selectedTone}"
  onchange={handleChange}
>
  <button>
    <selectedcontent></selectedcontent>
  </button>
  {#each options as option (option.value)}
    {@const OptionIcon = option.icon}
    <option value={option.value} style="--tone: {option.tone ?? 'var(--text-3)'}">
      {#if OptionIcon}
        <OptionIcon size={14} strokeWidth={1.6} aria-hidden="true" />
      {:else}
        <span class="select-option-dot" aria-hidden="true"></span>
      {/if}
      <span class="select-option-label">{option.label}</span>
    </option>
  {/each}
</select>
