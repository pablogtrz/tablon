<script lang="ts">
  import { cn } from '$lib/utils'

  let { identifier, class: className }: { identifier: string; class?: string } = $props()

  const parts = $derived.by(() => {
    const index = identifier.lastIndexOf('-')
    if (index <= 0 || index === identifier.length - 1) {
      return { prefix: '', number: identifier }
    }
    return {
      prefix: identifier.slice(0, index + 1),
      number: identifier.slice(index + 1),
    }
  })
</script>

<span class={cn('font-mono text-[11.5px] whitespace-nowrap', className)}>
  {#if parts.prefix}<span class="text-[var(--text-3)]">{parts.prefix}</span>{/if}<span class="font-medium text-foreground">{parts.number}</span>
</span>
