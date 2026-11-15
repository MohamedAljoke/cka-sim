import { el, resizer } from './ui'

type Tab = 'claude' | 'tutor'

export function assistPanel(claude: HTMLElement, tutor: { body: HTMLElement; focus: () => void }): HTMLElement {
  const toggle = el('button', 'tutor-summary assist-toggle')
  toggle.type = 'button'
  toggle.append(el('span', 'tutor-title', 'Assistant'))
  toggle.setAttribute('aria-label', 'Show the assistant')

  const tabs = el('div', 'assist-tabs')
  tabs.setAttribute('role', 'tablist')
  const claudeTab = tabButton('Claude Code', 'Your own Claude, with your task and cluster as its only tools')
  const tutorTab = tabButton('Ask the tutor', 'It reads your task and cluster, then helps you find the answer yourself')
  tabs.append(claudeTab, tutorTab)

  const head = el('div', 'assist-head')
  head.append(toggle, tabs)

  const views: Record<Tab, { tab: HTMLButtonElement; body: HTMLElement }> = {
    claude: { tab: claudeTab, body: claude },
    tutor: { tab: tutorTab, body: tutor.body },
  }
  const bodies = el('div', 'assist-bodies')
  for (const [name, view] of Object.entries(views)) {
    view.body.id = `assist-${name}`
    view.body.setAttribute('role', 'tabpanel')
    view.tab.setAttribute('aria-controls', view.body.id)
    bodies.append(view.body)
  }

  const panel = el('section', 'tutor-panel assist-panel')
  panel.append(
    head,
    bodies,
    resizer({
      label: 'Resize the assistant',
      axis: 'y',
      grow: -1,
      key: 'cka-sim.assist-height',
      size: () => panel.offsetHeight,
      limits: () => [200, (panel.parentElement?.clientHeight ?? 0) - 200],
      apply: (px) => panel.style.setProperty('--assist-height', `${px}px`),
    }),
  )

  let current: Tab = 'claude'
  const render = (open: boolean) => {
    panel.dataset.open = String(open)
    bodies.hidden = !open
    toggle.setAttribute('aria-expanded', String(open))
    for (const [name, view] of Object.entries(views)) {
      const selected = name === current
      view.tab.setAttribute('aria-selected', String(selected))
      view.tab.tabIndex = selected ? 0 : -1
      view.body.hidden = !selected
    }
  }
  const select = (tab: Tab) => {
    current = tab
    render(true)
    if (tab === 'tutor') tutor.focus()
  }
  render(false)

  toggle.onclick = () => render(panel.dataset.open !== 'true')
  claudeTab.onclick = () => select('claude')
  tutorTab.onclick = () => select('tutor')
  tabs.onkeydown = (event) => {
    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return
    const next: Tab = current === 'claude' ? 'tutor' : 'claude'
    select(next)
    views[next].tab.focus()
  }

  return panel
}

function tabButton(title: string, intro: string): HTMLButtonElement {
  const tab = el('button', 'assist-tab')
  tab.type = 'button'
  tab.setAttribute('role', 'tab')
  tab.title = intro
  tab.textContent = title
  return tab
}
