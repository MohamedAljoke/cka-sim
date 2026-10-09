import '@fontsource-variable/ibm-plex-sans'
import '@fontsource/barlow-semi-condensed/600.css'
import '@fontsource/barlow-semi-condensed/700.css'
import './style.css'
import { openTerminal } from './terminal'
import { showTasks } from './tasks'

openTerminal(document.querySelector<HTMLElement>('#terminal')!)
showTasks(document.querySelector<HTMLElement>('#question')!)
