import { checkTask, type Grade, type Lab } from './api'
import { el, message } from './ui'

type CheckState = 'waiting' | 'idle' | 'running' | 'done' | 'failed'

const waiting = 'Start the lab, then check your work here.'

export function checkPanel(id: string, signal: AbortSignal): { panel: HTMLElement; follow: (lab: Lab) => void } {
  const dot = el('span', 'lab-dot')
  const text = el('span')
  const line = el('p', 'lab-line')
  line.append(dot, text)
  const action = el('button', 'action')
  action.type = 'button'
  const list = el('ul', 'check-list')
  const output = el('pre', 'setup-output')
  const panel = el('section', 'lab-panel check-panel')
  panel.setAttribute('aria-live', 'polite')
  panel.append(line, action, list, output)

  const show = (state: CheckState, says: string, offer = '', grade?: Grade) => {
    panel.dataset.state = state
    text.className = state === 'failed' && !grade?.checks.length ? 'error' : ''
    text.textContent = says
    action.hidden = !offer
    action.textContent = offer
    list.replaceChildren(...(grade?.checks ?? []).map(checkItem))
    list.hidden = !grade?.checks.length
    output.textContent = grade?.output ?? ''
    output.hidden = !output.textContent
  }

  action.onclick = async () => {
    show('running', 'Checking your work on the cluster…')
    try {
      const grade = await checkTask(id, signal)
      if (signal.aborted) return
      if (!grade.checks.length) show('failed', `The check broke with exit code ${grade.exitCode ?? 0}.`, 'Check again', grade)
      else if (grade.earned === grade.total) show('done', `Solved · ${score(grade)}`, 'Check again', grade)
      else show('failed', `Not yet · ${score(grade)}`, 'Check again', grade)
    } catch (err) {
      if (!signal.aborted) show('failed', `Cannot check the task: ${message(err)}`, 'Try again')
    }
  }

  let ready = false
  const follow = (lab: Lab) => {
    if (lab.state === 'ready' && !ready) show('idle', 'Solve the task in the terminal, then check it.', 'Check')
    if (lab.state !== 'ready') show('waiting', waiting)
    ready = lab.state === 'ready'
  }

  show('waiting', waiting)
  return { panel, follow }
}

function checkItem(check: Grade['checks'][number]): HTMLElement {
  const item = el('li', check.passed ? 'passed' : 'failed')
  item.append(
    el('span', 'check-mark', check.passed ? '✓' : '✗'),
    el('span', '', check.description),
    el('span', 'check-points muted', `${check.points} pt`),
  )
  item.setAttribute('aria-label', `${check.passed ? 'Passed' : 'Failed'}: ${check.description}, ${check.points} points`)
  return item
}

function score(grade: Grade): string {
  return `${grade.earned}/${grade.total} points`
}
