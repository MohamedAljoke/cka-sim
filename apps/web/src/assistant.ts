import { startAssistant } from './api'
import { icon } from './icons'
import { openTerminal } from './terminal'
import { el, message, setStatus } from './ui'

export type Provider = {
  id: string
  label: string
  intro: string
  prepare(): Promise<void>
  mount(screen: HTMLElement, ended: () => void): () => void
}

function cli(id: string, label: string, intro: string): Provider {
  return {
    id,
    label,
    intro,
    prepare: () => startAssistant(id),
    mount: (screen, ended) => openTerminal(screen, `/ws/assistant?provider=${encodeURIComponent(id)}`, ended),
  }
}

const providers: Provider[] = [
  cli(
    'claude',
    'Claude · your subscription',
    'Claude Code opens here and signs in with your own Claude Pro or Max plan, once per lab. It can read the task and your cluster and give hints, but it cannot change anything.',
  ),
]

export type Assistant = {
  lab(ready: boolean): void
  end(): void
}

export function showAssistant(root: HTMLElement): Assistant {
  let provider = providers[0]
  let labReady = false
  let attempt = 0
  let close: (() => void) | undefined

  const toggle = el('button', 'assistant-toggle')
  toggle.type = 'button'
  const picker = el('select', 'assistant-picker')
  picker.setAttribute('aria-label', 'AI provider')
  picker.append(...providers.map((p) => new Option(p.label, p.id)))
  const status = el('span', 'status')
  const connect = el('button', 'action primary')
  connect.type = 'button'
  const disconnect = el('button', 'action', 'Disconnect')
  disconnect.type = 'button'
  const head = el('div', 'assistant-head')
  const body = el('div', 'assistant-body')
  root.replaceChildren(head, body)

  const expand = (open: boolean) => {
    root.classList.toggle('open', open)
    toggle.setAttribute('aria-expanded', String(open))
    toggle.replaceChildren(icon(open ? 'chevron-down' : 'chevron-up', 14), 'Assistant')
  }

  const end = (error = '') => {
    attempt++
    close?.()
    close = undefined
    picker.disabled = false
    connect.disabled = !labReady
    connect.textContent = error ? 'Retry' : 'Connect'
    head.replaceChildren(toggle, picker, status)
    setStatus(status, '', labReady ? '' : 'Starts once your lab is ready')
    body.className = 'assistant-body'
    body.replaceChildren(el('p', 'muted', provider.intro), connect)
    if (error) body.append(el('p', 'error', error))
  }

  const start = async () => {
    const mine = ++attempt
    picker.disabled = connect.disabled = true
    setStatus(status, '', 'Connecting…')
    try {
      await provider.prepare()
    } catch (err) {
      if (mine === attempt) end(message(err))
      return
    }
    if (mine !== attempt) return
    const screen = el('div', 'assistant-screen')
    body.className = 'assistant-body terminal'
    body.replaceChildren(screen)
    head.replaceChildren(toggle, picker, status, disconnect)
    setStatus(status, '', '')
    expand(true)
    close = provider.mount(screen, () => end('The assistant disconnected.'))
  }

  toggle.onclick = () => expand(!root.classList.contains('open'))
  picker.onchange = () => {
    provider = providers.find((p) => p.id === picker.value) ?? providers[0]
    end()
  }
  connect.onclick = start
  disconnect.onclick = () => end()

  expand(false)
  end()
  return {
    lab(ready) {
      labReady = ready
      end()
    },
    end: () => end(),
  }
}
