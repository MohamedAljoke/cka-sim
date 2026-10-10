import { el } from './ui'

export type Option = { value: string; label: string; count: number }

// multiSelect is a dropdown of checkboxes with everything chosen at first, so leaving it alone
// means "no filter". The trigger says how much is chosen; the panel can pick all or none.
export function multiSelect(noun: string, options: Option[], changed: () => void) {
  const chosen = new Set(options.map((o) => o.value))
  const root = el('div', 'multiselect')
  const trigger = el('button', 'select-trigger')
  trigger.type = 'button'
  trigger.setAttribute('aria-haspopup', 'true')
  const panel = el('div', 'menu')
  panel.hidden = true

  const label = () => {
    if (chosen.size === options.length) return `All ${noun} (${options.length})`
    if (chosen.size === 0) return `No ${noun}`
    if (chosen.size === 1) return options.find((o) => chosen.has(o.value))!.label
    return `${chosen.size} of ${options.length} ${noun}`
  }
  const boxes = options.map((option) => {
    const row = el('label', 'choice')
    const box = el('input')
    box.type = 'checkbox'
    box.checked = true
    box.onchange = () => {
      if (box.checked) chosen.add(option.value)
      else chosen.delete(option.value)
      sync()
    }
    row.append(box, el('span', 'choice-label', option.label), el('span', 'choice-count', String(option.count)))
    panel.append(row)
    return { option, box }
  })
  const setAll = (on: boolean) => {
    for (const { option, box } of boxes) {
      box.checked = on
      if (on) chosen.add(option.value)
      else chosen.delete(option.value)
    }
    sync()
  }
  const all = el('button', 'menu-link', 'Select all')
  const none = el('button', 'menu-link', 'Clear')
  all.type = none.type = 'button'
  all.onclick = () => setAll(true)
  none.onclick = () => setAll(false)
  const head = el('div', 'menu-head')
  const title = el('div', 'menu-title')
  title.append(el('span', 'menu-label', noun), all, none)
  head.append(title)
  // A long list gets a search box; it only hides rows, it never changes what is chosen.
  if (options.length > 8) {
    const search = el('input', 'menu-search')
    Object.assign(search, { type: 'search', placeholder: `Search ${noun}…` })
    search.oninput = () => {
      const q = search.value.trim().toLowerCase()
      for (const { option, box } of boxes) box.parentElement!.hidden = !option.label.toLowerCase().includes(q)
    }
    head.append(search)
  }
  head.append(el('div', 'menu-separator'))
  panel.prepend(head)

  const sync = () => {
    trigger.textContent = label()
    trigger.classList.toggle('narrowed', chosen.size < options.length)
    changed()
  }
  const open = (on: boolean) => {
    panel.hidden = !on
    trigger.setAttribute('aria-expanded', String(on))
  }
  trigger.onclick = () => open(panel.hidden === true)
  // Clicks inside the panel keep it open, so several boxes can be ticked in a row.
  const outside = (event: MouseEvent) => {
    if (!root.contains(event.target as Node)) open(false)
  }
  document.addEventListener('click', outside)
  root.addEventListener('keydown', (event) => {
    if (event.key === 'Escape' && !panel.hidden) {
      open(false)
      trigger.focus()
    }
  })

  root.append(trigger, panel)
  open(false)
  trigger.textContent = label()
  return {
    element: root,
    // narrowed lists what is chosen, or nothing when everything is: the server reads empty as "any".
    narrowed: () => (chosen.size === options.length ? [] : [...chosen]),
    has: (value: string) => chosen.has(value),
    all: () => chosen.size === options.length,
    dispose: () => document.removeEventListener('click', outside),
  }
}
