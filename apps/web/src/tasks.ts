import { marked } from 'marked'
import { listTasks, startTask, type Task } from './api'

export async function showTasks(pane: HTMLElement) {
  pane.replaceChildren(el('p', 'muted', 'Loading tasks…'))
  let tasks: Task[]
  try {
    tasks = await listTasks()
  } catch (err) {
    pane.replaceChildren(el('p', 'error', `Cannot load tasks: ${message(err)}`))
    return
  }

  const list = el('ol', 'task-list')
  tasks.forEach((task, i) => {
    const button = el('button', 'task')
    button.append(el('span', 'task-title', `${i + 1}. ${task.title}`), chips(task))
    button.onclick = () => openTask(pane, task, i + 1)
    const item = el('li')
    item.append(button)
    list.append(item)
  })
  pane.replaceChildren(el('h2', '', 'Tasks'), list)
}

async function openTask(pane: HTMLElement, task: Task, number: number) {
  const back = el('button', 'back', '← Back to tasks')
  back.onclick = () => showTasks(pane)
  const heading = el('h2', '', `${number}. ${task.title}`)
  const body = el('div', 'muted', 'Setting up…')
  pane.replaceChildren(back, heading, chips(task), body)

  try {
    const question = await startTask(task.id)
    body.className = 'question-text'
    // The question comes from the bundled catalog, not from users, so it's rendered as is.
    body.innerHTML = marked.parse(question, { async: false })
  } catch (err) {
    body.className = 'error'
    body.textContent = `Setup failed: ${message(err)}`
  }
}

function chips(task: Task): HTMLElement {
  const row = el('span', 'chips')
  row.append(el('span', 'chip domain', task.domain), ...task.topics.map((t) => el('span', 'chip', t)))
  return row
}

function el<K extends keyof HTMLElementTagNameMap>(tag: K, className = '', text = '') {
  const node = document.createElement(tag)
  node.className = className
  node.textContent = text
  return node
}

function message(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}
