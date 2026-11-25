import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { ProjectListItem } from './types'

vi.mock('./api', () => ({ listProjects: vi.fn() }))

import { listProjects } from './api'
import {
  patchProjectNav,
  projectNav,
  refreshProjectNav,
  removeProjectNav,
} from './project-nav.svelte'

function makeProject(id: string, name: string): ProjectListItem {
  return {
    id,
    identifier: id.toUpperCase(),
    name,
    description: '',
    status: 'active',
    logo_gradient: 0,
    logo_emoji: '',
    logo_gradient_from: '',
    logo_gradient_to: '',
    logo_gradient_angle: 135,
    created_at: '2026-11-24T10:00:00Z',
    updated_at: '2026-11-24T10:00:00Z',
    epic_count: 0,
    issue_count: 0,
    done_count: 0,
  }
}

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((done) => {
    resolve = done
  })
  return { promise, resolve }
}

const project = makeProject('project-1', 'Project')
const other = makeProject('project-2', 'Other')

beforeEach(() => {
  vi.mocked(listProjects).mockReset()
  projectNav.items = []
  projectNav.loading = false
  projectNav.loaded = false
  projectNav.error = ''
})

describe('project nav store', () => {
  it('shares one in-flight refresh', async () => {
    const pending = deferred<ProjectListItem[]>()
    vi.mocked(listProjects).mockReturnValueOnce(pending.promise)

    const first = refreshProjectNav()
    const second = refreshProjectNav()
    expect(listProjects).toHaveBeenCalledTimes(1)

    pending.resolve([project])
    await Promise.all([first, second])

    expect(projectNav.loaded).toBe(true)
    expect(projectNav.loading).toBe(false)
    expect(projectNav.items).toEqual([project])
  })

  it('patches and removes local projects', () => {
    projectNav.items = [project, other]
    projectNav.loaded = true

    patchProjectNav({ ...project, name: 'Renamed' })
    expect(projectNav.items[0].name).toBe('Renamed')

    removeProjectNav(other.id)
    expect(projectNav.items).toEqual([{ ...project, name: 'Renamed' }])
  })

  it('ignores a stale refresh after a local patch', async () => {
    projectNav.items = [project]
    projectNav.loaded = true
    const pending = deferred<ProjectListItem[]>()
    vi.mocked(listProjects).mockReturnValueOnce(pending.promise)

    const refresh = refreshProjectNav()
    patchProjectNav({ ...project, name: 'Renamed' })
    pending.resolve([project])
    await refresh

    expect(projectNav.items[0].name).toBe('Renamed')
  })

  it('refreshes when patching an unknown project', async () => {
    projectNav.loaded = true
    vi.mocked(listProjects).mockResolvedValueOnce([other])

    patchProjectNav(other)
    await vi.waitFor(() => expect(projectNav.items).toEqual([other]))
    expect(listProjects).toHaveBeenCalledTimes(1)
  })
})
