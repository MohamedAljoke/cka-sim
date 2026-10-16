import { getSolution, getTask, type TaskDetail } from './api'
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
  card.append(el('h1', '', task.title), chips(task), hostLine(task), question, solutionPanel(task.id, signal))
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

function solutionPanel(id: string, signal: AbortSignal): HTMLElement {
  const toggle = el('button', 'action')
  toggle.type = 'button'
  const body = el('div', 'markdown solution')
  let loaded = false
  const show = (open: boolean) => {
    body.hidden = !open
    toggle.textContent = open ? 'Hide solution' : 'Show solution'
    toggle.setAttribute('aria-expanded', String(open))
  }
  show(false)

  toggle.onclick = async () => {
    if (!body.hidden) return show(false)
    if (!loaded) {
      toggle.disabled = true
      toggle.textContent = 'Loading the solution…'
      try {
        const text = await getSolution(id, signal)
        body.replaceChildren()
        markdown(body, text)
        loaded = true
      } catch (err) {
        if (signal.aborted) return
        body.replaceChildren(el('p', 'error', `Cannot load the solution: ${message(err)}`))
      } finally {
        toggle.disabled = false
      }
    }
    show(true)
  }

  const panel = el('section', 'solution-panel')
  panel.append(toggle, body)
  return panel
}
