import { checkTask, startTask, taskQuestion, taskSolution, tidyTask, type Task } from './api'
import type { App } from './app'
import { checkList, chips, el, hostLine, markdown, message, setStatus } from './ui'

export function showPractice(app: App, task: Task) {
  const { pane } = app
  app.split(true)
  // Aborting the request stops its script on the server, so leaving never blocks the next task.
  let running = new AbortController()
  const restart = () => {
    running.abort()
    running = new AbortController()
    return running.signal
  }
  // Leaving tidies the task, so the next setup finds nothing to delete.
  const tidy = () => tidyTask(task.id)
  window.addEventListener('pagehide', tidy)
  app.onLeave(() => {
    running.abort()
    window.removeEventListener('pagehide', tidy)
    tidy()
  })

  const home = el('button', 'back', '← Home')
  home.onclick = () => app.home()
  const status = el('p', 'status')
  const check = el('button', 'action primary', 'Check')
  const solution = el('button', 'action', 'Solution')
  const reset = el('button', 'action', 'Reset')
  const actions = el('div', 'actions')
  actions.append(check, solution, reset)
  const result = el('div', 'result')
  const body = el('div', 'question-text')
  const explain = el('section', 'question-text solution')
  explain.hidden = true
  pane.replaceChildren(home, el('h2', '', task.title), chips(task), hostLine(task), status, actions, result, body, explain)

  const prepare = () => {
    check.disabled = reset.disabled = true
    result.replaceChildren()
    setStatus(status, '', 'Preparing the cluster (the lab starts first if there is none)…')
    const signal = restart()
    startTask(task.id, signal).then(
      () => {
        setStatus(status, 'ready', 'Ready: the cluster is set up.')
        check.disabled = false
      },
      (err) => signal.aborted || setStatus(status, 'error', `Setup failed: ${message(err)}`),
    ).finally(() => (reset.disabled = false))
    app.labChanged()
  }

  check.onclick = () => {
    check.disabled = reset.disabled = true
    result.replaceChildren(el('p', 'muted', 'Checking…'))
    const signal = restart()
    checkTask(task.id, signal)
      .then(
        (r) => result.replaceChildren(checkList(r)),
        (err) => signal.aborted || result.replaceChildren(el('p', 'error', `Check failed: ${message(err)}`)),
      )
      .finally(() => (check.disabled = reset.disabled = false))
    app.labChanged()
  }

  solution.onclick = async () => {
    if (!explain.hidden) {
      explain.hidden = true
      return
    }
    if (!explain.hasChildNodes()) {
      try {
        const text = await taskSolution(task.id)
        explain.append(el('h3', '', 'Solution'))
        markdown(explain, text)
      } catch (err) {
        result.replaceChildren(el('p', 'error', `Cannot load the solution: ${message(err)}`))
        return
      }
    }
    explain.hidden = false
  }

  reset.onclick = prepare
  prepare()
  taskQuestion(task.id).then(
    (question) => markdown(body, question),
    (err) => setStatus(status, 'error', `Cannot load the question: ${message(err)}`),
  )
}
