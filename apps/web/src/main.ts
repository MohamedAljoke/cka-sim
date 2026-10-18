import './style.css'
import { getStudent } from './api'
import { showExam } from './exam'
import { labStatus, type LabStatus } from './lab'
import { parseRoute } from './routes'
import { showTask } from './task'
import { showTasks } from './tasks'
import { el } from './ui'

const pane = document.querySelector<HTMLElement>('#app')!
const topbar = document.querySelector<HTMLElement>('.topbar')!
let leaving: AbortController | undefined

function show(lab: LabStatus) {
  leaving?.abort()
  leaving = new AbortController()
  window.scrollTo(0, 0)
  const route = parseRoute(location.hash)
  document.body.dataset.page = route.page
  if (route.page === 'task') showTask(pane, route.id, lab, leaving.signal)
  else if (route.page === 'exam') showExam(pane, lab, leaving.signal)
  else showTasks(pane, route.filter, leaving.signal)
}

function needsToken() {
  const card = el('section', 'card needs-token')
  card.append(
    el('h1', '', 'Open your student link'),
    el('p', '', 'This site gives each student their own lab. Open the link you were given in this browser.'),
  )
  pane.replaceChildren(card)
}

async function start() {
  const student = await getStudent().catch(() => '')
  if (student === null) {
    needsToken()
    return
  }
  // The one person on their own machine is "local"; a name only helps when the site has several.
  if (student && student !== 'local') topbar.append(el('span', 'student', student))
  const lab = labStatus()
  topbar.append(lab.panel)
  window.addEventListener('hashchange', () => show(lab))
  show(lab)
}

start()
