<script lang="ts">
  import { onMount } from 'svelte'
  import Button from '$lib/components/tablon/Button.svelte'
  import { ApiError, getSystemStatus, getToken, setToken } from '$lib/api'
  import { messageOf } from '$lib/errors'
  import { inputClass, labelClass } from '$lib/ui'
  import type { SystemStatus } from '$lib/types'

  let status = $state<SystemStatus | null>(null)
  let loading = $state(true)
  let error = $state('')
  let needsToken = $state(false)
  let token = $state('')
  let hasToken = $state(false)

  async function loadStatus() {
    loading = true
    error = ''
    needsToken = false
    try {
      status = await getSystemStatus()
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        needsToken = true
      } else {
        error = messageOf(err)
      }
    } finally {
      hasToken = Boolean(getToken())
      loading = false
    }
  }

  function saveToken() {
    setToken(token)
    token = ''
    void loadStatus()
  }

  function clearToken() {
    setToken('')
    token = ''
    void loadStatus()
  }

  onMount(() => {
    void loadStatus()
  })
</script>

<svelte:head>
  <title>Settings · Tablon</title>
</svelte:head>

<div>
  <p class="text-xs font-medium text-[var(--text-3)]">Workspace</p>
  <h1 class="mt-1 text-xl font-semibold tracking-[-.02em] outline-none" tabindex="-1" data-route-focus>Settings</h1>
  <p class="mt-1 text-sm text-[var(--text-2)]">Local configuration and API access.</p>
</div>

<div class="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
  <section class="rounded-2xl border border-border bg-[var(--surface-1)] p-4">
    <h2 class="text-sm font-semibold">System</h2>
    {#if loading}
      <p class="mt-3 text-xs text-[var(--text-3)]">Checking…</p>
    {:else if needsToken}
      <p class="mt-3 text-xs leading-relaxed text-[var(--text-2)]">
        This server requires an API token. Enter it below to load system status and use the API from this browser.
      </p>
    {:else if status}
      <dl class="mt-3 space-y-2 text-xs">
        <div class="flex items-center justify-between gap-3">
          <dt class="text-[var(--text-3)]">Version</dt>
          <dd class="font-mono text-[var(--text-2)]">{status.version}</dd>
        </div>
        <div class="flex items-center justify-between gap-3">
          <dt class="text-[var(--text-3)]">Database</dt>
          <dd class="text-[var(--text-2)]">{status.database}</dd>
        </div>
        <div class="flex items-center justify-between gap-3">
          <dt class="text-[var(--text-3)]">Auth</dt>
          <dd class="text-[var(--text-2)]">{status.auth_enabled ? 'Enabled' : 'Optional, currently off'}</dd>
        </div>
      </dl>
    {:else if error}
      <p class="mt-3 text-xs text-destructive" role="alert">{error}</p>
    {/if}
  </section>

  <section class="rounded-2xl border border-border bg-[var(--surface-1)] p-4">
    <h2 class="text-sm font-semibold">API token</h2>
    <p class="mt-1 text-xs leading-relaxed text-[var(--text-2)]">
      Stored in this browser session only. Set <code class="font-mono text-[11px]">TABLON_API_TOKEN</code> on the server to require it.
    </p>
    <form class="mt-3 space-y-3" onsubmit={(event) => { event.preventDefault(); saveToken() }}>
      <div>
        <label class={labelClass} for="api-token">Bearer token</label>
        <input id="api-token" class={inputClass} type="password" bind:value={token} autocomplete="off" placeholder={hasToken ? 'Token stored' : 'Paste token'} />
      </div>
      <div class="flex flex-wrap gap-2">
        <Button type="submit">Save token</Button>
        {#if hasToken}
          <Button variant="secondary" onclick={clearToken}>Clear token</Button>
        {/if}
      </div>
    </form>
    {#if error}
      <p class="mt-3 text-xs text-destructive" role="alert">{error}</p>
    {/if}
  </section>
</div>
