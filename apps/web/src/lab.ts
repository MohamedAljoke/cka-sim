import { endLab, getLab, type LabState } from './api'
import { el, message, setStatus } from './ui'

type Hooks = {
  ready: () => void
  ended: () => void
}

// showLab keeps the topbar in step with the lab. There's no Start button: practising a task or
// starting an exam starts the lab on the server, and refresh picks that up.
export function showLab(bar: HTMLElement, hooks: Hooks) {
  let poll: number | undefined
  let wasReady = false
  const render = (state: LabState) => {
    clearTimeout(poll)
    const { lab } = state
    const status = el('span', 'status')
    if (lab.state === 'ready') {
      const took = Date.parse(lab.ready!) - Date.parse(lab.started!)
      setStatus(status, 'ready', `Lab ready in ${duration(took)} · ${lab.provider}`)
      bar.replaceChildren(status, button('End lab', end))
      if (!wasReady) hooks.ready()
      wasReady = true
      return
    }
    if (wasReady) hooks.ended()
    wasReady = false
    if (lab.state === 'starting') {
      const since = Date.parse(state.now) - Date.parse(lab.started!)
      setStatus(status, '', `Starting lab on ${lab.provider}… ${duration(since)}`)
      bar.replaceChildren(status)
      poll = window.setTimeout(refresh, 1000)
    } else if (lab.state === 'failed') {
      setStatus(status, 'error', `Lab failed: ${lab.error}`)
      bar.replaceChildren(status)
    } else {
      bar.replaceChildren()
    }
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
  async function end() {
    try {
      await endLab()
    } catch (err) {
      failed(err)
      return
    }
    refresh()
  }
  refresh()
  return { refresh }
}

function button(label: string, onclick: () => void): HTMLButtonElement {
  const b = el('button', 'action', label)
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
