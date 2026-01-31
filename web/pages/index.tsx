import React from 'react'
import type { GetServerSideProps } from 'next'

type Props = {
  collections: string[]
  error?: string
}

export default function Home({ collections, error }: Props) {
  if (error) return <div>Error: {error}</div>

  return (
    <main style={{ padding: 24, fontFamily: 'Inter, system-ui' }}>
      <h1>Nexus — Collections</h1>
      <p>Collections stored in repository</p>
      <ul>
        {collections.map((c) => (
          <li key={c}>{c}</li>
        ))}
      </ul>
    </main>
  )
}

export const getServerSideProps: GetServerSideProps<Props> = async (context) => {
  const backend = process.env.BACKEND_URL || 'http://localhost:8080'
  try {
    const res = await fetch(`${backend}/api/collections`)
    if (!res.ok) {
      return { props: { collections: [], error: `backend responded ${res.status}` } }
    }
    const data = await res.json()
    return { props: { collections: data.collections || [] } }
  } catch (err: any) {
    return { props: { collections: [], error: err.message } }
  }
}
