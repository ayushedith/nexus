export function connectWS(room: string, onMessage: (m: string) => void) {
  const backend = process.env.NEXT_PUBLIC_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:8080'
  const url = backend.replace(/^http/, 'ws') + `/ws?room=${encodeURIComponent(room)}`
  try {
    const ws = new WebSocket(url)
    ws.addEventListener('message', (ev) => onMessage(ev.data))
    ws.addEventListener('open', () => onMessage('[ws] connected'))
    ws.addEventListener('close', () => onMessage('[ws] disconnected'))
    ws.addEventListener('error', () => onMessage('[ws] error'))
    return ws
  } catch (err) {
    onMessage('[ws] connection failed')
    return null
  }
}
