import { getSolution, getTask, type TaskDetail } from './api'
import { checkTool } from './check'
import type { LabStatus } from './lab'
import { homePath } from './routes'
import { setupPanel } from './setup'
import { shellPanel } from './shell'
import { tutorPanel } from './tutor'
import { chips, copyable, el, markdown, message, resizer } from './ui'

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
  const setup = setupPanel(task.id, signal)
  const check = checkTool(task.id, signal)
  const shell = shellPanel(signal, lab.refresh)
  lab.watch(signal, (state) => {
    setup.follow(state)
    check.follow(state)
    shell.follow(state)
  })

  const brief = el('article', 'task-brief')
  brief.append(
    backLink(),
    el('h1', '', task.title),
    chips(task),
    hostLine(task),
    question,
    check.result,
    solutionPanel(task.id, signal),
  )

  const bar = el('div', 'task-bar')
  bar.append(setup.panel, check.button)
  const workspace = el('section', 'task-lab')
  workspace.setAttribute('aria-label', 'Workspace')
  workspace.append(bar, shell.panel, tutorPanel(task.id, signal))

  const layout = el('div', 'task-layout')
  layout.append(
    brief,
    resizer({
      label: 'Resize the task',
      axis: 'x',
      grow: 1,
      key: 'cka-sim.task-width',
      size: () => brief.offsetWidth,
      limits: () => [280, layout.clientWidth - 420],
      apply: (px) => layout.style.setProperty('--task-width', `min(${px}px, calc(100% - 429px))`),
    }),
    workspace,
  )
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
