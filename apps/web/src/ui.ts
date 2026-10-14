import { Marked, marked } from 'marked'
import type { ChatSource, Result, Task } from './api'
import { icon } from './icons'

// Questions and explanations come from the bundled catalog, not from users, so they render as is.
export function markdown(target: HTMLElement, text: string) {
  target.insertAdjacentHTML('beforeend', marked.parse(text, { async: false }))
}

// Model answers can repeat text from the learner's cluster, so raw HTML is shown as text, and a link only stays a link
// to docs we know; any other URL is shown, so a link can't quietly carry cluster data away.
const untrusted = new Marked({
  renderer: {
    html: ({ text }) => escapeHTML(text),
    image: ({ text }) => escapeHTML(text),
    link({ href, tokens }) {
      const text = this.parser.parseInline(tokens)
      if (docsLink(href)) return `<a href="${escapeHTML(href)}" target="_blank" rel="noopener noreferrer">${text}</a>`
      return text === escapeHTML(href) ? text : `${text} (${escapeHTML(href)})`
    },
  },
})

const docsHosts = ['kubernetes.io', 'helm.sh', 'github.com']

function docsLink(href: string): boolean {
  if (!URL.canParse(href)) return false
  const { protocol, hostname } = new URL(href)
  return protocol === 'https:' && docsHosts.some((host) => hostname === host || hostname.endsWith(`.${host}`))
}

export function untrustedMarkdown(target: HTMLElement, text: string) {
  target.insertAdjacentHTML('beforeend', untrusted.parse(text, { async: false }))
}

function escapeHTML(text: string): string {
  return text.replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`)
}

export function sourceList(sources: ChatSource[]): HTMLElement {
  const list = el('ul', 'source-list')
  for (const { title, url } of sources) {
    const item = el('li')
    if (docsLink(url)) {
      const a = el('a', '', title)
      a.href = url
      a.target = '_blank'
      a.rel = 'noopener noreferrer'
      item.append(a)
    } else item.textContent = `${title} (${url})`
    list.append(item)
  }
  return list
}

export function checkList(r: Result): HTMLElement {
  const box = el('div')
  box.append(el('p', 'score', `${r.earned} / ${r.total}`))
  const list = el('ul', 'checks')
  for (const c of r.checks) {
    const item = el('li', c.passed ? 'passed' : 'failed')
    item.append(el('span', 'mark', c.passed ? '✓' : '✗'), `${c.description} (${c.points} pts)`)
    list.append(item)
  }
  box.append(list)
  return box
}

export function setStatus(status: HTMLElement, kind: '' | 'ready' | 'error', text: string) {
  status.className = `status ${kind}`.trim()
  status.textContent = text
}

export function copyable(command: string): HTMLElement {
  const button = el('button', 'copy-command')
  button.type = 'button'
  button.title = 'Copy to clipboard'
  const label = el('span', 'copy-label', 'Copy')
  const showCopied = (copied: boolean) => {
    button.classList.toggle('copied', copied)
    label.replaceChildren(icon(copied ? 'check' : 'copy', 14), copied ? 'Copied' : 'Copy')
  }
  showCopied(false)
  button.append(el('code', '', command), label)

  let reset: number | undefined
  button.onclick = async () => {
    try {
      await navigator.clipboard.writeText(command)
    } catch {
      // No clipboard access (e.g. plain http on a remote host): select it for a manual copy.
      getSelection()?.selectAllChildren(button.querySelector('code')!)
      return
    }
    showCopied(true)
    clearTimeout(reset)
    reset = window.setTimeout(() => showCopied(false), 1500)
  }
  return button
}

export function hostLine(task: Task): HTMLElement {
  const host = el('div', 'host')
  host.append(el('span', 'host-label', 'Connect first'), copyable(`ssh ${task.host}`))
  return host
}

export function chips(task: Task): HTMLElement {
  const row = el('span', 'chips')
  row.append(el('span', 'chip domain', task.domain), ...task.topics.map((t) => el('span', 'chip', t)))
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
