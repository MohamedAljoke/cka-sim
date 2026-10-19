import { chat, type ChatMessage } from './api'
import { el, message, resizer, untrustedMarkdown } from './ui'

const toolLabels: Record<string, string> = {
  current_task: 'Read your task',
  run_check: 'Ran the check',
  inspect_cluster: 'Looked at your cluster',
  explain: 'Read the solution',
}

export function tutorPanel(taskId: string, signal: AbortSignal): HTMLElement {
  const log = el('div', 'tutor-log')
  log.setAttribute('role', 'log')
  log.setAttribute('aria-live', 'polite')
  log.hidden = true

  const input = el('textarea', 'tutor-input')
  input.rows = 2
  input.placeholder = 'Why does my check fail? What is a PodDisruptionBudget?'
  input.setAttribute('aria-label', 'Your question')
  const ask = el('button', 'action primary', 'Ask')
  ask.type = 'submit'
  const form = el('form', 'tutor-form')
  form.append(input, ask)

  const toggle = el('button', 'tutor-summary')
  toggle.type = 'button'
  toggle.append(
    el('span', 'tutor-title', 'Ask the tutor'),
    el('span', 'muted tutor-intro', 'It reads your task and cluster, then helps you find the answer yourself'),
  )
  const body = el('div', 'tutor-body')
  body.id = `tutor-${taskId}`
  body.append(log, form)
  toggle.setAttribute('aria-controls', body.id)
  const panel = el('section', 'tutor-panel')
  panel.append(
    toggle,
    body,
    resizer({
      label: 'Resize the tutor',
      axis: 'y',
      grow: -1,
      key: 'cka-sim.tutor-height',
      size: () => panel.offsetHeight,
      limits: () => [140, (panel.parentElement?.clientHeight ?? 0) - 260],
      apply: (px) => panel.style.setProperty('--tutor-height', `${px}px`),
    }),
  )
  const show = (open: boolean) => {
    panel.dataset.open = String(open)
    body.hidden = !open
    toggle.setAttribute('aria-expanded', String(open))
  }
  show(false)
  toggle.onclick = () => {
    const open = panel.dataset.open !== 'true'
    show(open)
    if (open) input.focus()
  }

  const history: ChatMessage[] = []
  let busy = false

  const send = async () => {
    const text = input.value.trim()
    if (!text || busy) return
    busy = true
    ask.disabled = true
    input.value = ''
    history.push({ role: 'student', text })

    const tools = el('ul', 'tutor-tools')
    const answer = el('div', 'tutor-text muted', 'Thinking…')
    const reply = el('div', 'tutor-message tutor')
    reply.append(tools, answer)
    log.hidden = false
    log.append(el('p', 'tutor-message student', text), reply)
    reply.scrollIntoView({ block: 'nearest' })

    let said = ''
    try {
      await chat(taskId, history, signal, (event) => {
        if (event.type === 'tool') tools.append(el('li', '', toolLabels[event.name] ?? event.name))
        else {
          said += event.text
          answer.className = 'tutor-text markdown'
          untrustedMarkdown(answer, said)
        }
      })
      if (!said) throw new Error('the tutor gave no answer')
      history.push({ role: 'tutor', text: said })
    } catch (err) {
      if (signal.aborted) return
      // The question stays unanswered in the conversation, so it is taken out and offered again.
      history.pop()
      if (!said) answer.remove()
      reply.append(el('p', 'tutor-text error', `The tutor can't answer: ${message(err)}`))
      if (!input.value) input.value = text
    } finally {
      busy = false
      ask.disabled = false
    }
  }

  form.onsubmit = (event) => {
    event.preventDefault()
    send()
  }
  input.onkeydown = (event) => {
    if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) {
      event.preventDefault()
      send()
    }
  }

  return panel
}
