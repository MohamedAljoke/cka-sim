import { icon } from './icons'

type Theme = 'light' | 'dark'

const root = document.documentElement
const current = (): Theme =>
  (root.dataset.theme as Theme | undefined) ?? (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')

// themeToggle flips between light and dark and remembers the choice; until then the OS decides.
export function themeToggle(button: HTMLButtonElement) {
  const render = () => {
    const dark = current() === 'dark'
    button.replaceChildren(icon(dark ? 'sun' : 'moon', 18))
    button.setAttribute('aria-label', dark ? 'Switch to light theme' : 'Switch to dark theme')
  }
  button.onclick = () => {
    const next: Theme = current() === 'dark' ? 'light' : 'dark'
    root.dataset.theme = next
    try {
      localStorage.setItem('cka-sim:color-scheme', next)
    } catch {
      // Not persisting the theme is harmless.
    }
    render()
  }
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', render)
  render()
}
