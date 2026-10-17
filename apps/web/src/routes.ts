import { filterQuery, noFilter, type Filter } from './api'

export type Route = { page: 'tasks'; filter: Filter } | { page: 'task'; id: string }

export const homePath = '#/'

export function tasksPath(filter: Filter): string {
  const query = filterQuery(filter)
  return query ? `${homePath}?${query}` : homePath
}

export function taskPath(id: string): string {
  return `#/tasks/${encodeURIComponent(id)}`
}

export function parseRoute(hash: string): Route {
  const task = /^#\/tasks\/([^/]+)$/.exec(hash)
  if (task) {
    try {
      return { page: 'task', id: decodeURIComponent(task[1]) }
    } catch {
      return { page: 'tasks', filter: noFilter }
    }
  }
  const params = new URLSearchParams(/^#\/\?(.*)$/.exec(hash)?.[1] ?? '')
  return { page: 'tasks', filter: { domains: params.getAll('domain'), topics: params.getAll('topic'), search: params.get('q') ?? '' } }
}
