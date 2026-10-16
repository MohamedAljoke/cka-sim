import { listTasks, type Task } from './api'
import { el, message } from './ui'

export async function showTasks(pane: HTMLElement) {
  pane.replaceChildren(el('p', 'muted', 'Loading tasks…'))
  let tasks: Task[]
  try {
    tasks = await listTasks()
  } catch (err) {
    const retry = el('button', 'action', 'Retry')
    retry.type = 'button'
    retry.onclick = () => showTasks(pane)
    pane.replaceChildren(el('p', 'error', `Cannot load tasks: ${message(err)}`), retry)
    return
  }

  const list = el('ol', 'task-list')
  if (tasks.length === 0) list.append(el('li', 'muted', 'The catalog has no tasks yet.'))
  else list.append(...tasks.map(taskItem))
  const card = el('section', 'card')
  card.append(el('h2', '', 'Tasks'), el('p', 'muted', `${tasks.length} tasks`), list)
  pane.replaceChildren(card)
}

function taskItem(task: Task): HTMLElement {
  const item = el('li', 'task')
  item.append(el('span', 'task-title', task.title), chips(task))
  return item
}

function chips(task: Task): HTMLElement {
  const row = el('span', 'chips')
  row.append(el('span', 'chip domain', task.domainTitle), ...task.topics.map((topic) => el('span', 'chip', topic)))
  return row
}
