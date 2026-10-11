import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { refreshSession } from './api'

// The DSA guide's dark code panel, in hex because that is what xterm parses most reliably.
const theme = {
  background: '#0d1219', // --color-terminal
  foreground: '#ebe7df', // --ink (dark)
  cursor: '#f48a64', // --p2
  cursorAccent: '#0d1219',
  selectionBackground: '#5a5012', // --hl (dark)
  black: '#1f2530',
  red: '#f47c6e', // --bad
  green: '#65c98c', // --ok
  yellow: '#e6b55d', // --amber
  blue: '#7eb1f3', // --p1
  magenta: '#ce95df', // --p3
  cyan: '#7fd0d8',
  white: '#c2bdb4', // --ink-2
  brightBlack: '#918c81', // --ink-3
  brightRed: '#f6a097',
  brightGreen: '#94e2b4',
  brightYellow: '#f6d68f',
  brightBlue: '#adcbf6',
  brightMagenta: '#edb7ea',
  brightCyan: '#a3e0e6',
  brightWhite: '#ebe7df',
}

const mono = "14px 'JetBrains Mono Variable'"
const fallback = 'ui-monospace, Menlo, Consolas, monospace'
const fontFamily = `'JetBrains Mono Variable', ${fallback}`

// openTerminal returns a close for when the lab behind the shell ends.
export function openTerminal(pane: HTMLElement, onDisconnect: () => void): () => void {
  const term = new Terminal({
    cursorBlink: true,
    fontFamily: document.fonts.check(mono) ? fontFamily : fallback,
    fontSize: 14,
    theme,
  })
  const fit = new FitAddon()
  term.loadAddon(fit)
  term.open(pane)
  fit.fit()
  // xterm measures cells once per font, so swap to JetBrains Mono only after it has loaded.
  if (term.options.fontFamily === fallback) {
    document.fonts.load(mono).then(() => {
      term.options.fontFamily = fontFamily
      fit.fit()
    })
  }

  const scheme = location.protocol === 'https:' ? 'wss' : 'ws'
  let socket: WebSocket
  let closed = false
  const isOpen = () => socket.readyState === WebSocket.OPEN

  const sendSize = () => {
    if (isOpen()) socket.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
  }
  const encoder = new TextEncoder()

  // A refused upgrade gives no status, so a socket that never opened may just be an expired session.
  const connect = (retried: boolean) => {
    let opened = false
    socket = new WebSocket(`${scheme}://${location.host}/ws/terminal`)
    socket.binaryType = 'arraybuffer'
    socket.onopen = () => {
      opened = true
      sendSize()
      term.focus()
    }
    socket.onmessage = (event) => term.write(new Uint8Array(event.data))
    socket.onclose = async () => {
      if (!opened && !retried && (await refreshSession()) && !closed) connect(true)
      else if (!closed) {
        term.write('\r\n[disconnected]\r\n')
        onDisconnect()
      }
    }
  }
  connect(false)

  term.onData((keys) => {
    if (isOpen()) socket.send(encoder.encode(keys))
  })
  term.onResize(sendSize)
  const resizes = new ResizeObserver(() => fit.fit())
  resizes.observe(pane)

  return () => {
    closed = true
    socket.onclose = null
    socket.close()
    resizes.disconnect()
    term.dispose()
  }
}
