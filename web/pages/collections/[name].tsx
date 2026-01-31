import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/router'
import { connectWS } from '../../lib/ws'
import Editor from '../../components/Editor'

export default function CollectionPage() {
  const router = useRouter()
  const name = Array.isArray(router.query.name) ? router.query.name[0] : router.query.name || ''

  const [content, setContent] = useState('')
  const [status, setStatus] = useState<string | null>(null)
  const [wsLog, setWsLog] = useState<string[]>([])
  const [ws, setWs] = useState<WebSocket | null>(null)

  useEffect(() => {
    if (!name) return
    const backend = process.env.NEXT_PUBLIC_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:8080'
    fetch(`${backend}/api/collections/get?name=${encodeURIComponent(name)}`)
      .then((r) => r.json())
      .then((data) => {
        setContent(JSON.stringify(data, null, 2))
      })
      .catch((e) => setStatus(String(e)))

    const socket = connectWS(name, (m) => setWsLog((s) => [...s, m]))
    setWs(socket)
    return () => {
      try {
        socket?.close()
      } catch {}
    }
  }, [name])

  function save() {
    const backend = process.env.NEXT_PUBLIC_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:8080'
    fetch(`${backend}/api/collections/save`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, content }),
    })
      .then((r) => r.json())
      .then(() => setStatus('saved'))
      .catch((e) => setStatus(String(e)))
  }

  function run() {
    const backend = process.env.NEXT_PUBLIC_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:8080'
    setStatus('running')
    fetch(`${backend}/api/run`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    })
      .then((r) => r.json())
      .then((data) => setStatus(JSON.stringify(data)))
      .catch((e) => setStatus(String(e)))
  }

  return (
    <main style={{ padding: 24 }}>
      <h1>Collection: {name}</h1>
      <div style={{ marginBottom: 8 }}>
        <button onClick={save} style={{ marginRight: 8 }}>Save</button>
        <button onClick={run}>Run</button>
      </div>

      <Editor value={content} onChange={(v) => setContent(v ?? '')} height="60vh" />

      <div style={{ marginTop: 12 }}>
        <strong>Status:</strong> {status}
      </div>

      <div style={{ marginTop: 12 }}>
        <strong>WS log:</strong>
        <div style={{ maxHeight: 200, overflow: 'auto', border: '1px solid #eee', padding: 8 }}>
          {wsLog.map((l, i) => (
            <div key={i}>{l}</div>
          ))}
        </div>
      </div>
    </main>
  )
}
