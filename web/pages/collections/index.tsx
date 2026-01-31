import React from 'react'
import Link from 'next/link'
import type { GetServerSideProps } from 'next'

type Props = { collections: string[] }

export default function Collections({ collections }: Props) {
  return (
    <main style={{ padding: 24 }}>
      <h1>Collections</h1>
      <ul>
        {collections.map((c) => (
          <li key={c}>
            <Link href={`/collections/${encodeURIComponent(c)}`}>{c}</Link>
          </li>
        ))}
      </ul>
    </main>
  )
}

export const getServerSideProps: GetServerSideProps<Props> = async () => {
  const backend = process.env.BACKEND_URL || 'http://localhost:8080'
  try {
    const res = await fetch(`${backend}/api/collections`)
    const data = await res.json()
    return { props: { collections: data.collections || [] } }
  } catch (err) {
    return { props: { collections: [] } }
  }
}
