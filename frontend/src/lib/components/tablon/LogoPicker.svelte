<script lang="ts">
  import Check from '@lucide/svelte/icons/check'
  import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal'
  import GradientMaker from './GradientMaker.svelte'
  import { LOGO_GRADIENT_OPTIONS } from '$lib/meta'
  import { cn } from '$lib/utils'

  const DEFAULT_FROM = '#ffa443'
  const DEFAULT_TO = '#8b5cf6'

  let {
    value = $bindable(0),
    from = $bindable(''),
    to = $bindable(''),
    angle = $bindable(135),
    name = 'logo-gradient',
    class: className,
  }: {
    value?: number
    from?: string
    to?: string
    angle?: number
    name?: string
    class?: string
  } = $props()

  const custom = $derived(Boolean(from && to))
  const customPreview = $derived(
    custom
      ? `linear-gradient(${angle}deg, ${from} 0%, ${to} 100%)`
      : `linear-gradient(135deg, ${DEFAULT_FROM} 0%, ${DEFAULT_TO} 100%)`,
  )

  function chooseCustom() {
    if (from && to) return
    from = DEFAULT_FROM
    to = DEFAULT_TO
    angle = 135
  }

  function chooseCurated(next: number) {
    value = next
    from = ''
    to = ''
  }
</script>

<fieldset class={cn('min-w-0', className)}>
  <legend class="mb-1.5 block text-xs font-medium text-[var(--text-2)]">Colour</legend>
  <div class="flex flex-wrap gap-2">
    {#each LOGO_GRADIENT_OPTIONS as option (option.value)}
      <label class="relative inline-grid cursor-pointer" title={option.label}>
        <input
          type="radio"
          {name}
          value={option.value}
          checked={!custom && value === option.value}
          onchange={() => chooseCurated(option.value)}
          class="peer sr-only"
          aria-label={option.label}
        />
        <span
          class="block size-10 rounded-[10px] border border-[var(--border)] ring-offset-2 ring-offset-[var(--bg)] transition-shadow duration-[120ms] peer-checked:ring-2 peer-checked:ring-[var(--text)] peer-focus-visible:ring-2 peer-focus-visible:ring-ring motion-reduce:transition-none sm:size-9"
          style="background-image: var(--logo-{option.value})"
        ></span>
        <Check
          size={14}
          strokeWidth={2.4}
          class="pointer-events-none absolute inset-0 m-auto hidden text-white peer-checked:block"
          aria-hidden="true"
        />
      </label>
    {/each}

    <label class="inline-flex cursor-pointer items-center gap-2" title="Custom gradient">
      <input
        type="radio"
        {name}
        value="custom"
        checked={custom}
        onchange={chooseCustom}
        class="peer sr-only"
        aria-label="Custom gradient"
      />
      <span
        class="relative inline-grid rounded-[10px] ring-offset-2 ring-offset-[var(--bg)] transition-shadow duration-[120ms] peer-checked:ring-2 peer-checked:ring-[var(--text)] peer-focus-visible:ring-2 peer-focus-visible:ring-ring motion-reduce:transition-none"
      >
        <span
          class="block size-10 rounded-[10px] border border-[var(--border)] sm:size-9"
          style="background-image: {customPreview}"
        ></span>
        {#if custom}
          <Check
            size={14}
            strokeWidth={2.4}
            class="pointer-events-none absolute inset-0 m-auto text-white"
            aria-hidden="true"
          />
        {:else}
          <SlidersHorizontal
            size={14}
            strokeWidth={1.8}
            class="pointer-events-none absolute inset-0 m-auto text-white"
            aria-hidden="true"
          />
        {/if}
      </span>
      <span class="text-xs text-[var(--text-2)]">Custom</span>
    </label>
  </div>

  {#if custom}
    <div class="mt-3">
      <GradientMaker bind:from bind:to bind:angle />
    </div>
  {/if}

  <p class="mt-1.5 text-[11px] text-[var(--text-3)]">
    An emoji, if set, is shown on top of the colour tile.
  </p>
</fieldset>
