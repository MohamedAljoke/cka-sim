import { openTerminal } from './terminal'
import { el } from './ui'

const duringExam = 4409

let made: HTMLElement | undefined
let closeMade: (() => void) | undefined

export function closeClaude() {
  closeMade?.()
}

// claudePanel runs the student's own Claude Code, started only when asked: it signs in with their Claude account and
// reaches nothing but cka-sim's tools. There is one for the whole visit, which each task page
// borrows, so a session that moves the page to another task keeps running.
export function claudePanel(): HTMLElement {
  made ??= newClaudePanel()
  return made
}

function newClaudePanel(): HTMLElement {
  const hint = el('p', 'claude-hint muted')
  const start = el('button', 'action primary', 'Start Claude Code')
  start.type = 'button'
  const status = el('div', 'claude-status')
  status.append(hint, start)
  const screen = el('div', 'claude-screen')
  screen.hidden = true

  const body = el('div', 'tutor-body claude-body')
  body.append(status, screen)

  const idle = (text: string, action: string) => {
    hint.textContent = text
    start.textContent = action
    start.hidden = false
    status.hidden = false
  }
  idle(
    'It runs on your Claude subscription, not on cka-sim. Sign in with your Claude account the first time; ' +
      'it has no shell, and sees your task and cluster only through cka-sim\'s tools.',
    'Start Claude Code',
  )

  let close: (() => void) | undefined
  const shut = () => {
    close?.()
    close = undefined
  }
  closeMade = () => {
    if (!close) return
    shut()
    screen.hidden = true
    idle('Claude Code ended when you left the task.', 'Start again')
  }

  start.onclick = () => {
    shut()
    screen.replaceChildren()
    screen.hidden = false
    status.hidden = true
    close = openTerminal(screen, '/ws/claude', (event) => {
      close = undefined
      if (event.code === 1000) idle('Claude Code ended.', 'Start again')
      else if (event.code === duringExam) idle('There is no Claude Code during an exam.', 'Try again')
      else idle(event.reason ? `Cannot run Claude Code: ${event.reason}.` : 'The connection dropped.', 'Try again')
    })
  }

  return body
}
