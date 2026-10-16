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

async function call(method: string, path: string): Promise<Response> {
  const res = await fetch(path, { method })
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
