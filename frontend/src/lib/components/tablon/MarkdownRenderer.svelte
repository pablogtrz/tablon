<script lang="ts">
  import { renderMarkdown } from '$lib/markdown'
  import { cn } from '$lib/utils'

  let {
    source,
    interactive = false,
    disabled = false,
    ontasktoggle,
    class: className,
  }: {
    source: string
    interactive?: boolean
    disabled?: boolean
    ontasktoggle?: (index: number, checked: boolean) => void
    class?: string
  } = $props()

  let container = $state<HTMLDivElement>()

  $effect(() => {
    if (!container) return
    const focusedTask = focusedTaskIndex()
    container.replaceChildren(renderMarkdown(source))
    syncCheckboxes()
    restoreTaskFocus(focusedTask)
  })

  $effect(() => {
    if (container) syncCheckboxes()
  })

  function focusedTaskIndex(): number | null {
    if (!container) return null
    const active = document.activeElement
    if (!(active instanceof HTMLElement) || !container.contains(active)) return null
    const input = active.closest('input[type="checkbox"]') as HTMLInputElement | null
    const raw = input?.dataset.taskIndex
    return raw === undefined ? null : Number(raw)
  }

  function restoreTaskFocus(index: number | null) {
    if (!container || index === null || Number.isNaN(index)) return
    const next = container.querySelector<HTMLInputElement>(`input[data-task-index="${index}"]`)
    next?.focus()
  }

  function syncCheckboxes() {
    if (!container) return
    for (const input of container.querySelectorAll<HTMLInputElement>('input[type="checkbox"]')) {
      input.disabled = disabled || !interactive
    }
  }

  async function handleClick(event: MouseEvent) {
    const target = event.target
    if (!(target instanceof Element)) return

    const copyButton = target.closest('.markdown-codeblock-copy') as HTMLButtonElement | null
    if (copyButton) {
      event.preventDefault()
      await copyCode(copyButton)
      return
    }

    if (!interactive || disabled || !ontasktoggle) return
    const input = target.closest('input[type="checkbox"]') as HTMLInputElement | null
    if (!input) return
    const raw = input.dataset.taskIndex
    if (raw === undefined) return
    const index = Number(raw)
    if (Number.isNaN(index)) return
    event.preventDefault()
    ontasktoggle(index, input.checked)
  }

  async function writeClipboard(text: string) {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
      return
    }
    const area = document.createElement('textarea')
    area.value = text
    area.setAttribute('readonly', '')
    area.style.position = 'fixed'
    area.style.top = '-1000px'
    document.body.append(area)
    area.select()
    const copied = document.execCommand('copy')
    area.remove()
    if (!copied) throw new Error('clipboard unavailable')
  }

  async function copyCode(button: HTMLButtonElement) {
    const wrapper = button.closest('.markdown-codeblock')
    const code = wrapper?.querySelector('code')
    const status = wrapper?.querySelector('.markdown-codeblock-status')
    try {
      const text = (code?.textContent ?? '').replace(/\n$/, '')
      await writeClipboard(text)
      button.textContent = 'Copied'
      if (status) status.textContent = 'Code copied'
    } catch {
      button.textContent = 'Copy failed'
      if (status) status.textContent = 'Could not copy code'
    }
    button.focus({ preventScroll: true })
    setTimeout(() => {
      button.textContent = 'Copy'
      if (status) status.textContent = ''
    }, 1500)
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div bind:this={container} class={cn('markdown-body', className)} onclick={handleClick}></div>
