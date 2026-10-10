import { beginExam, getExam, listTasks, type ExamState, type Task } from './api'
import type { App } from './app'
import { multiSelect, type Option } from './multiselect'
import { chips, el, message } from './ui'

// The CKA gives 120 minutes for about 16 tasks.
const minutesPerTask = 7.5

// showHome needs no lab: the catalog and the exam settings come from the server alone.
export async function showHome(app: App) {
  const { pane } = app
  app.split(false)
  pane.replaceChildren(el('p', 'muted', 'Loading tasks…'))
  let tasks: Task[]
  let last: ExamState | null
  try {
    ;[tasks, last] = await Promise.all([listTasks(), getExam()])
  } catch (err) {
    pane.replaceChildren(el('p', 'error', `Cannot load tasks: ${message(err)}`))
    return
  }

  // Mirrors tasks.Filter on the server. With every topic chosen, a task without topics still counts.
  const matching = () => tasks.filter((t) => domains.has(t.domain) && (topics.all() || t.topics.some((topic) => topics.has(topic))))
  const changed = () => {
    const found = matching()
    summary.textContent = `${found.length} of ${tasks.length} tasks`
    exam.update(found.length)
    drawList(found)
  }
  const domains = multiSelect('domains', options(tasks, (t) => [t.domain], (_, t) => t.domainTitle), changed)
  const topics = multiSelect('topics', options(tasks, (t) => t.topics, (topic) => topic), changed)
  app.onLeave(() => {
    domains.dispose()
    topics.dispose()
  })
  const summary = el('span', 'filter-summary')
  const filters = el('section', 'filter-bar')
  filters.append(domains.element, topics.element, summary)

  const exam = examCard(app, () => ({ domains: domains.narrowed(), topics: topics.narrowed() }))
  const list = el('ol', 'task-list')
  const drawList = (found: Task[]) => {
    if (found.length === 0) {
      list.replaceChildren(el('li', 'muted', 'No task matches. Pick at least one domain and one topic.'))
      return
    }
    list.replaceChildren(
      ...found.map((task) => {
        const button = el('button', 'task')
        button.append(el('span', 'task-title', task.title), chips(task))
        button.onclick = () => app.practice(task)
        const item = el('li')
        item.append(button)
        return item
      }),
    )
  }
  const catalog = el('section', 'card')
  catalog.append(el('h2', '', 'Practise a task'), el('p', 'muted', 'Work on one task with Check and Solution any time.'), list)

  const home = el('div', 'home')
  home.append(...lastExam(app, last), filters, exam.card, catalog)
  pane.replaceChildren(home)
  changed()
}

// options lists every value the catalog has, labelled, with how many tasks carry it.
function options(tasks: Task[], values: (t: Task) => string[], label: (v: string, t: Task) => string): Option[] {
  const found = new Map<string, Option>()
  for (const t of tasks) {
    for (const value of values(t)) {
      const option = found.get(value) ?? { value, label: label(value, t), count: 0 }
      option.count++
      found.set(value, option)
    }
  }
  return [...found.values()].sort((a, b) => a.label.localeCompare(b.label))
}

function lastExam(app: App, last: ExamState | null): HTMLElement[] {
  if (!last?.exam.ended) return []
  const line = el('section', 'card last-exam')
  const text = last.score ? `Last exam: ${last.score.percent}% · ${last.score.passed ? 'PASS' : 'FAIL'}` : 'Last exam: scoring…'
  const view = el('button', 'action', 'View results')
  view.onclick = () => app.exam()
  line.append(el('span', '', text), view)
  return [line]
}

function examCard(app: App, filter: () => { domains: string[]; topics: string[] }) {
  const card = el('section', 'card')
  const lead = el('p', 'muted')
  const form = el('form', 'exam-form')
  const count = number('Tasks')
  const minutes = number('Minutes')
  minutes.input.max = '240'
  const start = el('button', 'action primary', 'Start exam')
  start.type = 'submit'
  const error = el('p', 'error')
  form.append(count.label, minutes.label, start, error)
  card.append(el('h2', '', 'Mock exam'), lead, form)

  let minutesTouched = false
  minutes.input.oninput = () => (minutesTouched = true)
  const suggestMinutes = () => {
    if (!minutesTouched) minutes.input.value = String(Math.max(5, Math.round((count.input.valueAsNumber || 1) * minutesPerTask)))
  }
  count.input.oninput = suggestMinutes
  const update = (available: number) => {
    const { domains, topics } = filter()
    const scope = domains.length === 0 && topics.length === 0 ? 'the whole catalog, in the CKA domain mix' : 'the tasks you filtered'
    lead.textContent = `Drawn from ${scope}, timed, and scored at the end against the 66% pass mark.`
    count.input.max = String(Math.max(1, available))
    count.input.value = String(Math.min(16, available))
    start.disabled = available === 0
    suggestMinutes()
  }

  form.onsubmit = async (event) => {
    event.preventDefault()
    start.disabled = true
    error.textContent = ''
    try {
      await beginExam({ count: count.input.valueAsNumber, minutes: minutes.input.valueAsNumber, ...filter() })
      app.labChanged()
      app.exam()
    } catch (err) {
      error.textContent = `Cannot start the exam: ${message(err)}`
      start.disabled = false
    }
  }
  return { card, update }
}

function number(text: string) {
  const label = el('label', '', text)
  const input = el('input')
  Object.assign(input, { type: 'number', min: '1', required: true })
  label.append(input)
  return { label, input }
}
