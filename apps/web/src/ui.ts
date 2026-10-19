import { marked, type MarkedToken, type Token } from 'marked'
import type { Task } from './api'

// Questions and solutions come from the bundled catalog, not from users, so their markdown renders as is.
export function markdown(target: HTMLElement, text: string) {
  target.insertAdjacentHTML('beforeend', marked.parse(text, { async: false }))
}

// Untrusted markdown (the tutor's answers, from a model that reads the cluster) is built node by node: any HTML
// in it stays text, and only http(s) links become links.
export function untrustedMarkdown(target: HTMLElement, text: string) {
  target.replaceChildren(...blocks(marked.lexer(text)))
}

function blocks(tokens: Token[] = []): Node[] {
  return tokens.flatMap((token) => block(token as MarkedToken))
}

function block(token: MarkedToken): Node[] {
  switch (token.type) {
    case 'paragraph':
      return [wrap('p', inline(token.tokens))]
    case 'heading':
      return [wrap('h4', inline(token.tokens))]
    case 'code':
      return [wrap('pre', [el('code', '', token.text)])]
    case 'blockquote':
      return [wrap('blockquote', blocks(token.tokens))]
    case 'list': {
      const list = el(token.ordered ? 'ol' : 'ul')
      if (token.ordered && token.start !== '') (list as HTMLOListElement).start = token.start
      list.append(...token.items.map((item) => wrap('li', blocks(item.tokens))))
      return [list]
    }
    case 'table': {
      const table = el('table')
      table.append(wrap('tr', token.header.map((cell) => wrap('th', inline(cell.tokens)))))
      table.append(...token.rows.map((row) => wrap('tr', row.map((cell) => wrap('td', inline(cell.tokens))))))
      return [wrap('div', [table], 'table-scroll')]
    }
    case 'text':
      return token.tokens ? inline(token.tokens) : [document.createTextNode(token.text)]
    case 'hr':
      return [el('hr')]
    case 'space':
    case 'def':
      return []
    default:
      return [document.createTextNode(token.raw)]
  }
}

function inline(tokens: Token[] = []): Node[] {
  return tokens.flatMap((generic): Node[] => {
    const token = generic as MarkedToken
    switch (token.type) {
      case 'strong':
      case 'em':
      case 'del':
        return [wrap(token.type, inline(token.tokens))]
      case 'codespan':
        return [el('code', '', token.text)]
      case 'br':
        return [el('br')]
      case 'link':
        return [link(token.href, inline(token.tokens))]
      case 'text':
        return token.tokens ? inline(token.tokens) : [document.createTextNode(token.text)]
      case 'escape':
      case 'image':
        return [document.createTextNode(token.text)]
      default:
        return [document.createTextNode(token.raw)]
    }
  })
}

function link(href: string, children: Node[]): Node {
  if (!/^https?:\/\//i.test(href)) return wrap('span', children)
  const a = wrap('a', children)
  a.href = href
  a.target = '_blank'
  a.rel = 'noopener noreferrer'
  return a
}

function wrap<K extends keyof HTMLElementTagNameMap>(tag: K, children: Node[], className = '') {
  const node = el(tag, className)
  node.append(...children)
  return node
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

type Resize = {
  label: string
  axis: 'x' | 'y'
  // grow is 1 when moving the pointer right or down makes the target bigger, -1 when it makes it smaller.
  grow: 1 | -1
  key: string
  size: () => number
  limits: () => [number, number]
  apply: (px: number) => void
}

// resizer is a thin bar that sizes something by dragging or with the arrow keys; the size is kept per browser.
export function resizer({ label, axis, grow, key, size, limits, apply }: Resize): HTMLElement {
  const bar = el('div', `resizer resizer-${axis}`)
  bar.tabIndex = 0
  bar.setAttribute('role', 'separator')
  bar.setAttribute('aria-orientation', axis === 'x' ? 'vertical' : 'horizontal')
  bar.setAttribute('aria-label', label)

  const resize = (px: number) => {
    const [min, max] = limits()
    const kept = Math.round(Math.max(min, Math.min(px, max)))
    apply(kept)
    try {
      localStorage.setItem(key, String(kept))
    } catch {}
  }
  try {
    const saved = Number(localStorage.getItem(key))
    if (saved) apply(saved)
  } catch {}

  const at = (event: PointerEvent) => (axis === 'x' ? event.clientX : event.clientY)
  bar.onpointerdown = (down) => {
    down.preventDefault()
    bar.setPointerCapture(down.pointerId)
    const start = size()
    bar.onpointermove = (move) => resize(start + (at(move) - at(down)) * grow)
    bar.onpointerup = () => (bar.onpointermove = null)
  }
  const steps: Record<string, number> = axis === 'x' ? { ArrowLeft: -24, ArrowRight: 24 } : { ArrowUp: -24, ArrowDown: 24 }
  bar.onkeydown = (event) => {
    if (!steps[event.key]) return
    event.preventDefault()
    resize(size() + steps[event.key] * grow)
  }
  return bar
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
