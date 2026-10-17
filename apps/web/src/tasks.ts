import { listTasks, noFilter, type Filter, type Task } from './api'
import { picker, type Option } from './picker'
import { taskPath, tasksPath } from './routes'
import { chips, el, message } from './ui'

export async function showTasks(pane: HTMLElement, filter: Filter, signal: AbortSignal) {
  document.title = 'cka-sim'
  pane.replaceChildren(el('p', 'muted', 'Loading tasks…'))
  let all: Task[]
  try {
    all = await listTasks(noFilter, signal)
  } catch (err) {
    if (signal.aborted) return
    const retry = el('button', 'action', 'Retry')
    retry.type = 'button'
    retry.onclick = () => showTasks(pane, filter, signal)
    pane.replaceChildren(el('p', 'error', `Cannot load tasks: ${message(err)}`), retry)
    return
  }
  if (signal.aborted) return

  const count = el('p', 'muted')
  const list = el('ol', 'task-list')
  const draw = (found: Task[]) => {
    count.textContent = found.length === all.length ? `${all.length} tasks` : `${found.length} of ${all.length} tasks`
    if (all.length === 0) list.replaceChildren(el('li', 'muted', 'The catalog has no tasks yet.'))
    else if (found.length === 0) list.replaceChildren(el('li', 'muted', 'No task matches these filters.'))
    else list.replaceChildren(...found.map(taskItem))
  }

  let pending: AbortController | undefined
  const refresh = async () => {
    const chosen: Filter = { domains: domains.chosen(), topics: topics.chosen(), search: search.value }
    history.replaceState(null, '', tasksPath(chosen))
    clear.hidden = chosen.domains.length === 0 && chosen.topics.length === 0 && !chosen.search.trim()
    pending?.abort()
    pending = new AbortController()
    const mine = AbortSignal.any([signal, pending.signal])
    if (clear.hidden) return draw(all)
    try {
      const found = await listTasks(chosen, mine)
      if (!mine.aborted) draw(found)
    } catch (err) {
      if (mine.aborted) return
      count.textContent = ''
      list.replaceChildren(el('li', 'error', `Cannot filter the tasks: ${message(err)}`))
    }
  }

  const search = el('input', 'task-search')
  Object.assign(search, { type: 'search', placeholder: 'Search tasks…', value: filter.search })
  search.setAttribute('aria-label', 'Search tasks by id or title')
  let typing: number | undefined
  search.oninput = () => {
    clearTimeout(typing)
    typing = setTimeout(refresh, 200)
  }
  signal.addEventListener('abort', () => clearTimeout(typing))

  const domains = picker('domains', options(all, (t) => [t.domain], (_, t) => t.domainTitle), filter.domains, refresh, signal)
  const topics = picker('topics', options(all, (t) => t.topics, (topic) => topic), filter.topics, refresh, signal)
  const clear = el('button', 'link', 'Clear filters')
  clear.type = 'button'
  clear.onclick = () => {
    search.value = ''
    domains.clear()
    topics.clear()
    refresh()
  }
  const filters = el('div', 'filters')
  filters.append(search, domains.element, topics.element, clear)

  const card = el('section', 'card')
  card.append(el('h2', '', 'Tasks'), filters, count, list)
  pane.replaceChildren(card)
  refresh()
}

function options(all: Task[], valuesOf: (task: Task) => string[], labelOf: (value: string, task: Task) => string): Option[] {
  const found = new Map<string, Option>()
  for (const task of all) {
    for (const value of valuesOf(task)) {
      const option = found.get(value) ?? { value, label: labelOf(value, task), count: 0 }
      option.count++
      found.set(value, option)
    }
  }
  return [...found.values()].sort((a, b) => a.label.localeCompare(b.label))
}

function taskItem(task: Task): HTMLElement {
  const link = el('a', 'task')
  link.href = taskPath(task.id)
  link.append(el('span', 'task-title', task.title), chips(task))
  const item = el('li')
  item.append(link)
  return item
}
