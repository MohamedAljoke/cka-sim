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

export async function listTasks(): Promise<Task[]> {
  return (await call('GET', '/api/tasks')).json()
}

export async function taskQuestion(id: string): Promise<string> {
  const reply = await (await call('GET', `/api/tasks/${encodeURIComponent(id)}/question`)).json()
  return reply.question
}

export async function startTask(id: string, signal?: AbortSignal): Promise<void> {
  await call('POST', `/api/tasks/${encodeURIComponent(id)}/start`, undefined, signal)
}

// tidyTask outlives the page (keepalive), so closing the tab still cleans up.
export function tidyTask(id: string): void {
  fetch(`/api/tasks/${encodeURIComponent(id)}/tidy`, { method: 'POST', keepalive: true }).catch(() => {})
}

export type Check = {
  passed: boolean
  points: number
  description: string
}

export type Result = {
  checks: Check[]
  earned: number
  total: number
}

export async function checkTask(id: string, signal?: AbortSignal): Promise<Result> {
  return (await call('POST', `/api/tasks/${encodeURIComponent(id)}/check`, undefined, signal)).json()
}

export async function taskSolution(id: string): Promise<string> {
  const reply = await (await call('GET', `/api/tasks/${encodeURIComponent(id)}/solution`)).json()
  return reply.explain
}

export type ExamTask = {
  id: string
  setup: 'preparing' | 'ready' | 'failed'
  setupError?: string
  flagged: boolean
  result?: Result
  checkError?: string
}

export type Stage = 'lab' | 'heal' | 'setup' | 'cluster'

export type Exam = {
  tasks: ExamTask[]
  minutes: number
  stage?: Stage
  started?: string
  deadline?: string
  ended?: string
  scored?: string
}

export type ExamState = {
  exam: Exam
  now: string
  score?: { percent: number; passed: boolean }
}

export async function getExam(): Promise<ExamState | null> {
  const res = await withSession('/api/exam', {})
  if (res.status === 404) return null
  if (!res.ok) throw await failure(res, 'GET', '/api/exam')
  return res.json()
}

export type ExamPlan = {
  count: number
  minutes: number
  domains: string[]
  topics: string[]
}

export async function beginExam(plan: ExamPlan): Promise<ExamState> {
  return (await call('POST', '/api/exam', plan)).json()
}

export async function flagTask(id: string, flagged: boolean): Promise<void> {
  await call(flagged ? 'PUT' : 'DELETE', `/api/exam/flags/${encodeURIComponent(id)}`)
}

export async function endExam(): Promise<ExamState> {
  return (await call('POST', '/api/exam/end')).json()
}

export type Lab = {
  state: 'none' | 'starting' | 'ready' | 'failed'
  provider: string
  started?: string
  ready?: string
  error?: string
  ended?: string
}

export type LabState = {
  lab: Lab
  now: string
}

export async function getLab(): Promise<LabState> {
  return (await call('GET', '/api/lab')).json()
}

export async function endLab(): Promise<void> {
  await call('DELETE', '/api/lab')
}

export async function startAssistant(provider: string): Promise<void> {
  const path = '/api/assistant'
  const res = await withSession(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ provider }),
  })
  if (res.status === 404) throw new Error('The assistant is not available on this server yet.')
  if (!res.ok) throw await failure(res, 'POST', path)
}

export type ChatMessage = {
  role: 'user' | 'model'
  text: string
}

export type ChatTurn = {
  text: string
  tools: string[]
}

export type ChatCheck = Result & { task: string }

export type ChatSource = { title: string; url: string }

export type ChatProgress = {
  text(delta: string): void
  tool(name: string): void
  check(result: ChatCheck): void
  warning(tool: string): void
  sources(found: ChatSource[]): void
}

type ChatEvent = {
  event: string
  data: { text?: string; name?: string; tool?: string; detail?: string; tools?: string[]; sources?: ChatSource[] } &
    Partial<ChatCheck>
}

export async function chatAssistant(
  provider: string,
  messages: ChatMessage[],
  signal: AbortSignal,
  on: ChatProgress,
): Promise<ChatTurn> {
  const res = await call('POST', '/api/assistant/chat', { provider, messages }, signal)
  const reader = res.body!.pipeThrough(new TextDecoderStream()).getReader()
  let text = ''
  let buffer = ''
  try {
    for (;;) {
      const { value, done } = await reader.read()
      if (done) throw new Error('the assistant stopped before finishing its answer; ask again')
      buffer += value
      for (let end = buffer.indexOf('\n\n'); end >= 0; end = buffer.indexOf('\n\n')) {
        const { event, data } = parseEvent(buffer.slice(0, end))
        buffer = buffer.slice(end + 2)
        if (event === 'text') {
          text += data.text ?? ''
          on.text(data.text ?? '')
        } else if (event === 'tool') on.tool(data.name ?? '')
        else if (event === 'check') on.check(data as ChatCheck)
        else if (event === 'warning') on.warning(data.tool ?? '')
        else if (event === 'sources') on.sources(data.sources ?? [])
        else if (event === 'error') throw new Error(data.detail)
        else if (event === 'done') return { text, tools: data.tools ?? [] }
      }
    }
  } finally {
    reader.cancel().catch(() => {})
  }
}

function parseEvent(block: string): ChatEvent {
  let event = ''
  let data = ''
  for (const line of block.split('\n')) {
    if (line.startsWith('event: ')) event = line.slice('event: '.length)
    else if (line.startsWith('data: ')) data += line.slice('data: '.length)
  }
  return { event, data: JSON.parse(data || '{}') }
}

export type User = {
  id: string
  email: string
}

// currentUser is null when nobody is logged in; the session cookie is HttpOnly, so only the server can tell.
export async function currentUser(): Promise<User | null> {
  let res = await fetch('/api/auth/me')
  if (res.status === 401) {
    if (!(await refreshSession())) return null
    res = await fetch('/api/auth/me')
  }
  if (res.status === 401) return null
  if (!res.ok) throw await failure(res, 'GET', '/api/auth/me')
  return res.json()
}

let refreshing: Promise<boolean> | undefined

// One refresh at a time, across tabs too: each tab sends the cookie the previous refresh left behind.
export function refreshSession(): Promise<boolean> {
  const renew = () => fetch('/api/auth/refresh', { method: 'POST' }).then((res) => res.ok, () => false)
  refreshing ??= (navigator.locks ? navigator.locks.request('cka-sim-refresh', renew) : renew()).finally(() => {
    refreshing = undefined
  })
  return refreshing
}

export async function logIn(email: string, password: string): Promise<User> {
  return (await call('POST', '/api/auth/login', { email, password })).json()
}

export async function logOut(): Promise<void> {
  await call('POST', '/api/auth/logout')
}

async function call(method: string, path: string, body?: unknown, signal?: AbortSignal): Promise<Response> {
  const init: RequestInit = { method, signal }
  if (body !== undefined) {
    init.headers = { 'Content-Type': 'application/json' }
    init.body = JSON.stringify(body)
  }
  const res = path.startsWith('/api/auth/') ? await fetch(path, init) : await withSession(path, init)
  if (!res.ok) throw await failure(res, method, path)
  return res
}

async function withSession(path: string, init: RequestInit): Promise<Response> {
  let res = await fetch(path, init)
  if (res.status === 401 && (await refreshSession())) res = await fetch(path, init)
  if (res.status === 401) return dropSessionAndReload()
  return res
}

async function dropSessionAndReload(): Promise<never> {
  await logOut().catch(() => {})
  location.reload()
  return new Promise(() => {})
}

// The Go servers answer with a problem document; anything on the way (Vite, Caddy) answers with text.
async function failure(res: Response, method: string, path: string): Promise<Error> {
  if (res.headers.get('Content-Type')?.startsWith('application/problem+json')) {
    const problem: { title?: string; detail?: string } = await res.json().catch(() => ({}))
    return new Error(problem.detail || problem.title || `${method} ${path}: ${res.status}`)
  }
  return new Error((await res.text()).trim() || `${method} ${path}: ${res.status}`)
}
