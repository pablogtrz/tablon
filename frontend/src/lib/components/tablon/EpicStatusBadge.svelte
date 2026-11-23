<script lang="ts">
  import ToneBadge from './ToneBadge.svelte'
  import { EPIC_STATUS_META } from '$lib/meta'
  import type { EpicStatus } from '$lib/types'

  let { status, class: className }: { status: EpicStatus; class?: string } = $props()
  const meta = $derived(EPIC_STATUS_META[status])
</script>

<ToneBadge label={meta.label} tone={meta.tone} icon={meta.icon} class={className} />
