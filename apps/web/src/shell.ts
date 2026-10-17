import type { Lab, LabState } from './api'
import { openTerminal } from './terminal'
import { el } from './ui'

const waiting: Record<Exclude<LabState, 'ready'>, string> = {
  none: 'Start the lab to get a shell on the cluster here.',
  starting: 'Building the cluster. The shell opens here once it is ready.',
  stopping: 'Stopping the lab…',
  failed: 'The lab is down. Start it again to get a shell.',
}

export function shellPanel(signal: AbortSignal, onDisconnect: () => void): { panel: HTMLElement; follow: (lab: Lab) => void } {
  const screen = el('div', 'shell-screen')
  const hint = el('p', 'shell-hint', 'Checking the lab…')
  const reconnect = el('button', 'action', 'Reconnect')
  reconnect.type = 'button'
  reconnect.hidden = true
  const bar = el('div', 'shell-bar')
  bar.append(el('span', 'shell-title', 'Terminal'), reconnect)
  const panel = el('section', 'shell')
  panel.dataset.live = 'false'
  panel.append(bar, screen, hint)

  let close: (() => void) | undefined
  const shut = () => {
    close?.()
    close = undefined
    reconnect.hidden = true
  }
  const open = () => {
    shut()
    close = openTerminal(screen, '/ws/terminal', () => {
      reconnect.hidden = false
      onDisconnect()
    })
  }
  reconnect.onclick = open
  signal.addEventListener('abort', shut)

  // A dropped shell keeps its close until Reconnect, so a lab that is still ready never redials on its own.
  const follow = (lab: Lab) => {
    panel.dataset.live = String(lab.state === 'ready')
    if (lab.state !== 'ready') {
      shut()
      hint.textContent = waiting[lab.state]
    } else if (!close) open()
  }
  return { panel, follow }
}
