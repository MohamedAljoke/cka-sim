import '@fontsource-variable/ibm-plex-sans'
import '@fontsource/barlow-semi-condensed/600.css'
import '@fontsource/barlow-semi-condensed/700.css'
import './style.css'
import type { App } from './app'
import { showExam } from './exam'
import { showHome } from './home'
import { showLab } from './lab'
import { showPractice } from './practice'
import { openTerminal } from './terminal'

const layout = document.querySelector<HTMLElement>('main.exam')!
const pane = document.querySelector<HTMLElement>('#question')!
const terminal = document.querySelector<HTMLElement>('#terminal')!
let closeTerminal: (() => void) | undefined
let labReady = false
let cleanups: (() => void)[] = []

// The terminal opens once a split view has a ready lab, and stays open until the lab ends.
const syncTerminal = () => {
  if (labReady && !layout.classList.contains('single') && !closeTerminal) closeTerminal = openTerminal(terminal)
}

const go = (show: () => void) => {
  cleanups.forEach((cleanup) => cleanup())
  cleanups = []
  show()
}

const lab = showLab(document.querySelector<HTMLElement>('#lab')!, {
  ready() {
    labReady = true
    syncTerminal()
  },
  ended() {
    labReady = false
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
  labChanged: () => lab.refresh(),
  onLeave: (cleanup) => cleanups.push(cleanup),
}

// A running or finished exam decides the first view; with none, showExam goes home.
app.exam()
