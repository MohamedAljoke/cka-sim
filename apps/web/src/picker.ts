import { el } from './ui'

export type Option = { value: string; label: string; count: number }

export type Picker = {
  element: HTMLElement
  chosen(): string[]
  clear(): void
}

export function picker(noun: string, options: Option[], initial: string[], changed: () => void, signal: AbortSignal): Picker {
  const chosen = new Set(initial.filter((value) => options.some((option) => option.value === value)))
  const root = el('details', 'picker')
  const summary = el('summary', 'action picker-summary')
  const list = el('div', 'picker-options')

  const label = () => {
    if (chosen.size === 0) summary.textContent = `All ${noun}`
    else if (chosen.size === 1) summary.textContent = options.find((option) => chosen.has(option.value))!.label
    else summary.textContent = `${chosen.size} ${noun}`
    summary.classList.toggle('narrowed', chosen.size > 0)
  }

  const rows = options.map((option) => {
    const box = el('input')
    box.type = 'checkbox'
    box.checked = chosen.has(option.value)
    box.onchange = () => {
      if (box.checked) chosen.add(option.value)
      else chosen.delete(option.value)
      label()
      changed()
    }
    const row = el('label', 'choice')
    row.append(box, el('span', 'choice-label', option.label), el('span', 'muted', String(option.count)))
    return { option, box, row }
  })
  list.append(...rows.map(({ row }) => row))

  const panel = el('div', 'picker-panel')
  if (options.length > 8) {
    const search = el('input', 'picker-search')
    Object.assign(search, { type: 'search', placeholder: `Search ${noun}…` })
    search.setAttribute('aria-label', `Search ${noun}`)
    search.oninput = () => {
      const query = search.value.trim().toLowerCase()
      for (const { option, row } of rows) row.hidden = !option.label.toLowerCase().includes(query)
    }
    panel.append(search)
  }
  panel.append(list)
  root.append(summary, panel)
  label()

  document.addEventListener('click', (event) => {
    if (!root.contains(event.target as Node)) root.open = false
  }, { signal })
  root.addEventListener('keydown', (event) => {
    if (event.key !== 'Escape') return
    root.open = false
    summary.focus()
  })

  return {
    element: root,
    chosen: () => options.filter((option) => chosen.has(option.value)).map((option) => option.value),
    clear() {
      chosen.clear()
      for (const { box } of rows) box.checked = false
      label()
    },
  }
}
