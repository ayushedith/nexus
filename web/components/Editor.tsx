import dynamic from 'next/dynamic'
import React from 'react'

const Monaco = dynamic(() => import('@monaco-editor/react'), { ssr: false })

type Props = {
  value: string
  language?: string
  onChange?: (v: string | undefined) => void
  height?: string | number
}

export default function Editor({ value, language = 'yaml', onChange, height = '60vh' }: Props) {
  return (
    // @ts-ignore - dynamic import types
    <Monaco
      value={value}
      language={language}
      height={height}
      onChange={onChange}
      options={{ automaticLayout: true, tabSize: 2 }}
    />
  )
}
