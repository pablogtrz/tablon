import { listProjects } from './api'
import type { Project, ProjectListItem } from './types'

export const projectNav = $state({
  items: [] as ProjectListItem[],
  loading: false,
  loaded: false,
  error: '',
})

let request = 0

export async function refreshProjectNav(includeArchived = true) {
  const current = ++request
  projectNav.loading = !projectNav.loaded
  projectNav.error = ''
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
}

export function patchProjectNav(project: Project) {
  const index = projectNav.items.findIndex((item) => item.id === project.id)
  if (index === -1) {
    void refreshProjectNav()
    return
  }
  projectNav.items[index] = { ...projectNav.items[index], ...project }
}

export function removeProjectNav(id: string) {
  projectNav.items = projectNav.items.filter((item) => item.id !== id)
}
