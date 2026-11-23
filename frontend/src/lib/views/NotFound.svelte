<script lang="ts">
  import Button from '$lib/components/tablon/Button.svelte'
  import { navigate } from '$lib/router.svelte'
</script>

<svelte:head>
  <title>Not found · Tablon</title>
</svelte:head>

<div class="glass mx-auto mt-12 max-w-lg rounded-2xl p-6 text-center">
  <p class="font-mono text-xs text-[var(--text-3)]">404</p>
  <h1 class="mt-2 text-lg font-semibold outline-none" tabindex="-1" data-route-focus>This page does not exist</h1>
  <p class="mt-2 text-sm leading-relaxed text-[var(--text-2)]">
    The link may be old, or the project was deleted.
  </p>
  <Button class="mt-5" onclick={() => navigate('/')}>Back to projects</Button>
</div>
