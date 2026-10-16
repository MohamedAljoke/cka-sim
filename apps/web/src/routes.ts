export type Route = { page: 'tasks' } | { page: 'task'; id: string }

export const homePath = '#/'

export function taskPath(id: string): string {
  return `#/tasks/${encodeURIComponent(id)}`
}

export function parseRoute(hash: string): Route {
  const match = /^#\/tasks\/([^/]+)$/.exec(hash)
  if (!match) return { page: 'tasks' }
  try {
    return { page: 'task', id: decodeURIComponent(match[1]) }
  } catch {
    return { page: 'tasks' }
  }
}
