import { beginExam, discardExam, endExam, flagTask, getExam, getTask, listTasks, noFilter, type Exam, type ExamTask, type Task, type TaskDetail } from './api'
import { checkItem } from './check'
import type { LabStatus } from './lab'
import { taskPath } from './routes'
import { shellPanel } from './shell'
import { chips, copyable, el, markdown, message } from './ui'

const pollEvery = 1000

type Catalog = Map<string, Task>

export async function showExam(pane: HTMLElement, lab: LabStatus, signal: AbortSignal) {
  document.title = 'Exam · cka-sim'
  phase('loading')
  pane.replaceChildren(note(el('p', 'muted', 'Loading the exam…')))
  let catalog: Catalog
  let exam: Exam | null
  try {
    const [all, current] = await Promise.all([listTasks(noFilter, signal), getExam(signal)])
    catalog = new Map(all.map((task) => [task.id, task]))
    exam = current
  } catch (err) {
    if (!signal.aborted) pane.replaceChildren(note(el('p', 'error', `Cannot load the exam: ${message(err)}`)))
    return
  }
  if (!signal.aborted) render(pane, exam, catalog, lab, signal)
}

function render(pane: HTMLElement, exam: Exam | null, catalog: Catalog, lab: LabStatus, signal: AbortSignal) {
  if (!exam) return showForm(pane, catalog, lab, signal)
  if (!exam.started) return showSetup(pane, exam, catalog, lab, signal)
  if (!exam.ended) return showRunning(pane, exam, catalog, lab, signal)
  return showResults(pane, exam, catalog, lab, signal)
}

// Polls until the exam leaves the phase it is in, then draws the next one.
function follow(pane: HTMLElement, catalog: Catalog, lab: LabStatus, signal: AbortSignal, still: (exam: Exam) => boolean, redraw: (exam: Exam) => void) {
  const timer = setTimeout(async () => {
    try {
      const exam = await getExam(signal)
      if (signal.aborted) return
      if (exam && still(exam)) {
        redraw(exam)
        follow(pane, catalog, lab, signal, still, redraw)
      } else render(pane, exam, catalog, lab, signal)
    } catch {
      if (!signal.aborted) follow(pane, catalog, lab, signal, still, redraw)
    }
  }, pollEvery)
  signal.addEventListener('abort', () => clearTimeout(timer))
}

function showForm(pane: HTMLElement, catalog: Catalog, lab: LabStatus, signal: AbortSignal) {
  phase('form')
  const count = numberField('Questions', 16, 1, catalog.size)
  const minutes = numberField('Minutes', 120, 1, 240)
  const start = el('button', 'action primary', 'Start exam')
  start.type = 'submit'
  const error = el('p', 'error')
  error.hidden = true

  const form = el('form', 'exam-form')
  form.append(count.label, minutes.label, start)
  form.onsubmit = async (event) => {
    event.preventDefault()
    start.disabled = true
    error.hidden = true
    try {
      const exam = await beginExam(count.input.valueAsNumber, minutes.input.valueAsNumber, signal)
      if (!signal.aborted) render(pane, exam, catalog, lab, signal)
    } catch (err) {
      if (signal.aborted) return
      error.textContent = `Cannot start the exam: ${message(err)}`
      error.hidden = false
      start.disabled = false
    }
  }

  const card = el('section', 'card')
  card.append(
    el('h2', '', 'Exam'),
    el('p', 'muted', 'Questions are drawn by the CKA domain weights and all set up on the lab first. The clock starts once they are ready. You pass at 66%.'),
    form,
    error,
  )
  pane.replaceChildren(card)
}

function showSetup(pane: HTMLElement, exam: Exam, catalog: Catalog, lab: LabStatus, signal: AbortSignal) {
  phase('setup')
  const status = el('p', 'muted')
  const list = el('ol', 'exam-setup')
  const draw = (exam: Exam) => {
    const done = exam.tasks.filter((t) => t.setup !== 'preparing').length
    status.textContent = `Setting up the cluster: ${done} of ${exam.tasks.length} questions ready.`
    list.replaceChildren(
      ...exam.tasks.map((t) => {
        const item = el('li', 'exam-setup-item')
        item.dataset.state = t.setup
        item.append(el('span', 'lab-dot'), el('span', '', titleOf(catalog, t.id)), el('span', 'muted', setupLabel[t.setup]))
        return item
      }),
    )
  }
  draw(exam)
  const card = el('section', 'card')
  card.setAttribute('aria-live', 'polite')
  card.append(el('h2', '', 'Preparing your exam'), status, list)
  pane.replaceChildren(card)
  follow(pane, catalog, lab, signal, (e) => !e.started, draw)
}

const setupLabel = { preparing: 'Setting up…', ready: 'Ready', failed: 'Setup failed' }

function showRunning(pane: HTMLElement, exam: Exam, catalog: Catalog, lab: LabStatus, signal: AbortSignal) {
  phase('running')
  const offset = Date.parse(exam.now) - Date.now()
  const deadline = Date.parse(exam.deadline!)
  let current = 0

  const timer = el('span', 'exam-timer')
  timer.setAttribute('role', 'timer')
  const end = el('button', 'action primary', 'End exam')
  end.type = 'button'
  const bar = el('div', 'exam-bar')
  bar.append(el('span', 'exam-label', 'Exam'), timer, end)

  const pills = el('nav', 'exam-pills')
  pills.setAttribute('aria-label', 'Questions')
  const question = el('div', 'exam-question')
  const brief = el('article', 'task-brief')
  brief.append(bar, pills, question)

  const shell = shellPanel(signal, lab.refresh)
  lab.watch(signal, shell.follow)
  const workspace = el('section', 'task-lab')
  workspace.setAttribute('aria-label', 'Workspace')
  workspace.append(shell.panel)

  const layout = el('div', 'task-layout')
  layout.append(brief, workspace)
  pane.replaceChildren(layout)

  const finish = async () => {
    end.disabled = true
    end.textContent = 'Ending…'
    try {
      const ended = await endExam(signal)
      if (!signal.aborted) render(pane, ended, catalog, lab, signal)
    } catch (err) {
      if (signal.aborted) return
      end.disabled = false
      end.textContent = 'End exam'
      question.prepend(el('p', 'error', `Cannot end the exam: ${message(err)}`))
    }
  }
  end.onclick = () => {
    if (confirm('End the exam now? Every question is graded as it stands.')) finish()
  }

  const drawPills = () => {
    pills.replaceChildren(
      ...exam.tasks.map((t, i) => {
        const pill = el('button', 'exam-pill', String(i + 1))
        pill.type = 'button'
        pill.title = titleOf(catalog, t.id)
        if (i === current) pill.setAttribute('aria-current', 'true')
        if (t.flagged) pill.dataset.flagged = 'true'
        if (t.setup === 'failed') pill.dataset.failed = 'true'
        pill.setAttribute('aria-label', `Question ${i + 1}${t.flagged ? ', flagged' : ''}`)
        pill.onclick = () => select(i)
        return pill
      }),
    )
  }

  const details = new Map<string, Promise<TaskDetail>>()
  const select = async (i: number) => {
    current = i
    drawPills()
    const entry = exam.tasks[i]
    const head = el('div', 'exam-question-head')
    const flag = el('button', 'action', entry.flagged ? 'Unflag' : 'Flag for review')
    flag.type = 'button'
    flag.setAttribute('aria-pressed', String(entry.flagged))
    flag.onclick = async () => {
      flag.disabled = true
      try {
        await flagTask(entry.id, !entry.flagged, signal)
        entry.flagged = !entry.flagged
        if (!signal.aborted) select(i)
      } catch (err) {
        if (!signal.aborted) question.prepend(el('p', 'error', `Cannot flag the question: ${message(err)}`))
      } finally {
        flag.disabled = false
      }
    }
    head.append(el('h1', '', `${i + 1}. ${titleOf(catalog, entry.id)}`), flag)
    question.replaceChildren(head, el('p', 'muted', 'Loading the question…'))

    if (!details.has(entry.id)) details.set(entry.id, getTask(entry.id, signal))
    let task: TaskDetail
    try {
      task = await details.get(entry.id)!
    } catch (err) {
      details.delete(entry.id)
      if (!signal.aborted && current === i) question.replaceChildren(head, el('p', 'error', `Cannot load the question: ${message(err)}`))
      return
    }
    if (signal.aborted || current !== i) return

    const meta = el('div', 'exam-meta')
    meta.append(chips(task), el('span', 'chip', `weight ${entry.weight}`))
    if (task.order === 'last') meta.append(el('span', 'chip', 'Affects the whole cluster'))
    const host = el('p', 'host')
    host.append(el('span', 'muted', 'Connect first: '), copyable(`ssh ${task.host}`))
    const body = el('div', 'markdown')
    markdown(body, task.question)
    question.replaceChildren(head, meta)
    if (entry.setup === 'failed') question.append(el('p', 'error', 'This question could not be set up, so it is left out of your score.'))
    question.append(host, body)
  }

  const tick = () => {
    const left = deadline - (Date.now() + offset)
    timer.textContent = clock(left)
    timer.dataset.low = String(left < 10 * 60 * 1000)
    if (left > 0) return
    clearInterval(ticking)
    follow(pane, catalog, lab, signal, (e) => !e.ended, () => {})
  }
  const ticking = setInterval(tick, 1000)
  signal.addEventListener('abort', () => clearInterval(ticking))
  tick()
  select(0)
}

function showResults(pane: HTMLElement, exam: Exam, catalog: Catalog, lab: LabStatus, signal: AbortSignal) {
  phase('results')
  if (!exam.score) {
    const card = el('section', 'card')
    card.setAttribute('aria-live', 'polite')
    card.append(el('h2', '', 'Grading your exam…'), el('p', 'muted', 'Every question is being checked on the cluster.'))
    pane.replaceChildren(card)
    follow(pane, catalog, lab, signal, (e) => !e.score, () => {})
    return
  }

  const { percent, passed, pass } = exam.score
  const verdict = el('p', `exam-verdict ${passed ? 'passed' : 'failed'}`)
  verdict.append(el('strong', '', `${percent}%`), el('span', '', passed ? ` · Passed (${pass}% needed)` : ` · Not passed (${pass}% needed)`))

  const rows = el('ol', 'exam-results')
  rows.append(...exam.tasks.map((t, i) => resultRow(t, i, catalog)))

  const again = el('button', 'action primary', 'New exam')
  again.type = 'button'
  again.onclick = async () => {
    again.disabled = true
    try {
      await discardExam(signal)
      if (!signal.aborted) render(pane, null, catalog, lab, signal)
    } catch (err) {
      if (signal.aborted) return
      again.disabled = false
      card.append(el('p', 'error', `Cannot start over: ${message(err)}`))
    }
  }

  const card = el('section', 'card')
  card.append(
    el('h2', '', 'Your result'),
    verdict,
    rows,
    el('p', 'muted', 'The cluster keeps what this exam changed. Restart the lab for a clean one before the next exam.'),
    again,
  )
  pane.replaceChildren(card)
}

function resultRow(t: ExamTask, i: number, catalog: Catalog): HTMLElement {
  const item = el('li', 'exam-result')
  const head = el('div', 'exam-result-head')
  const link = el('a', '', `${i + 1}. ${titleOf(catalog, t.id)}`)
  link.href = taskPath(t.id)
  head.append(link, el('span', 'muted', resultLabel(t)))
  if (t.flagged) head.append(el('span', 'chip', 'flagged'))
  item.dataset.state = t.setup === 'failed' ? 'skipped' : t.grade && t.grade.total > 0 && t.grade.earned === t.grade.total ? 'passed' : 'failed'
  item.append(head)

  if (t.grade?.checks.length) {
    const list = el('ul', 'check-list')
    list.append(...t.grade.checks.map(checkItem))
    item.append(list)
  }
  const why = t.setup === 'failed' ? t.setupError : t.checkError
  if (why) item.append(el('pre', 'setup-output', why))
  return item
}

function resultLabel(t: ExamTask): string {
  if (t.setup === 'failed') return 'setup failed · not scored'
  if (!t.grade || t.grade.total === 0) return `0 points · weight ${t.weight}`
  return `${t.grade.earned}/${t.grade.total} points · weight ${t.weight}`
}

function numberField(text: string, value: number, min: number, max: number) {
  const input = el('input', 'exam-number')
  Object.assign(input, { type: 'number', value: String(Math.min(value, max)), min: String(min), max: String(max), required: true })
  const label = el('label', 'exam-field')
  label.append(el('span', 'muted', text), input)
  return { label, input }
}

function titleOf(catalog: Catalog, id: string): string {
  return catalog.get(id)?.title ?? id
}

function clock(ms: number): string {
  const total = Math.max(0, Math.ceil(ms / 1000))
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  const mm = String(m).padStart(2, '0')
  const ss = String(s).padStart(2, '0')
  return h > 0 ? `${h}:${mm}:${ss}` : `${mm}:${ss}`
}

function phase(name: string) {
  document.body.dataset.phase = name
}

function note(...children: HTMLElement[]): HTMLElement {
  const box = el('div', 'task-note')
  box.append(...children)
  return box
}
