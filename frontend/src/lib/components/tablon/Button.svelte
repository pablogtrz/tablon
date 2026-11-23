<script lang="ts">
  import type { Snippet } from 'svelte'
  import { cn } from '$lib/utils'

  let {
    variant = 'primary',
    type = 'button',
    form,
    class: className,
    disabled = false,
    title,
    ariaLabel,
    onclick,
    children,
  }: {
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
    type?: 'button' | 'submit'
    form?: string
    class?: string
    disabled?: boolean
    title?: string
    ariaLabel?: string
    onclick?: (event: MouseEvent) => void
    children: Snippet
  } = $props()

  const base =
    'inline-flex h-10 cursor-pointer items-center justify-center gap-2 rounded-[10px] px-3.5 text-sm font-medium whitespace-nowrap transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 sm:h-9'

  const variants = {
    primary: 'bg-primary text-primary-foreground hover:bg-brand-hover active:bg-brand-active',
    secondary: 'border border-border bg-[var(--surface-1)] text-foreground hover:bg-[var(--surface-2)]',
    ghost: 'text-[var(--text-2)] hover:bg-[var(--surface-2)] hover:text-foreground',
    danger: 'border border-destructive/30 bg-destructive/10 text-destructive hover:bg-destructive/15',
  }
</script>

<button {type} {form} {disabled} {title} aria-label={ariaLabel} class={cn(base, variants[variant], className)} {onclick}>
  {@render children()}
</button>
