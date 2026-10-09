export type Task = {
  id: string
  title: string
  domain: string
  topics: string[]
  weight: number
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

async function call(method: string, path: string): Promise<Response> {
  const res = await fetch(path, { method })
  if (!res.ok) throw new Error((await res.text()).trim() || `${method} ${path}: ${res.status}`)
  return res
}
