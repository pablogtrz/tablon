<script lang="ts">
  import { cn } from '$lib/utils'

  let {
    agentId,
    live = false,
    class: className,
  }: { agentId: string; live?: boolean; class?: string } = $props()
</script>

<span
  class={cn(
    'inline-flex h-[21px] items-center gap-1 rounded-full bg-white/[.035] px-2 font-mono text-[9.5px] font-medium whitespace-nowrap text-agent',
    className,
  )}
>
  <span class="relative flex size-1.5 shrink-0" aria-hidden="true">
    {#if live}
      <span class="absolute inline-flex size-full animate-ping rounded-full bg-agent opacity-60 motion-reduce:animate-none"></span>
    {/if}
    <span class="relative inline-flex size-1.5 rounded-full bg-agent"></span>
  </span>
  {agentId}
</span>
