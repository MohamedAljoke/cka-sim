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
}

export const noFilter: Filter = { domains: [], topics: [] }

export function filterQuery(filter: Filter): string {
  const params = new URLSearchParams()
  for (const domain of filter.domains) params.append('domain', domain)
  for (const topic of filter.topics) params.append('topic', topic)
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

async function call(method: string, path: string, signal?: AbortSignal): Promise<Response> {
  const res = await fetch(path, { method, signal })
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
