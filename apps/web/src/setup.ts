import { startTask, tidyTask, type Lab } from './api'
import { el, message } from './ui'

type SetupState = 'waiting' | 'running' | 'done' | 'failed'

const waiting = 'Start the lab, and the task is set up once it is ready.'

export function setupPanel(id: string, signal: AbortSignal): { panel: HTMLElement; follow: (lab: Lab) => void } {
  const dot = el('span', 'lab-dot')
  const text = el('span')
  const line = el('p', 'lab-line')
  line.append(dot, text)
  const action = el('button', 'action')
  action.type = 'button'
  const output = el('pre', 'setup-output')
  const panel = el('section', 'lab-panel setup-panel')
  panel.setAttribute('aria-live', 'polite')
  panel.append(line, action, output)

  const show = (state: SetupState, says: string, offer = '', log = '') => {
    panel.dataset.state = state
    text.className = state === 'failed' ? 'error' : ''
    text.textContent = says
    action.hidden = !offer
    action.textContent = offer
    output.hidden = !log
    output.textContent = log
    output.scrollTop = output.scrollHeight
  }

  let setUp = false
  signal.addEventListener('abort', () => {
    if (setUp) tidyTask(id)
  })

  const setUpTask = async () => {
    setUp = true
    const started = Date.now()
    show('running', 'Setting the task up on the cluster…')
    try {
      const setup = await startTask(id, signal)
      if (signal.aborted) return
      if (setup.passed) show('done', `Task set up in ${seconds(started)}. Setting it up again starts it over.`, 'Set up again')
      else show('failed', `The setup failed with exit code ${setup.exitCode}.`, 'Try again', setup.output)
    } catch (err) {
      if (!signal.aborted) show('failed', `Cannot set the task up: ${message(err)}`, 'Try again')
    }
  }
  action.onclick = setUpTask

  // A lab that comes back after a stop is a new cluster, so the task is set up again.
  let ready = false
  const follow = (lab: Lab) => {
    if (lab.state === 'ready' && !ready) setUpTask()
    if (lab.state !== 'ready') show('waiting', waiting)
    ready = lab.state === 'ready'
  }

  show('waiting', waiting)
  return { panel, follow }
}

function seconds(since: number): string {
  return `${Math.max(1, Math.round((Date.now() - since) / 1000))}s`
}
