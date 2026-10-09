import './style.css'
import { openTerminal } from './terminal'
import { showTasks } from './tasks'

openTerminal(document.querySelector<HTMLElement>('#terminal')!)
showTasks(document.querySelector<HTMLElement>('#question')!)
