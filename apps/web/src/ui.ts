import { marked } from 'marked'
import type { Task } from './api'

// Questions and solutions come from the bundled catalog, not from users, so their markdown renders as is.
export function markdown(target: HTMLElement, text: string) {
  target.insertAdjacentHTML('beforeend', marked.parse(text, { async: false }))
}

export function chips(task: Task): HTMLElement {
  const row = el('span', 'chips')
  row.append(el('span', 'chip domain', task.domainTitle), ...task.topics.map((topic) => el('span', 'chip', topic)))
  return row
}

export function copyable(text: string): HTMLElement {
  const label = el('span', 'copy-label', 'Copy')
  const button = el('button', 'copyable')
  button.type = 'button'
  button.title = 'Copy to clipboard'
  button.append(el('code', '', text), label)
  let reset: ReturnType<typeof setTimeout> | undefined
  button.onclick = async () => {
    clearTimeout(reset)
    try {
      await navigator.clipboard.writeText(text)
      label.textContent = 'Copied'
    } catch {
      label.textContent = 'Copy failed'
    }
    reset = setTimeout(() => (label.textContent = 'Copy'), 1500)
  }
  return button
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
