import { endExam, flagTask, getExam, listTasks, taskQuestion, taskSolution, type ExamState, type ExamTask, type Task } from './api'
import type { App } from './app'
import { checkList, chips, el, hostLine, markdown, message } from './ui'

// showExam shows whatever the exam is doing; with no exam it goes home. The catalog is loaded
// once here: the views below poll only the exam.
export async function showExam(app: App) {
  let state: ExamState | null
  let catalog: Map<string, Task>
  try {
    ;[state, catalog] = await Promise.all([getExam(), listTasks().then((all) => new Map(all.map((t) => [t.id, t])))])
  } catch (err) {
    app.split(false)
    app.pane.replaceChildren(el('p', 'error', `Cannot load the exam: ${message(err)}`))
    return
  }
  if (!state) app.home()
  else if (state.exam.ended) showResults(app, state, catalog)
  else if (!state.exam.started) showPreparing(app, state, catalog)
  else showRunning(app, state, catalog)
}

const settingUp = (exam: ExamState['exam']) => exam.tasks.some((t) => t.setup === 'preparing')

function showPreparing(app: App, state: ExamState, catalog: Map<string, Task>) {
  app.split(true)
  const status = el('p', 'status')
  const list = el('ol', 'setup-list')
  app.pane.replaceChildren(el('h2', '', 'Preparing your exam'), status, list)
  let poll: number | undefined
  app.onLeave(() => clearTimeout(poll))

  const draw = (state: ExamState) => {
    const tasks = state.exam.tasks
    const done = tasks.filter((t) => t.setup !== 'preparing').length
    status.textContent = `${done} of ${tasks.length} tasks set up. The clock starts once the quick ones are ready; the slow ones finish while you read.`
    list.replaceChildren(
      ...tasks.map((t, i) => {
        const row = el('li', `setup-row ${t.setup}`)
        const label = { preparing: 'setting up…', ready: 'ready', failed: 'failed' }[t.setup]
        row.append(el('span', 'number', number(i)), el('span', 'row-title', catalog.get(t.id)!.title), el('span', 'setup-state', label))
        return row
      }),
    )
  }
  const refresh = async () => {
    try {
      const next = await getExam()
      if (!next) return app.home()
      if (next.exam.started) return showRunning(app, next, catalog)
      draw(next)
    } catch (err) {
      status.textContent = `Cannot load the exam: ${message(err)}`
    }
    poll = window.setTimeout(refresh, 2000)
  }
  draw(state)
  poll = window.setTimeout(refresh, 2000)
}

function showRunning(app: App, state: ExamState, catalog: Map<string, Task>) {
  app.split(true)
  const exam = state.exam
  const timer = el('span', 'timer')
  const end = el('button', 'action', 'End exam')
  const bar = el('div', 'exam-bar')
  bar.append(timer, end)
  const pills = el('nav', 'pills')
  const question = el('div', 'exam-question')
  app.pane.replaceChildren(bar, pills, question)

  let offset = Date.parse(state.now) - Date.now()
  let deadline = Date.parse(exam.deadline!)
  const tick = () => {
    const left = deadline - (Date.now() + offset)
    timer.textContent = left >= 0 ? clock(left) : `Time's up +${clock(-left)}`
    timer.classList.toggle('low', left < 10 * 60_000)
  }
  tick()
  const ticker = window.setInterval(tick, 1000)
  let poll: number | undefined
  app.onLeave(() => {
    clearInterval(ticker)
    clearTimeout(poll)
  })

  const syncEnd = () => {
    end.disabled = settingUp(exam)
    end.title = end.disabled ? 'Available once every task is set up' : ''
  }
  let current = 0
  const drawPills = () => {
    pills.replaceChildren(
      ...exam.tasks.map((t, i) => {
        const preparing = t.setup === 'preparing'
        const pill = el('button', 'pill', `${number(i)}${preparing ? ' …' : ''}${t.flagged ? ' ⚑' : ''}`)
        pill.classList.toggle('current', i === current)
        pill.classList.toggle('flagged', t.flagged)
        pill.classList.toggle('preparing', preparing)
        pill.classList.toggle('failed', t.setup === 'failed')
        pill.title = preparing ? 'Still being set up' : ''
        pill.onclick = () => open(i)
        return pill
      }),
    )
  }
  const open = (i: number) => {
    current = i
    drawPills()
    showQuestion(question, exam.tasks[i], catalog.get(exam.tasks[i].id)!, i, drawPills)
  }
  open(0)
  syncEnd()

  // The slow tasks finish setting up after the clock starts; their time is added to the deadline.
  // Only setup fields are taken from the poll, so a flag set meanwhile isn't undone.
  const refresh = async () => {
    try {
      const next = await getExam()
      if (!next || next.exam.ended) return app.exam()
      const was = exam.tasks[current].setup
      next.exam.tasks.forEach((t, i) => Object.assign(exam.tasks[i], { setup: t.setup, setupError: t.setupError }))
      offset = Date.parse(next.now) - Date.now()
      deadline = Date.parse(next.exam.deadline!)
      tick()
      drawPills()
      syncEnd()
      if (exam.tasks[current].setup !== was) open(current)
    } catch {
      // The next poll tries again.
    }
    if (settingUp(exam)) poll = window.setTimeout(refresh, 2000)
  }
  if (settingUp(exam)) poll = window.setTimeout(refresh, 2000)

  end.onclick = async () => {
    if (!confirm('End the exam? Every task is scored in the background.')) return
    end.disabled = true
    try {
      showResults(app, await endExam(), catalog)
    } catch (err) {
      syncEnd()
      question.prepend(el('p', 'error', `Cannot end the exam: ${message(err)}`))
    }
  }
}

function showQuestion(target: HTMLElement, entry: ExamTask, task: Task, index: number, onFlag: () => void) {
  const flag = el('button', 'action')
  const label = () => (flag.textContent = entry.flagged ? '⚑ Flagged' : 'Flag for later')
  label()
  flag.onclick = async () => {
    flag.disabled = true
    try {
      await flagTask(entry.id, !entry.flagged)
      entry.flagged = !entry.flagged
      label()
      onFlag()
    } finally {
      flag.disabled = false
    }
  }
  const head = el('div', 'exam-question-head')
  head.append(el('h2', '', `${number(index)}. ${task.title}`), flag)
  const body = el('div', 'question-text')
  target.replaceChildren(head, chips(task), hostLine(task))
  if (entry.setup === 'preparing') {
    target.append(el('p', 'status', 'This task is still being set up; it opens in a moment.'))
    return
  }
  if (entry.setup === 'failed') target.append(el('p', 'status error', `Setup failed: ${entry.setupError}`))
  target.append(body)
  taskQuestion(task.id).then(
    (text) => markdown(body, text),
    (err) => body.append(el('p', 'error', `Cannot load the question: ${message(err)}`)),
  )
}

// showResults fills in each task as its check comes back, polling until the exam is scored.
function showResults(app: App, state: ExamState, catalog: Map<string, Task>) {
  app.split(false)
  const opened = new Set<string>()
  let poll: number | undefined
  app.onLeave(() => clearTimeout(poll))

  const draw = (state: ExamState) => {
    const scoring = !state.exam.scored
    const head = el('div', 'results-head')
    head.append(el('h2', '', 'Exam results'))
    if (state.score) {
      const { percent, passed } = state.score
      head.append(el('p', `verdict ${passed ? 'pass' : 'fail'}`, `${percent}% · ${passed ? 'PASS' : 'FAIL'}`))
      if (scoring) head.append(el('p', 'status', 'Tidying up the cluster… Try again unlocks when it is done.'))
    } else {
      head.append(el('p', 'status', 'Scoring… each task fills in as its check comes back.'))
    }
    const rows = el('div', 'results')
    state.exam.tasks.forEach((entry, i) => rows.append(resultRow(app, entry, catalog.get(entry.id)!, i, scoring, opened)))
    const home = el('button', 'action primary', 'Home')
    home.onclick = () => app.home()
    app.pane.replaceChildren(head, el('p', 'muted', 'The CKA pass mark is 66%, weighted by task.'), rows, home)
    if (scoring) poll = window.setTimeout(refresh, 2000)
  }
  const refresh = async () => {
    try {
      const next = await getExam()
      if (next) draw(next)
    } catch (err) {
      app.pane.append(el('p', 'error', `Cannot load the results: ${message(err)}`))
      poll = window.setTimeout(refresh, 2000)
    }
  }
  draw(state)
}

function resultRow(app: App, entry: ExamTask, task: Task, index: number, scoring: boolean, opened: Set<string>) {
  const r = entry.result
  const wrong = !!r && r.earned < r.total
  const row = el('details', `result-row${wrong ? ' wrong' : ''}`)
  const summary = el('summary')
  const points = r ? el('span', 'points', `${r.earned} / ${r.total}`) : el('span', 'points pending', 'checking…')
  summary.append(el('span', 'number', number(index)), el('span', 'row-title', task.title), points)
  row.append(summary)
  if (!r) return row

  // Wrong answers start open; what the user opens or closes stays that way across refreshes.
  row.open = opened.has(task.id) || (wrong && !opened.has(`closed:${task.id}`))
  row.ontoggle = () => {
    opened.delete(task.id)
    opened.delete(`closed:${task.id}`)
    opened.add(row.open ? task.id : `closed:${task.id}`)
    if (row.open) loadSolution()
  }
  if (entry.checkError) row.append(el('p', 'error', `Check failed: ${entry.checkError}`))
  row.append(checkList(r))
  const retry = el('button', 'action', 'Try again')
  retry.disabled = scoring
  retry.title = scoring ? 'Available once every task is scored' : 'Practise this task'
  retry.onclick = () => app.practice(task)
  row.append(retry)
  const solution = el('section', 'question-text solution')
  row.append(solution)
  const loadSolution = () => {
    if (solution.hasChildNodes()) return
    solution.append(el('h3', '', 'Solution'))
    taskSolution(task.id).then(
      (text) => markdown(solution, text),
      (err) => solution.append(el('p', 'error', `Cannot load the solution: ${message(err)}`)),
    )
  }
  if (row.open) loadSolution()
  return row
}

const number = (i: number) => String(i + 1).padStart(2, '0')

function clock(ms: number): string {
  const s = Math.floor(ms / 1000)
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${h}:${pad(m)}:${pad(s % 60)}`
}
