import { listProjects } from './api'
import type { Project, ProjectListItem } from './types'

export const projectNav = $state({
  items: [] as ProjectListItem[],
  loading: false,
  loaded: false,
  error: '',
})

let request = 0
let inFlight: Promise<void> | null = null
let inFlightToken: object | null = null

export function refreshProjectNav(includeArchived = true): Promise<void> {
  if (inFlight) return inFlight

  const current = ++request
  const token = {}
  projectNav.loading = !projectNav.loaded
  projectNav.error = ''
  inFlightToken = token

  const pending = (async () => {
    try {
      const items = await listProjects(includeArchived)
      if (current !== request) return
      projectNav.items = items
      projectNav.loaded = true
    } catch {
      if (current !== request) return
      projectNav.error = 'Could not load projects.'
    } finally {
      if (current === request) projectNav.loading = false
    }
  })()

  inFlight = pending
  void pending.finally(() => {
    if (inFlightToken !== token) return
    inFlight = null
    inFlightToken = null
  })
  return pending
}

function invalidateInFlight() {
  request += 1
  inFlight = null
  inFlightToken = null
  projectNav.loading = false
}

export function patchProjectNav(project: Project) {
  invalidateInFlight()
  const index = projectNav.items.findIndex((item) => item.id === project.id)
  if (index === -1) {
    void refreshProjectNav()
    return
  }
  projectNav.items[index] = { ...projectNav.items[index], ...project }
}

export function removeProjectNav(id: string) {
  invalidateInFlight()
  projectNav.items = projectNav.items.filter((item) => item.id !== id)
}
