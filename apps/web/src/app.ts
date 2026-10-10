import type { Task } from './api'

// App is what every view gets to move between views and shape the page around itself.
export type App = {
  pane: HTMLElement
  home(): void
  practice(task: Task): void
  exam(): void
  // split shows the terminal beside the pane; views without a cluster task use the full width.
  split(on: boolean): void
  // labChanged tells the topbar the server may have started the lab.
  labChanged(): void
  // onLeave runs once when the next view replaces this one.
  onLeave(cleanup: () => void): void
}
