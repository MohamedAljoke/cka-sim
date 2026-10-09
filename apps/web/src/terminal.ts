import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

export function openTerminal(pane: HTMLElement) {
  const term = new Terminal({
    cursorBlink: true,
    fontFamily: 'ui-monospace, Menlo, Consolas, monospace',
    fontSize: 14,
    theme: { background: getComputedStyle(pane).getPropertyValue('--color-terminal').trim() },
  })
  const fit = new FitAddon()
  term.loadAddon(fit)
  term.open(pane)
  fit.fit()

  const scheme = location.protocol === 'https:' ? 'wss' : 'ws'
  const socket = new WebSocket(`${scheme}://${location.host}/ws/terminal`)
  socket.binaryType = 'arraybuffer'
  const isOpen = () => socket.readyState === WebSocket.OPEN

  const sendSize = () => {
    if (isOpen()) socket.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
  }
  const encoder = new TextEncoder()

  socket.onopen = () => {
    sendSize()
    term.focus()
  }
  socket.onmessage = (event) => term.write(new Uint8Array(event.data))
  socket.onclose = () => term.write('\r\n[disconnected]\r\n')

  term.onData((keys) => {
    if (isOpen()) socket.send(encoder.encode(keys))
  })
  term.onResize(sendSize)
  new ResizeObserver(() => fit.fit()).observe(pane)
}
