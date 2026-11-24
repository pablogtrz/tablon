import { afterEach, describe, expect, it } from 'vitest'
import { mount, unmount } from 'svelte'
import IssueId from './IssueId.svelte'

let app: ReturnType<typeof mount> | null = null

afterEach(() => {
  if (app) unmount(app)
  app = null
  document.body.innerHTML = ''
})

describe('IssueId', () => {
  it('splits the identifier into a muted prefix and highlighted number', () => {
    app = mount(IssueId, { target: document.body, props: { identifier: 'TABLON-17' } })
    const text = document.body.textContent ?? ''
    expect(text).toBe('TABLON-17')
    expect(document.querySelectorAll('span span')).toHaveLength(2)
  })

  it('keeps identifiers without a suffix intact', () => {
    app = mount(IssueId, { target: document.body, props: { identifier: 'TABLON' } })
    expect(document.body.textContent).toBe('TABLON')
  })
})
