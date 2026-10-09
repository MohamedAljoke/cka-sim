export type Task = {
  id: string
  title: string
  domain: string
  topics: string[]
  weight: number
  host: string
}

export async function listTasks(): Promise<Task[]> {
  return (await call('GET', '/api/tasks')).json()
}

export async function taskQuestion(id: string): Promise<string> {
  const reply = await (await call('GET', `/api/tasks/${encodeURIComponent(id)}/question`)).json()
  return reply.question
}

export async function startTask(id: string): Promise<void> {
  await call('POST', `/api/tasks/${encodeURIComponent(id)}/start`)
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

export async function checkTask(id: string): Promise<Result> {
  return (await call('POST', `/api/tasks/${encodeURIComponent(id)}/check`)).json()
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

export type Exam = {
  tasks: ExamTask[]
  minutes: number
  started?: string
  deadline?: string
  ended?: string
}

export type ExamState = {
  exam: Exam
  now: string
  score?: { percent: number; passed: boolean }
}

export async function getExam(): Promise<ExamState | null> {
  const res = await fetch('/api/exam')
  if (res.status === 404) return null
  if (!res.ok) throw await failure(res, 'GET', '/api/exam')
  return res.json()
}

export async function beginExam(count: number, minutes: number): Promise<ExamState> {
  return (await call('POST', '/api/exam', { count, minutes })).json()
}

export async function flagTask(id: string, flagged: boolean): Promise<void> {
  await call(flagged ? 'PUT' : 'DELETE', `/api/exam/flags/${encodeURIComponent(id)}`)
}

export async function endExam(): Promise<ExamState> {
  return (await call('POST', '/api/exam/end')).json()
}

export async function discardExam(): Promise<void> {
  await call('DELETE', '/api/exam')
}

async function call(method: string, path: string, body?: unknown): Promise<Response> {
  const init: RequestInit = { method }
  if (body !== undefined) {
    init.headers = { 'Content-Type': 'application/json' }
    init.body = JSON.stringify(body)
  }
  const res = await fetch(path, init)
  if (!res.ok) throw await failure(res, method, path)
  return res
}

async function failure(res: Response, method: string, path: string): Promise<Error> {
  return new Error((await res.text()).trim() || `${method} ${path}: ${res.status}`)
}
