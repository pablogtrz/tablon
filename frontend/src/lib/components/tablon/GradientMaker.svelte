<script lang="ts">
  import { logoGradient } from '$lib/logo'
  import { labelClass } from '$lib/ui'

  let {
    from = $bindable('#ffa443'),
    to = $bindable('#8b5cf6'),
    angle = $bindable(135),
  }: {
    from?: string
    to?: string
    angle?: number
  } = $props()

  const angleId = $props.id()
  const preview = $derived(logoGradient(from, to, angle))
</script>

<div class="space-y-3 rounded-xl border border-border bg-[var(--surface-1)] p-3">
  <div class="grid grid-cols-2 gap-3">
    <label class="block">
      <span class={labelClass}>From</span>
      <span class="flex items-center gap-2">
        <input
          type="color"
          bind:value={from}
          class="h-10 w-12 cursor-pointer rounded-[10px] border border-border bg-[var(--surface-1)] p-1"
          aria-label="Gradient start colour"
        />
        <span class="font-mono text-[11px] text-[var(--text-2)]">{from.toUpperCase()}</span>
      </span>
    </label>

    <label class="block">
      <span class={labelClass}>To</span>
      <span class="flex items-center gap-2">
        <input
          type="color"
          bind:value={to}
          class="h-10 w-12 cursor-pointer rounded-[10px] border border-border bg-[var(--surface-1)] p-1"
          aria-label="Gradient end colour"
        />
        <span class="font-mono text-[11px] text-[var(--text-2)]">{to.toUpperCase()}</span>
      </span>
    </label>
  </div>

  <div>
    <div class="mb-1.5 flex items-center justify-between gap-3">
      <label class="text-xs font-medium text-[var(--text-2)]" for={angleId}>Angle</label>
      <span class="tnum text-[11px] text-[var(--text-3)]">{angle}°</span>
    </div>
    <input
      id={angleId}
      type="range"
      min="0"
      max="360"
      step="5"
      bind:value={angle}
      class="w-full cursor-pointer accent-[var(--brand)]"
    />
  </div>

  <div class="flex items-center gap-3">
    <span
      class="h-10 w-20 shrink-0 rounded-[10px] border border-border"
      style="background-image: {preview}"
      aria-hidden="true"
    ></span>
    <p class="text-[11px] leading-relaxed text-[var(--text-3)]">
      Pick two colours and drag the angle. The preview updates live.
    </p>
  </div>
</div>
