export type Task = {
  id: string
  title: string
  domain: string
  domainTitle: string
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

export type Exam = {
  tasks: ExamTask[]
  minutes: number
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
  const res = await fetch('/api/exam')
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

async function call(method: string, path: string, body?: unknown, signal?: AbortSignal): Promise<Response> {
  const init: RequestInit = { method, signal }
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
