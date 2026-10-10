import { endLab, getLab, startLab, type LabState } from './api'
import { el, message, setStatus } from './ui'

type Hooks = {
  ready: () => void
  ended: () => void
}

let poll: number | undefined

// showLab keeps the topbar in step with the lab. The panes only work once it's ready.
export async function showLab(bar: HTMLElement, hooks: Hooks) {
  let wasReady = false
  const render = (state: LabState) => {
    clearTimeout(poll)
    const { lab } = state
    const status = el('span', 'status')
    if (lab.state === 'starting') {
      const since = Date.parse(state.now) - Date.parse(lab.started!)
      setStatus(status, '', `Starting lab on ${lab.provider}… ${duration(since)}`)
      bar.replaceChildren(status)
      poll = window.setTimeout(refresh, 1000)
      return
    }
    if (lab.state === 'ready') {
      const took = Date.parse(lab.ready!) - Date.parse(lab.started!)
      setStatus(status, 'ready', `Lab ready in ${duration(took)} · ${lab.provider}`)
      bar.replaceChildren(status, button('End lab', end))
      if (!wasReady) hooks.ready()
      wasReady = true
      return
    }
    if (lab.state === 'failed') setStatus(status, 'error', `Lab failed: ${lab.error}`)
    else setStatus(status, '', `No lab · ${lab.provider}`)
    bar.replaceChildren(status, button(lab.state === 'failed' ? 'Try again' : 'Start lab', start, true))
  }
  const failed = (err: unknown) => {
    const status = el('span', 'status')
    setStatus(status, 'error', message(err))
    bar.replaceChildren(status, button('Retry', refresh))
  }
  async function refresh() {
    try {
      render(await getLab())
    } catch (err) {
      failed(err)
    }
  }
  async function start() {
    try {
      render(await startLab())
    } catch (err) {
      failed(err)
    }
  }
  async function end() {
    try {
      await endLab()
    } catch (err) {
      failed(err)
      return
    }
    wasReady = false
    hooks.ended()
    refresh()
  }
  await refresh()
}

function button(label: string, onclick: () => void, primary = false): HTMLButtonElement {
  const b = el('button', primary ? 'action primary' : 'action', label)
  b.type = 'button'
  b.onclick = () => {
    b.disabled = true
    onclick()
  }
  return b
}

function duration(ms: number): string {
  if (ms < 60_000) return `${(Math.max(ms, 0) / 1000).toFixed(1)}s`
  const s = Math.round(ms / 1000)
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
}
