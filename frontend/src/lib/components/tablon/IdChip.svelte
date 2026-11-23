<script lang="ts">
  import { cn } from '$lib/utils'

  let { identifier, class: className }: { identifier: string; class?: string } = $props()
</script>

<span
  class={cn(
    'inline-flex items-center gap-1 rounded-[5px] bg-white/[.035] px-1.5 py-0.5 font-mono text-[9.5px] tracking-[.04em] whitespace-nowrap text-muted-foreground',
    className,
  )}
>
  {identifier}
</span>
