import { getCurrentTask, getExam } from './api'
import { examRunning } from './exam'
import type { LabStatus } from './lab'
import { examPath, parseRoute, taskPath } from './routes'
import { el } from './ui'

const pollEvery = 3000
const noteFor = 6000

export type Move = (hash: string, alreadyTidied: boolean) => void

export function followSession(lab: LabStatus, move: Move) {
  const note = el('p', 'follow-note')
  note.setAttribute('role', 'status')
  note.hidden = true
  document.body.append(note)
  let hideNote: ReturnType<typeof setTimeout> | undefined

  let ready = false
  let busy = false
  let timer: ReturnType<typeof setTimeout> | undefined
  let seenTask: string | null | undefined
  let seenExam = false

  const go = (hash: string, says: string, alreadyTidied: boolean) => {
    if (location.hash === hash) return
    note.textContent = says
    note.hidden = false
    clearTimeout(hideNote)
    hideNote = setTimeout(() => (note.hidden = true), noteFor)
    move(hash, alreadyTidied)
  }

  const look = async () => {
    if (examRunning(await getExam())) {
      if (!seenExam && parseRoute(location.hash).page !== 'exam') {
        go(examPath, 'A practice exam has begun, so the page moved to it.', false)
      }
      seenExam = true
      return
    }
    seenExam = false

    const current = await getCurrentTask()
    const id = current?.id ?? null
    const route = parseRoute(location.hash)
    const showing = route.page === 'task' && route.id === id
    if (seenTask !== undefined && id !== seenTask && current && !showing) {
      go(taskPath(current.id), `“${current.title}” was set up on your lab, so the page moved to it.`, true)
    }
    seenTask = id
  }

  const schedule = () => {
    clearTimeout(timer)
    if (ready && !document.hidden) timer = setTimeout(tick, pollEvery)
  }
  const tick = async () => {
    if (busy) return
    busy = true
    try {
      await look()
    } catch {
    } finally {
      busy = false
      schedule()
    }
  }

  lab.watch(new AbortController().signal, (state) => {
    const was = ready
    ready = state.state === 'ready'
    if (ready && !was) tick()
    if (!ready) clearTimeout(timer)
  })
  document.addEventListener('visibilitychange', () => {
    if (document.hidden) clearTimeout(timer)
    else if (ready) tick()
  })
}
