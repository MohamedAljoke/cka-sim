import { marked } from 'marked'
import type { Task } from './api'

// Questions come from the bundled catalog, not from users, so their markdown renders as is.
export function markdown(target: HTMLElement, text: string) {
  target.insertAdjacentHTML('beforeend', marked.parse(text, { async: false }))
}

export function chips(task: Task): HTMLElement {
  const row = el('span', 'chips')
  row.append(el('span', 'chip domain', task.domainTitle), ...task.topics.map((topic) => el('span', 'chip', topic)))
  return row
}

export function el<K extends keyof HTMLElementTagNameMap>(tag: K, className = '', text = '') {
  const node = document.createElement(tag)
  node.className = className
  node.textContent = text
  return node
}

export function message(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}
