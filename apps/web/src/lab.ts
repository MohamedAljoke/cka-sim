import { getLab, startLab, stopLab, type Lab } from './api'
import { el, message } from './ui'

const pollEvery = 1000

export type LabStatus = {
  panel: HTMLElement
  refresh: () => void
  watch: (signal: AbortSignal, onChange: (lab: Lab) => void) => void
}

export function labStatus(): LabStatus {
  const dot = el('span', 'lab-dot')
  const text = el('span')
  const line = el('p', 'lab-line')
  line.append(dot, text)
  const action = el('button', 'action')
  action.type = 'button'
  const panel = el('section', 'lab-panel')
  panel.setAttribute('aria-label', 'Lab')
  panel.setAttribute('aria-live', 'polite')
  panel.append(line, action)

  let timer: ReturnType<typeof setTimeout> | undefined
  let current: Lab | undefined
  const watchers = new Set<(lab: Lab) => void>()

  const offer = (label: string, run: () => Promise<void>) => {
    action.hidden = false
    action.textContent = label
    action.onclick = run
  }
  const show = (lab: Lab) => {
    current = lab
    panel.dataset.state = lab.state
    text.className = ''
    action.hidden = true
    switch (lab.state) {
      case 'none':
        text.textContent = 'No lab running'
        offer('Start lab', () => change(startLab))
        break
      case 'starting':
        text.textContent = ['Starting the lab…', lab.step, since(lab.startedAt)].filter(Boolean).join(' · ')
        break
      case 'ready':
        text.textContent = lab.startedAt && lab.readyAt ? `Lab ready · built in ${duration(lab.startedAt, lab.readyAt)}` : 'Lab ready'
        offer('Stop lab', () => change(stopLab))
        break
      case 'stopping':
        text.textContent = 'Stopping the lab…'
        break
      case 'failed':
        text.className = 'error'
        text.textContent = lab.error || 'The lab failed.'
        offer('Start again', () => change(startLab))
        break
    }
    if (lab.state === 'starting' || lab.state === 'stopping') timer = setTimeout(refresh, pollEvery)
    watchers.forEach((onChange) => onChange(lab))
  }
  const failed = (err: unknown) => {
    panel.dataset.state = 'failed'
    text.className = 'error'
    text.textContent = `Cannot reach the lab: ${message(err)}`
    offer('Retry', refresh)
  }
  async function refresh() {
    clearTimeout(timer)
    try {
      show(await getLab())
    } catch (err) {
      failed(err)
    }
  }
  async function change(run: () => Promise<Lab>) {
    clearTimeout(timer)
    action.disabled = true
    try {
      show(await run())
    } catch (err) {
      failed(err)
    } finally {
      action.disabled = false
    }
  }
  const watch = (signal: AbortSignal, onChange: (lab: Lab) => void) => {
    watchers.add(onChange)
    signal.addEventListener('abort', () => watchers.delete(onChange))
    if (current) onChange(current)
  }

  text.textContent = 'Checking the lab…'
  action.hidden = true
  refresh()
  return { panel, refresh, watch }
}

function since(startedAt?: string): string {
  return startedAt ? duration(startedAt, new Date().toISOString()) : ''
}

function duration(from: string, to: string): string {
  const seconds = Math.max(0, Math.round((Date.parse(to) - Date.parse(from)) / 1000))
  return seconds < 60 ? `${seconds}s` : `${Math.floor(seconds / 60)}m ${seconds % 60}s`
}
