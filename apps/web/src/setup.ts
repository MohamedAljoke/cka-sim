import { getCurrentTask, startTask, tidyTask, type Lab } from './api'
import { alreadyTidied } from './routes'
import { el, message } from './ui'

type SetupState = 'waiting' | 'running' | 'done' | 'failed'

const waiting = 'Start the lab, and the task is set up once it is ready.'
const recheckEvery = 2000

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
    if (setUp && signal.reason !== alreadyTidied) tidyTask(id)
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

  let ready = false
  let recheck: ReturnType<typeof setTimeout> | undefined
  signal.addEventListener('abort', () => clearTimeout(recheck))

  // A task the study session already set up isn't set up twice; one it is still setting up is waited for.
  const setUpOnce = async (first: boolean) => {
    const current = await getCurrentTask(signal).catch(() => null)
    if (signal.aborted || !ready) return
    const ours = current?.id === id
    if (first && (!ours || current.setup === 'failed')) return setUpTask()
    if (!ours) return show('failed', 'The task is no longer set up on your lab.', 'Set up task')
    setUp = true
    if (current.setup === 'ready') {
      show('done', 'Task already set up on your lab. Setting it up again starts it over.', 'Set up again')
      return
    }
    if (current.setup === 'failed') {
      show('failed', 'The setup failed.', 'Try again')
      return
    }
    show('running', 'The task is being set up on the cluster…')
    recheck = setTimeout(() => setUpOnce(false), recheckEvery)
  }

  // A lab that becomes ready while the page watches is a new cluster, so the task is set up again, whatever the
  // platform remembers from the cluster before.
  let sawDown = false
  const follow = (lab: Lab) => {
    const wasReady = ready
    ready = lab.state === 'ready'
    if (ready && !wasReady) {
      if (sawDown) setUpTask()
      else setUpOnce(true)
    }
    if (!ready) {
      sawDown = true
      clearTimeout(recheck)
      show('waiting', waiting)
    }
  }

  show('waiting', waiting)
  return { panel, follow }
}

function seconds(since: number): string {
  return `${Math.max(1, Math.round((Date.now() - since) / 1000))}s`
}
