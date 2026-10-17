import './style.css'
import { labStatus } from './lab'
import { parseRoute } from './routes'
import { showTask } from './task'
import { showTasks } from './tasks'

const pane = document.querySelector<HTMLElement>('#app')!
const lab = labStatus()
document.querySelector('.topbar')!.append(lab.panel)
let leaving: AbortController | undefined

function show() {
  leaving?.abort()
  leaving = new AbortController()
  window.scrollTo(0, 0)
  const route = parseRoute(location.hash)
  document.body.dataset.page = route.page
  if (route.page === 'task') showTask(pane, route.id, lab, leaving.signal)
  else showTasks(pane, route.filter, leaving.signal)
}

window.addEventListener('hashchange', show)
show()
