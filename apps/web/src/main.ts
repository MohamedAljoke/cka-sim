import '@fontsource-variable/ibm-plex-sans'
import '@fontsource/barlow-semi-condensed/600.css'
import '@fontsource/barlow-semi-condensed/700.css'
import './style.css'
import { openTerminal } from './terminal'
import { showExam } from './exam'
import { showLab } from './lab'
import { el } from './ui'

const question = document.querySelector<HTMLElement>('#question')!
const terminal = document.querySelector<HTMLElement>('#terminal')!
let closeTerminal: (() => void) | undefined

const waiting = () => question.replaceChildren(el('p', 'muted', 'Start a lab to practise or take an exam.'))
waiting()
showLab(document.querySelector<HTMLElement>('#lab')!, {
  ready() {
    closeTerminal = openTerminal(terminal)
    showExam(question)
  },
  ended() {
    closeTerminal?.()
    closeTerminal = undefined
    waiting()
  },
})
