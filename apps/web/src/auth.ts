import { logIn, logOut, type User } from './api'
import { el, message, setStatus } from './ui'

export function showLogin(pane: HTMLElement, onLoggedIn: (user: User) => void) {
  const form = el('form', 'login-form')
  const email = field('Email', { type: 'email', autocomplete: 'username', required: true })
  const password = field('Password', { type: 'password', autocomplete: 'current-password', required: true })
  const submit = el('button', 'action primary', 'Log in')
  submit.type = 'submit'
  const error = el('p', 'error')
  form.append(email.label, password.label, submit, error)

  const card = el('section', 'card login')
  card.append(el('h2', '', 'Log in'), el('p', 'muted', 'Practise CKA tasks on a real cluster.'), form)
  pane.replaceChildren(card)
  email.input.focus()

  form.onsubmit = async (event) => {
    event.preventDefault()
    submit.disabled = true
    error.textContent = ''
    try {
      onLoggedIn(await logIn(email.input.value, password.input.value))
    } catch (err) {
      error.textContent = message(err)
      password.input.value = ''
      password.input.focus()
      submit.disabled = false
    }
  }
}

// A reload is the whole logout on the page: it drops the lab poll, the terminal and the view at once.
export function showAccount(bar: HTMLElement, user: User) {
  const out = el('button', 'action', 'Log out')
  out.type = 'button'
  out.onclick = async () => {
    out.disabled = true
    try {
      await logOut()
    } catch (err) {
      const status = el('span', 'status')
      setStatus(status, 'error', message(err))
      bar.replaceChildren(status, out)
      out.disabled = false
      return
    }
    location.reload()
  }
  bar.replaceChildren(el('span', 'email', user.email), out)
}

function field(text: string, attrs: Partial<HTMLInputElement>) {
  const label = el('label', '', text)
  const input = el('input')
  Object.assign(input, attrs)
  label.append(input)
  return { label, input }
}
