import '@fontsource-variable/instrument-sans'
import '@fontsource-variable/fraunces/opsz.css'
import '@fontsource-variable/jetbrains-mono'
import './style.css'
import { currentUser, type User } from './api'
import type { App } from './app'
import { showAssistant } from './assistant'
import { showAccount, showLogin } from './auth'
import { showExam } from './exam'
import { showHome } from './home'
import { showLab } from './lab'
import { showPractice } from './practice'
import { themeToggle } from './theme'
import { openTerminal } from './terminal'
import { el, message } from './ui'

themeToggle(document.querySelector<HTMLButtonElement>('#theme')!)

const layout = document.querySelector<HTMLElement>('main.exam')!
const pane = document.querySelector<HTMLElement>('#question')!

// Nothing of the app shows, or asks the platform anything, until the auth server knows who you are.
async function boot() {
  let user: User | null
  try {
    user = await currentUser()
  } catch (err) {
    const retry = el('button', 'action', 'Retry')
    retry.type = 'button'
    retry.onclick = boot
    pane.replaceChildren(el('p', 'error', `Cannot reach the auth server: ${message(err)}`), retry)
    return
  }
  if (user) start(user)
  else showLogin(pane, start)
}

function start(user: User) {
  showAccount(document.querySelector<HTMLElement>('#account')!, user)
  const terminal = document.querySelector<HTMLElement>('#terminal')!
  let closeTerminal: (() => void) | undefined
  let labReady = false
  let cleanups: (() => void)[] = []

  // The terminal opens once a split view has a ready lab, and stays open until the lab ends.
  const syncTerminal = () => {
    if (labReady && !layout.classList.contains('single') && !closeTerminal) closeTerminal = openTerminal(terminal, '/ws/terminal', () => lab.refresh())
  }

  const assistant = showAssistant(document.querySelector<HTMLElement>('#assistant')!)

  const go = (show: () => void) => {
    cleanups.forEach((cleanup) => cleanup())
    cleanups = []
    app.assist(false)
    show()
  }

  const lab = showLab(document.querySelector<HTMLElement>('#lab')!, {
    ready() {
      labReady = true
      syncTerminal()
      assistant.lab(true)
    },
    ended() {
      labReady = false
      assistant.lab(false)
      closeTerminal?.()
      closeTerminal = undefined
    },
  })

  const app: App = {
    pane,
    home: () => go(() => showHome(app)),
    practice: (task) => go(() => showPractice(app, task)),
    exam: () => go(() => showExam(app)),
    split(on) {
      layout.classList.toggle('single', !on)
      syncTerminal()
    },
    assist(on) {
      if (!on) assistant.end()
      layout.classList.toggle('assist', on)
    },
    labChanged: () => lab.refresh(),
    onLeave: (cleanup) => cleanups.push(cleanup),
  }

  // A running or finished exam decides the first view; with none, showExam goes home.
  app.exam()
}

boot()
