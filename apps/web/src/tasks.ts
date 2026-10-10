import { beginExam, checkTask, listTasks, startTask, taskQuestion, taskSolution, type Task } from './api'
import { showExam } from './exam'
import { checkList, chips, el, hostLine, markdown, message, setStatus } from './ui'

export async function showTasks(pane: HTMLElement) {
  pane.replaceChildren(el('p', 'muted', 'Loading tasks…'))
  let tasks: Task[]
  try {
    tasks = await listTasks()
  } catch (err) {
    pane.replaceChildren(el('p', 'error', `Cannot load tasks: ${message(err)}`))
    return
  }

  const list = el('ol', 'task-list')
  tasks.forEach((task, i) => {
    const button = el('button', 'task')
    button.append(el('span', 'task-title', `${i + 1}. ${task.title}`), chips(task))
    button.onclick = () => openTask(pane, task, i + 1)
    const item = el('li')
    item.append(button)
    list.append(item)
  })
  pane.replaceChildren(examStarter(pane, tasks.length), el('h2', '', 'Tasks'), list)
}

function examStarter(pane: HTMLElement, available: number): HTMLElement {
  const box = el('section', 'exam-start')
  const open = el('button', 'action primary', 'Start exam')
  const form = el('form', 'exam-form')
  form.hidden = true
  const count = number('Tasks', 1, available, Math.min(16, available))
  const minutes = number('Minutes', 1, 240, 120)
  const start = el('button', 'action primary', 'Start')
  start.type = 'submit'
  const error = el('p', 'error')
  form.append(count.label, minutes.label, start, error)

  open.onclick = () => {
    form.hidden = !form.hidden
  }
  form.onsubmit = async (event) => {
    event.preventDefault()
    start.disabled = true
    error.textContent = ''
    try {
      await beginExam(count.input.valueAsNumber, minutes.input.valueAsNumber)
      showExam(pane)
    } catch (err) {
      error.textContent = `Cannot start the exam: ${message(err)}`
      start.disabled = false
    }
  }
  box.append(open, form)
  return box
}

function number(text: string, min: number, max: number, value: number) {
  const label = el('label', '', text)
  const input = el('input')
  Object.assign(input, { type: 'number', min: String(min), max: String(max), value: String(value), required: true })
  label.append(input)
  return { label, input }
}

function openTask(pane: HTMLElement, task: Task, number: number) {
  // Aborting the request stops its script on the server, so leaving never blocks the next task.
  let running = new AbortController()
  const restart = () => {
    running.abort()
    running = new AbortController()
    return running.signal
  }
  const back = el('button', 'back', '← Back to tasks')
  back.onclick = () => {
    running.abort()
    showTasks(pane)
  }
  const host = hostLine(task)
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
  pane.replaceChildren(back, el('h2', '', `${number}. ${task.title}`), chips(task), host, status, actions, result, body, explain)

  const prepare = () => {
    check.disabled = reset.disabled = true
    result.replaceChildren()
    setStatus(status, '', 'Preparing the cluster…')
    const signal = restart()
    startTask(task.id, signal).then(
      () => {
        setStatus(status, 'ready', 'Ready: the cluster is set up.')
        check.disabled = false
      },
      (err) => signal.aborted || setStatus(status, 'error', `Setup failed: ${message(err)}`),
    ).finally(() => (reset.disabled = false))
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
