import { checkTask, type Grade, type Lab } from './api'
import { el, message } from './ui'

type CheckState = 'running' | 'done' | 'failed'

// The button sits with the task's actions beside the terminal; the result sits under the question it grades.
export function checkTool(id: string, signal: AbortSignal): { button: HTMLButtonElement; result: HTMLElement; follow: (lab: Lab) => void } {
  const button = el('button', 'action primary', 'Check')
  button.type = 'button'
  button.hidden = true

  const dot = el('span', 'lab-dot')
  const text = el('span')
  const line = el('p', 'lab-line')
  line.append(dot, text)
  const list = el('ul', 'check-list')
  const output = el('pre', 'setup-output')
  const result = el('section', 'check-result')
  result.setAttribute('aria-live', 'polite')
  result.append(el('h2', '', 'Your result'), line, list, output)
  result.hidden = true

  const show = (state: CheckState, says: string, grade?: Grade) => {
    result.hidden = false
    result.dataset.state = state
    text.className = state === 'failed' && !grade?.checks.length ? 'error' : ''
    text.textContent = says
    list.replaceChildren(...(grade?.checks ?? []).map(checkItem))
    list.hidden = !grade?.checks.length
    output.textContent = grade?.output ?? ''
    output.hidden = !output.textContent
  }

  button.onclick = async () => {
    button.disabled = true
    button.textContent = 'Checking…'
    show('running', 'Checking your work on the cluster…')
    try {
      const grade = await checkTask(id, signal)
      if (signal.aborted) return
      if (!grade.checks.length) show('failed', `The check broke with exit code ${grade.exitCode ?? 0}.`, grade)
      else if (grade.earned === grade.total) show('done', `Solved · ${score(grade)}`, grade)
      else show('failed', `Not yet · ${score(grade)}`, grade)
    } catch (err) {
      if (!signal.aborted) show('failed', `Cannot check the task: ${message(err)}`)
    } finally {
      button.disabled = false
      button.textContent = 'Check'
    }
  }

  // A result belongs to the cluster it was read from, so a lab that goes away takes it along.
  const follow = (lab: Lab) => {
    button.hidden = lab.state !== 'ready'
    if (lab.state !== 'ready') result.hidden = true
  }

  return { button, result, follow }
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
