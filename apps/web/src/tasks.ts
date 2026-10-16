import { listTasks, type Task } from './api'
import { taskPath } from './routes'
import { chips, el, message } from './ui'

export async function showTasks(pane: HTMLElement, signal: AbortSignal) {
  document.title = 'cka-sim'
  pane.replaceChildren(el('p', 'muted', 'Loading tasks…'))
  let tasks: Task[]
  try {
    tasks = await listTasks(signal)
  } catch (err) {
    if (signal.aborted) return
    const retry = el('button', 'action', 'Retry')
    retry.type = 'button'
    retry.onclick = () => showTasks(pane, signal)
    pane.replaceChildren(el('p', 'error', `Cannot load tasks: ${message(err)}`), retry)
    return
  }
  if (signal.aborted) return

  const list = el('ol', 'task-list')
  if (tasks.length === 0) list.append(el('li', 'muted', 'The catalog has no tasks yet.'))
  else list.append(...tasks.map(taskItem))
  const card = el('section', 'card')
  card.append(el('h2', '', 'Tasks'), el('p', 'muted', `${tasks.length} tasks`), list)
  pane.replaceChildren(card)
}

function taskItem(task: Task): HTMLElement {
  const link = el('a', 'task')
  link.href = taskPath(task.id)
  link.append(el('span', 'task-title', task.title), chips(task))
  const item = el('li')
  item.append(link)
  return item
}
