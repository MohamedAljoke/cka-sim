import './style.css'
import { parseRoute } from './routes'
import { showTask } from './task'
import { showTasks } from './tasks'

const pane = document.querySelector<HTMLElement>('#app')!
let leaving: AbortController | undefined

function show() {
  leaving?.abort()
  leaving = new AbortController()
  window.scrollTo(0, 0)
  const route = parseRoute(location.hash)
  if (route.page === 'task') showTask(pane, route.id, leaving.signal)
  else showTasks(pane, route.filter, leaving.signal)
}

window.addEventListener('hashchange', show)
show()
