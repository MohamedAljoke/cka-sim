import { FitAddon } from '@xterm/addon-fit'
import { Terminal } from '@xterm/xterm'
import '@xterm/xterm/css/xterm.css'

const theme = {
  background: '#0d1219',
  foreground: '#ebe7df',
  cursor: '#f48a64',
  cursorAccent: '#0d1219',
  selectionBackground: '#5a5012',
  black: '#1f2530',
  red: '#f47c6e',
  green: '#65c98c',
  yellow: '#e6b55d',
  blue: '#7eb1f3',
  magenta: '#ce95df',
  cyan: '#7fd0d8',
  white: '#c2bdb4',
  brightBlack: '#918c81',
  brightRed: '#f6a097',
  brightGreen: '#94e2b4',
  brightYellow: '#f6d68f',
  brightBlue: '#adcbf6',
  brightMagenta: '#edb7ea',
  brightCyan: '#a3e0e6',
  brightWhite: '#ebe7df',
}

export function openTerminal(pane: HTMLElement, path: string, onDisconnect: () => void): () => void {
  const term = new Terminal({
    cursorBlink: true,
    fontFamily: "'Cascadia Mono', ui-monospace, Menlo, Consolas, monospace",
    fontSize: 14,
    theme,
  })
  const fit = new FitAddon()
  term.loadAddon(fit)
  term.open(pane)
  fit.fit()

  const scheme = location.protocol === 'https:' ? 'wss' : 'ws'
  const socket = new WebSocket(`${scheme}://${location.host}${path}`)
  socket.binaryType = 'arraybuffer'
  const isOpen = () => socket.readyState === WebSocket.OPEN
  const sendSize = () => {
    if (isOpen()) socket.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
  }

  socket.onopen = () => {
    sendSize()
    term.focus()
  }
  socket.onmessage = (event) => term.write(new Uint8Array(event.data))
  socket.onclose = () => {
    term.write('\r\n[disconnected]\r\n')
    onDisconnect()
  }

  term.attachCustomKeyEventHandler((event) => {
    const copy = event.ctrlKey && event.shiftKey && event.code === 'KeyC'
    if (!copy) return true
    event.preventDefault()
    if (event.type === 'keydown' && term.hasSelection()) navigator.clipboard.writeText(term.getSelection())
    return false
  })

  const encoder = new TextEncoder()
  term.onData((keys) => {
    if (isOpen()) socket.send(encoder.encode(keys))
  })
  term.onResize(sendSize)
  const resizes = new ResizeObserver(() => fit.fit())
  resizes.observe(pane)

  return () => {
    socket.onclose = null
    socket.close()
    resizes.disconnect()
    term.dispose()
  }
}
