import { chatAssistant, startAssistant, type ChatMessage } from './api'
import { icon } from './icons'
import { openTerminal } from './terminal'
import { el, message, setStatus, untrustedMarkdown } from './ui'

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

function chat(id: string, label: string, intro: string): Provider {
  return { id, label, intro, prepare: async () => {}, mount: (screen) => openChat(screen, id) }
}

const chatHistoryLimit = 39

function openChat(screen: HTMLElement, provider: string): () => void {
  const history: ChatMessage[] = []
  const controller = new AbortController()
  const log = el('div', 'chat-log')
  const input = el('textarea', 'chat-input')
  input.rows = 2
  input.placeholder = 'Ask for a review or a hint. Enter sends, Shift+Enter adds a line.'
  input.setAttribute('aria-label', 'Message to the assistant')
  const send = el('button', 'action primary', 'Send')
  send.type = 'button'
  const form = el('div', 'chat-form')
  form.append(input, send)
  screen.classList.add('chat')
  screen.replaceChildren(log, form)

  const bubble = (kind: string) => {
    const b = el('div', `chat-bubble ${kind}`)
    log.append(b)
    b.scrollIntoView({ block: 'end' })
    return b
  }

  const ask = async () => {
    const question = input.value.trim()
    if (!question || send.disabled) return
    input.value = ''
    send.disabled = true
    history.push({ role: 'user', text: question })
    bubble('user').textContent = question
    const answer = bubble('model muted')
    let text = ''
    let activity = 'Thinking…'
    let frame = 0
    const render = () => {
      frame = 0
      answer.className = text ? 'chat-bubble model' : 'chat-bubble model muted'
      answer.replaceChildren()
      if (text) untrustedMarkdown(answer, text)
      if (activity) answer.append(el('p', 'chat-tools', activity))
      answer.scrollIntoView({ block: 'end' })
    }
    const renderSoon = () => {
      frame ||= requestAnimationFrame(render)
    }
    render()
    try {
      const turn = await chatAssistant(provider, history.slice(-chatHistoryLimit), controller.signal, {
        text: (delta) => {
          text += delta
          activity = ''
          renderSoon()
        },
        tool: (name) => {
          activity = `Checking: ${name}…`
          renderSoon()
        },
      })
      cancelAnimationFrame(frame)
      history.push({ role: 'model', text: turn.text })
      text = turn.text
      activity = turn.tools.length ? `Used ${[...new Set(turn.tools)].join(', ')}` : ''
      render()
    } catch (err) {
      cancelAnimationFrame(frame)
      if (controller.signal.aborted) return
      history.pop()
      answer.className = 'chat-bubble error'
      answer.textContent = message(err)
      input.value ||= question
    } finally {
      send.disabled = false
      answer.scrollIntoView({ block: 'end' })
    }
  }

  send.onclick = ask
  input.onkeydown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
      e.preventDefault()
      ask()
    }
  }
  input.focus()
  return () => controller.abort()
}

const providers: Provider[] = [
  cli(
    'claude',
    'Claude · your subscription',
    'Claude Code opens here and signs in with your own Claude Pro or Max plan, once per lab. It can read the task and your cluster and give hints, but it cannot change anything.',
  ),
  chat(
    'gemini',
    'Gemini · cka-sim',
    'Gemini runs on cka-sim, so there is nothing to sign in to. It can read the task and your cluster and give hints, but it cannot change anything.',
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
