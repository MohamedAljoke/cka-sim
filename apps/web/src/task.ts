import { getTask, type TaskDetail } from './api'
import { homePath } from './routes'
import { chips, el, markdown, message } from './ui'

export async function showTask(pane: HTMLElement, id: string, signal: AbortSignal) {
  pane.replaceChildren(el('p', 'muted', 'Loading the task…'))
  let task: TaskDetail
  try {
    task = await getTask(id, signal)
  } catch (err) {
    if (signal.aborted) return
    pane.replaceChildren(backLink(), el('p', 'error', `Cannot open this task: ${message(err)}`))
    return
  }
  if (signal.aborted) return

  document.title = `${task.title} · cka-sim`
  const question = el('div', 'markdown')
  markdown(question, task.question)
  const card = el('article', 'card task-page')
  card.append(el('h1', '', task.title), chips(task), hostLine(task), question)
  pane.replaceChildren(backLink(), card)
}

function backLink(): HTMLElement {
  const link = el('a', 'back', '← All tasks')
  link.href = homePath
  return link
}

function hostLine(task: TaskDetail): HTMLElement {
  const line = el('p', 'host')
  line.append(el('span', 'muted', 'Connect first: '), el('code', '', `ssh ${task.host}`))
  return line
}
