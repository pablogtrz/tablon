export type Route =
  | { name: 'dashboard' }
  | { name: 'project'; projectRef: string }
  | { name: 'issue'; projectRef: string; issueRef: string }
  | { name: 'settings' }
  | { name: 'not-found' }

type NavigationGuard = (to: string) => boolean | Promise<boolean>

export const route = $state({
  path: typeof window === 'undefined' ? '/' : window.location.pathname,
  search: typeof window === 'undefined' ? '' : window.location.search,
})

let guard: NavigationGuard | null = null
let navigating = false

export function setNavigationGuard(next: NavigationGuard) {
  guard = next
  return () => {
    if (guard === next) guard = null
  }
}

function sync() {
  route.path = window.location.pathname
  route.search = window.location.search
}

async function allowNavigation(to: string): Promise<boolean> {
  if (!guard) return true
  return guard(to)
}

if (typeof window !== 'undefined') {
  window.addEventListener('popstate', () => {
    void (async () => {
      const to = window.location.pathname + window.location.search
      if (!(await allowNavigation(to))) {
        window.history.pushState({}, '', route.path + route.search)
        return
      }
      sync()
    })()
  })
}

export async function navigate(to: string, replace = false): Promise<boolean> {
  if (navigating) return false
  navigating = true
  try {
    if (!(await allowNavigation(to))) return false
    if (replace) {
      window.history.replaceState({}, '', to)
    } else {
      window.history.pushState({}, '', to)
    }
    sync()
    return true
  } finally {
    navigating = false
  }
}

export function parseRoute(path: string): Route {
  if (path === '/' || path === '') return { name: 'dashboard' }
  if (path === '/settings') return { name: 'settings' }

  const issueMatch = path.match(/^\/projects\/([^/]+)\/issues\/([^/]+)$/)
  if (issueMatch) {
    return {
      name: 'issue',
      projectRef: decodeURIComponent(issueMatch[1]),
      issueRef: decodeURIComponent(issueMatch[2]),
    }
  }

  const projectMatch = path.match(/^\/projects\/([^/]+)$/)
  if (projectMatch) {
    return { name: 'project', projectRef: decodeURIComponent(projectMatch[1]) }
  }

  return { name: 'not-found' }
}

export function projectPath(ref: string): string {
  return `/projects/${encodeURIComponent(ref)}`
}

export function issuePath(projectRef: string, issueRef: string): string {
  return `/projects/${encodeURIComponent(projectRef)}/issues/${encodeURIComponent(issueRef)}`
}
