import { marked } from 'marked'
import { listTasks, startTask, taskQuestion, type Task } from './api'
import { icon } from './icons'

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

function openTask(pane: HTMLElement, task: Task, number: number) {
  const back = el('button', 'back', '← Back to tasks')
  back.onclick = () => showTasks(pane)
  const status = el('p', 'status', 'Preparing the cluster…')
  const body = el('div', 'question-text')
  const host = el('div', 'host')
  host.append(el('span', 'host-label', 'Connect first'), copyable(`ssh ${task.host}`))
  pane.replaceChildren(back, el('h2', '', `${number}. ${task.title}`), chips(task), host, status, body)

  startTask(task.id).then(
    () => setStatus(status, 'ready', 'Ready: the cluster is set up.'),
    (err) => setStatus(status, 'error', `Setup failed: ${message(err)}`),
  )
  taskQuestion(task.id).then(
    // The question comes from the bundled catalog, not from users, so it's rendered as is.
    (question) => (body.innerHTML = marked.parse(question, { async: false })),
    (err) => setStatus(status, 'error', `Cannot load the question: ${message(err)}`),
  )
}

function setStatus(status: HTMLElement, kind: 'ready' | 'error', text: string) {
  status.className = `status ${kind}`
  status.textContent = text
}

function copyable(command: string): HTMLElement {
  const button = el('button', 'copy-command')
  button.type = 'button'
  button.title = 'Copy to clipboard'
  const label = el('span', 'copy-label', 'Copy')
  const showCopied = (copied: boolean) => {
    button.classList.toggle('copied', copied)
    label.replaceChildren(icon(copied ? 'check' : 'copy', 14), copied ? 'Copied' : 'Copy')
  }
  showCopied(false)
  button.append(el('code', '', command), label)

  let reset: number | undefined
  button.onclick = async () => {
    try {
      await navigator.clipboard.writeText(command)
    } catch {
      // No clipboard access (e.g. plain http on a remote host): select it for a manual copy.
      getSelection()?.selectAllChildren(button.querySelector('code')!)
      return
    }
    showCopied(true)
    clearTimeout(reset)
    reset = window.setTimeout(() => showCopied(false), 1500)
  }
  return button
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
