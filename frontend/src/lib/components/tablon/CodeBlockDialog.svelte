<script lang="ts">
  import { tick } from 'svelte'
  import Button from './Button.svelte'
  import Modal from './Modal.svelte'
  import { CODE_LANGUAGES, normalizeLanguage } from '$lib/markdown-editor'
  import { inputClass, labelClass } from '$lib/ui'

  let {
    open = $bindable(false),
    oninsert,
  }: {
    open?: boolean
    oninsert: (language: string) => void
  } = $props()

  let language = $state('')
  let input = $state<HTMLInputElement>()

  $effect(() => {
    if (!open) return
    language = ''
    void tick().then(() => input?.focus())
  })

  function submit() {
    oninsert(normalizeLanguage(language))
    open = false
  }
</script>

<Modal bind:open title="Insert code block" description="Choose a language or leave it empty for plain text.">
  <form
    id="code-block-form"
    class="space-y-4"
    onsubmit={(event) => {
      event.preventDefault()
      submit()
    }}
  >
    <div>
      <label class={labelClass} for="code-block-language">Language</label>
      <input
        id="code-block-language"
        bind:this={input}
        bind:value={language}
        class={inputClass}
        list="code-block-languages"
        placeholder="typescript"
        autocomplete="off"
        maxlength="32"
      />
      <datalist id="code-block-languages">
        {#each CODE_LANGUAGES as item (item)}
          <option value={item}></option>
        {/each}
      </datalist>
      <p class="mt-1.5 text-[11px] leading-relaxed text-[var(--text-3)]">
        Leave empty for plain text. The fenced block keeps the Markdown source editable.
      </p>
    </div>
  </form>

  {#snippet footer()}
    <Button variant="secondary" onclick={() => (open = false)}>Cancel</Button>
    <Button type="submit" form="code-block-form">Insert code block</Button>
  {/snippet}
</Modal>
