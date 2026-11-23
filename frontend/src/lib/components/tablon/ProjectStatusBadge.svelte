<script lang="ts">
  import ToneBadge from './ToneBadge.svelte'
  import { PROJECT_STATUS_META } from '$lib/meta'
  import type { ProjectStatus } from '$lib/types'

  let { status, class: className }: { status: ProjectStatus; class?: string } = $props()
  const meta = $derived(PROJECT_STATUS_META[status])
</script>

<ToneBadge label={meta.label} tone={meta.tone} icon={meta.icon} class={className} />
