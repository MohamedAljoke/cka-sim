import {
  discardExam,
  endExam,
  flagTask,
  getExam,
  listTasks,
  taskQuestion,
  taskSolution,
  type ExamState,
  type ExamTask,
  type Task,
} from './api'
import { showTasks } from './tasks'
import { checkList, chips, el, hostLine, markdown, message } from './ui'

let ticker: number | undefined

export async function showExam(pane: HTMLElement) {
  clearInterval(ticker)
  let state: ExamState | null
  let catalog: Map<string, Task>
  try {
    ;[state, catalog] = await Promise.all([getExam(), listTasks().then((all) => new Map(all.map((t) => [t.id, t])))])
  } catch (err) {
    pane.replaceChildren(el('p', 'error', `Cannot load the exam: ${message(err)}`))
    return
  }
  if (!state) showTasks(pane)
  else if (state.exam.ended) showResults(pane, state, catalog)
  else if (!state.exam.started) showPreparing(pane, state)
  else showRunning(pane, state, catalog)
}

function showPreparing(pane: HTMLElement, state: ExamState) {
  const done = state.exam.tasks.filter((t) => t.setup !== 'preparing').length
  pane.replaceChildren(
    el('h2', '', 'Preparing your exam'),
    el('p', 'status', `${done} of ${state.exam.tasks.length} tasks set up. The clock starts when all are ready.`),
  )
  window.setTimeout(() => showExam(pane), 2000)
}

function showRunning(pane: HTMLElement, state: ExamState, catalog: Map<string, Task>) {
  const exam = state.exam
  const timer = el('span', 'timer')
  const end = el('button', 'action', 'End exam')
  const bar = el('div', 'exam-bar')
  bar.append(timer, end)
  const pills = el('nav', 'pills')
  const question = el('div', 'exam-question')
  pane.replaceChildren(bar, pills, question)

  const offset = Date.parse(state.now) - Date.now()
  const deadline = Date.parse(exam.deadline!)
  const tick = () => {
    const left = deadline - (Date.now() + offset)
    timer.textContent = left >= 0 ? clock(left) : `Time's up +${clock(-left)}`
    timer.classList.toggle('low', left < 10 * 60_000)
  }
  tick()
  ticker = window.setInterval(tick, 1000)

  let current = 0
  const drawPills = () => {
    pills.replaceChildren(
      ...exam.tasks.map((t, i) => {
        const pill = el('button', 'pill', `${i + 1}${t.flagged ? ' ⚑' : ''}`)
        pill.classList.toggle('current', i === current)
        pill.classList.toggle('flagged', t.flagged)
        pill.classList.toggle('failed', t.setup === 'failed')
        pill.onclick = () => open(i)
        return pill
      }),
    )
  }
  const open = (i: number) => {
    current = i
    drawPills()
    showQuestion(question, exam.tasks[i], catalog.get(exam.tasks[i].id)!, i + 1, drawPills)
  }
  open(0)

  end.onclick = async () => {
    if (!confirm('End the exam and score every task?')) return
    end.disabled = true
    end.textContent = 'Scoring…'
    try {
      showResults(pane, await endExam(), catalog)
    } catch (err) {
      end.disabled = false
      end.textContent = 'End exam'
      question.prepend(el('p', 'error', `Cannot end the exam: ${message(err)}`))
    }
  }
}

function showQuestion(target: HTMLElement, entry: ExamTask, task: Task, number: number, onFlag: () => void) {
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
  head.append(el('h2', '', `${number}. ${task.title}`), flag)
  const body = el('div', 'question-text')
  target.replaceChildren(head, chips(task), hostLine(task))
  if (entry.setup === 'failed') target.append(el('p', 'status error', `Setup failed: ${entry.setupError}`))
  target.append(body)
  taskQuestion(task.id).then(
    (text) => markdown(body, text),
    (err) => body.append(el('p', 'error', `Cannot load the question: ${message(err)}`)),
  )
}

function showResults(pane: HTMLElement, state: ExamState, catalog: Map<string, Task>) {
  clearInterval(ticker)
  const score = state.score!
  const verdict = el('p', `verdict ${score.passed ? 'pass' : 'fail'}`, `${score.percent}% · ${score.passed ? 'PASS' : 'FAIL'}`)
  const rows = el('div', 'results')
  state.exam.tasks.forEach((entry, i) => {
    const task = catalog.get(entry.id)!
    const r = entry.result!
    const row = el('details', 'result-row')
    const summary = el('summary')
    summary.append(el('span', '', `${i + 1}. ${task.title}`), el('span', 'points', `${r.earned} / ${r.total}`))
    row.append(summary)
    if (entry.checkError) row.append(el('p', 'error', `Check failed: ${entry.checkError}`))
    row.append(checkList(r))
    const solution = el('section', 'question-text solution')
    row.append(solution)
    row.ontoggle = () => {
      if (!row.open || solution.hasChildNodes()) return
      solution.append(el('h3', '', 'Solution'))
      taskSolution(task.id).then(
        (text) => markdown(solution, text),
        (err) => solution.append(el('p', 'error', `Cannot load the solution: ${message(err)}`)),
      )
    }
    rows.append(row)
  })
  const back = el('button', 'action primary', 'Back to practice')
  back.onclick = async () => {
    back.disabled = true
    try {
      await discardExam()
      showTasks(pane)
    } catch (err) {
      back.disabled = false
      pane.append(el('p', 'error', `Cannot leave the exam: ${message(err)}`))
    }
  }
  pane.replaceChildren(el('h2', '', 'Exam results'), verdict, el('p', 'muted', 'The CKA pass mark is 66%, weighted by task.'), rows, back)
}

function clock(ms: number): string {
  const s = Math.floor(ms / 1000)
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${h}:${pad(m)}:${pad(s % 60)}`
}
