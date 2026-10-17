import { getSolution, getTask, type TaskDetail } from './api'
import type { LabStatus } from './lab'
import { homePath } from './routes'
import { setupPanel } from './setup'
import { shellPanel } from './shell'
import { chips, copyable, el, markdown, message } from './ui'

export async function showTask(pane: HTMLElement, id: string, lab: LabStatus, signal: AbortSignal) {
  pane.replaceChildren(note(el('p', 'muted', 'Loading the task…')))
  let task: TaskDetail
  try {
    task = await getTask(id, signal)
  } catch (err) {
    if (signal.aborted) return
    pane.replaceChildren(note(backLink(), el('p', 'error', `Cannot open this task: ${message(err)}`)))
    return
  }
  if (signal.aborted) return

  document.title = `${task.title} · cka-sim`
  const question = el('div', 'markdown')
  markdown(question, task.question)
  const brief = el('article', 'task-brief')
  brief.append(backLink(), el('h1', '', task.title), chips(task), hostLine(task), question, solutionPanel(task.id, signal))

  const setup = setupPanel(task.id, signal)
  const shell = shellPanel(signal, lab.refresh)
  lab.watch(signal, (state) => {
    setup.follow(state)
    shell.follow(state)
  })
  const workspace = el('section', 'task-lab')
  workspace.setAttribute('aria-label', 'Workspace')
  workspace.append(setup.panel, shell.panel)

  const layout = el('div', 'task-layout')
  layout.append(brief, workspace)
  pane.replaceChildren(layout)
}

function note(...children: HTMLElement[]): HTMLElement {
  const box = el('div', 'task-note')
  box.append(...children)
  return box
}

function backLink(): HTMLElement {
  const link = el('a', 'back', '← All tasks')
  link.href = homePath
  return link
}

function hostLine(task: TaskDetail): HTMLElement {
  const line = el('p', 'host')
  line.append(el('span', 'muted', 'Connect first: '), copyable(`ssh ${task.host}`))
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
