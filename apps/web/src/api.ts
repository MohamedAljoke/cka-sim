export type Task = {
  id: string
  title: string
  domain: string
  domainTitle: string
  topics: string[]
  weight: number
  host: string
  order?: 'last'
}

export type TaskDetail = Task & {
  question: string
}

export type Filter = {
  domains: string[]
  topics: string[]
  search: string
}

export const noFilter: Filter = { domains: [], topics: [], search: '' }

export function filterQuery(filter: Filter): string {
  const params = new URLSearchParams()
  for (const domain of filter.domains) params.append('domain', domain)
  for (const topic of filter.topics) params.append('topic', topic)
  if (filter.search.trim()) params.set('q', filter.search.trim())
  return params.toString()
}

export async function listTasks(filter: Filter, signal?: AbortSignal): Promise<Task[]> {
  const query = filterQuery(filter)
  return (await call('GET', query ? `/api/tasks?${query}` : '/api/tasks', signal)).json()
}

export async function getTask(id: string, signal?: AbortSignal): Promise<TaskDetail> {
  return (await call('GET', `/api/tasks/${encodeURIComponent(id)}`, signal)).json()
}

export async function getSolution(id: string, signal?: AbortSignal): Promise<string> {
  const reply: { solution: string } = await (await call('GET', `/api/tasks/${encodeURIComponent(id)}/solution`, signal)).json()
  return reply.solution
}

export type Setup = {
  passed: boolean
  exitCode: number
  output?: string
}

export async function startTask(id: string, signal?: AbortSignal): Promise<Setup> {
  return (await call('POST', `/api/tasks/${encodeURIComponent(id)}/start`, signal)).json()
}

// Sent as the page leaves the task, so it must outlive the page: no signal, and keepalive.
export function tidyTask(id: string): void {
  fetch(`/api/tasks/${encodeURIComponent(id)}/tidy`, { method: 'POST', keepalive: true }).catch(() => {})
}

export type Grade = {
  earned: number
  total: number
  checks: { passed: boolean; points: number; description: string }[]
  exitCode?: number
  output?: string
}

export async function checkTask(id: string, signal?: AbortSignal): Promise<Grade> {
  return (await call('POST', `/api/tasks/${encodeURIComponent(id)}/check`, signal)).json()
}

export type LabState = 'none' | 'starting' | 'ready' | 'stopping' | 'failed'

export type Lab = {
  state: LabState
  step?: string
  startedAt?: string
  readyAt?: string
  error?: string
}

export async function getLab(signal?: AbortSignal): Promise<Lab> {
  return (await call('GET', '/api/lab', signal)).json()
}

export async function startLab(signal?: AbortSignal): Promise<Lab> {
  return (await call('POST', '/api/lab', signal)).json()
}

export async function stopLab(signal?: AbortSignal): Promise<Lab> {
  return (await call('DELETE', '/api/lab', signal)).json()
}

export type SetupState = 'preparing' | 'ready' | 'failed'

export type ExamTask = {
  id: string
  weight: number
  setup: SetupState
  setupError?: string
  flagged: boolean
  grade?: Grade
  checkError?: string
}

export type Exam = {
  tasks: ExamTask[]
  minutes: number
  started?: string
  deadline?: string
  ended?: string
  scored?: string
  now: string
  score?: { percent: number; passed: boolean; pass: number }
}

export async function getExam(signal?: AbortSignal): Promise<Exam | null> {
  const res = await fetch('/api/exam', { signal })
  if (res.status === 404) return null
  if (!res.ok) throw await failure(res, 'GET', '/api/exam')
  return res.json()
}

export async function beginExam(count: number, minutes: number, signal?: AbortSignal): Promise<Exam> {
  return (await call('POST', '/api/exam', signal, { count, minutes })).json()
}

export async function endExam(signal?: AbortSignal): Promise<Exam> {
  return (await call('POST', '/api/exam/end', signal)).json()
}

export async function discardExam(signal?: AbortSignal): Promise<void> {
  await call('DELETE', '/api/exam', signal)
}

export async function flagTask(id: string, flagged: boolean, signal?: AbortSignal): Promise<void> {
  await call(flagged ? 'PUT' : 'DELETE', `/api/exam/tasks/${encodeURIComponent(id)}/flag`, signal)
}

export type ChatMessage = { role: 'student' | 'tutor'; text: string }

export type ChatEvent = { type: 'tool'; name: string } | { type: 'text'; text: string }

// The answer comes as server-sent events over this POST: "tool" and "text" as the tutor works, then "done", or
// "error" once it has begun. EventSource can't send a body, so the stream is read by hand.
export async function chat(taskId: string, messages: ChatMessage[], signal: AbortSignal, onEvent: (event: ChatEvent) => void): Promise<void> {
  const res = await call('POST', '/api/assistant/chat', signal, { taskId, messages })
  const reader = res.body!.pipeThrough(new TextDecoderStream()).getReader()
  let buffer = ''
  for (;;) {
    const { value, done } = await reader.read()
    if (done) throw new Error('the answer stopped before it was complete')
    buffer += value
    for (let end = buffer.indexOf('\n\n'); end >= 0; end = buffer.indexOf('\n\n')) {
      const block = buffer.slice(0, end)
      buffer = buffer.slice(end + 2)
      const event = /^event: (.*)$/m.exec(block)?.[1]
      const data = JSON.parse(/^data: (.*)$/m.exec(block)?.[1] ?? '{}')
      if (event === 'done') return
      if (event === 'error') throw new Error(data.detail)
      if (event === 'tool') onEvent({ type: 'tool', name: data.name })
      if (event === 'text') onEvent({ type: 'text', text: data.text })
    }
  }
}

// Null when the site gives each student a token and this browser has none.
export async function getStudent(): Promise<string | null> {
  const res = await fetch('/api/student')
  if (res.status === 401) return null
  if (!res.ok) throw await failure(res, 'GET', '/api/student')
  const student: { name: string } = await res.json()
  return student.name
}

async function call(method: string, path: string, signal?: AbortSignal, body?: unknown): Promise<Response> {
  const res = await fetch(path, body === undefined ? { method, signal } : {
    method,
    signal,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw await failure(res, method, path)
  return res
}

// The Go server answers with a problem document; anything on the way (Vite, Caddy) answers with text.
async function failure(res: Response, method: string, path: string): Promise<Error> {
  if (res.headers.get('Content-Type')?.startsWith('application/problem+json')) {
    const problem: { title?: string; detail?: string } = await res.json().catch(() => ({}))
    return new Error(problem.detail || problem.title || `${method} ${path}: ${res.status}`)
  }
  return new Error((await res.text()).trim() || `${method} ${path}: ${res.status}`)
}
