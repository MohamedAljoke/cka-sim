import '@fontsource-variable/ibm-plex-sans'
import '@fontsource/barlow-semi-condensed/600.css'
import '@fontsource/barlow-semi-condensed/700.css'
import './style.css'
import { openTerminal } from './terminal'
import { showExam } from './exam'

openTerminal(document.querySelector<HTMLElement>('#terminal')!)
showExam(document.querySelector<HTMLElement>('#question')!)
