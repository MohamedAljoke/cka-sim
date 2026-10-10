import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

// openTerminal returns a close for when the lab behind the shell ends.
export function openTerminal(pane: HTMLElement): () => void {
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
  const resizes = new ResizeObserver(() => fit.fit())
  resizes.observe(pane)

  return () => {
    socket.onclose = null
    socket.close()
    resizes.disconnect()
    term.dispose()
  }
}
