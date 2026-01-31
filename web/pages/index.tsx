import React, {useEffect, useRef, useState} from 'react'
import Head from 'next/head'
import Link from 'next/link'

export default function Landing() {
  const [fontCombo, setFontCombo] = useState<'combo1'|'combo2'>('combo1')
  const revealRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    const root = document.querySelector('.landing-root')
    if (!root) return
    root.classList.remove('font-combo-1','font-combo-2')
    root.classList.add(fontCombo === 'combo1' ? 'font-combo-1' : 'font-combo-2')
  }, [fontCombo])

  useEffect(() => {
    const obs = new IntersectionObserver((entries)=>{
      entries.forEach(e=>{
        if (e.isIntersecting) (e.target as HTMLElement).classList.add('visible')
      })
    }, {threshold: 0.12})
    const nodes = revealRef.current?.querySelectorAll('.reveal') || []
    nodes.forEach(n=>obs.observe(n))
    return ()=>obs.disconnect()
  }, [])

  function copyCLI() {
    try {
      navigator.clipboard.writeText('nexus run examples/collections/demo.yaml')
      // friendly small feedback
      alert('Copied! You can paste and run that in your terminal.')
    } catch {
      alert('Could not copy. You can select and copy manually.')
    }
  }

  return (
    <div className={`landing-root ${fontCombo === 'combo1' ? 'font-combo-1' : 'font-combo-2'}`}>
      <Head>
        <title>Nexus — API Collections Runner</title>
        <meta name="description" content="Run, test and mock APIs locally — with collaboration and AI helpers." />
        <link rel="icon" href="/favicon.ico" />
      </Head>

      <div className="container">
        <div style={{display:'flex',justifyContent:'flex-end',marginBottom:8}}>
          <button className="btn ghost" onClick={()=>setFontCombo(f=>f==='combo1'?'combo2':'combo1')} aria-label="Toggle font combo">Try a different font</button>
        </div>

        <section className="hero">
          <div className="hero-inner">
            <div className="hero-copy">
              <h1>Nexus — API tooling that just works for teams</h1>
              <p className="tagline">A simple toolkit to run, mock, and share API collections from your repo. Fast local runs and private mocks so teams can move faster together.</p>

              <div className="cta">
                <a className="btn primary" href="/collections">Open the app</a>
                <a className="btn ghost" href="https://github.com/ayushedith/nexus" target="_blank" rel="noreferrer">See it on GitHub</a>
              </div>

              <div style={{display:'flex',alignItems:'center',marginTop:14}}>
                <div className="code-sample">$ nexus run examples/collections/demo.yaml</div>
                <button className="copy-btn" onClick={copyCLI}>Copy command</button>
              </div>
            </div>

            <div className="hero-image">
              <div className="illustration" aria-hidden>
                <svg width="340" height="220" viewBox="0 0 340 220" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <defs>
                    <linearGradient id="g1" x1="0" x2="1">
                      <stop offset="0%" stopColor="#7c3aed" stopOpacity="0.9" />
                      <stop offset="100%" stopColor="#5b21b6" stopOpacity="0.9" />
                    </linearGradient>
                  </defs>
                  <rect x="8" y="16" width="320" height="188" rx="12" fill="url(#g1)" opacity="0.12"/>
                  <rect x="26" y="34" width="288" height="140" rx="8" fill="#061025" />
                  <g transform="translate(40,54)" fill="#0ea5a4">
                    <rect x="0" y="0" width="220" height="10" rx="4" />
                    <rect x="0" y="22" width="180" height="10" rx="4" />
                    <rect x="0" y="44" width="260" height="10" rx="4" />
                  </g>
                  <circle cx="280" cy="50" r="18" fill="#f97316" opacity="0.95" />
                </svg>
              </div>
            </div>
          </div>
        </section>

        <div ref={revealRef}>
          <section className="features">
            <h2 className="reveal">What Nexus helps you do</h2>
            <div className="grid">
              <div className="card reveal">
                <h3>Fast local runs</h3>
                <p>Run collections from files and see responses, timings, and helpful details instantly.</p>
              </div>
              <div className="card reveal">
                <h3>Private mocks</h3>
                <p>Stand up mock endpoints for tests or to develop against when services are not available.</p>
              </div>
              <div className="card reveal">
                <h3>AI assisted help</h3>
                <p>Ask for example request bodies, quick assertions, or sensible test drafts powered by OpenAI.</p>
              </div>
            </div>
          </section>

          <section style={{padding:'30px 0'}} className="reveal">
            <h2>How it works</h2>
            <div className="grid" style={{gridTemplateColumns:'repeat(3,1fr)'}}>
              <div className="card">
                <h3>Store collections in your repo</h3>
                <p>Keep requests as YAML files so runs are reproducible and easy to review with git.</p>
              </div>
              <div className="card">
                <h3>Run locally or from the UI</h3>
                <p>Use the CLI or the web UI to run requests, inspect responses, and iterate quickly.</p>
              </div>
              <div className="card">
                <h3>Share work with your team</h3>
                <p>Collaborate on collections and attach mock endpoints for tests and demos.</p>
              </div>
            </div>
          </section>

          <section style={{padding:'30px 0'}} className="reveal">
            <h2>What people say</h2>
            <div className="grid" style={{gridTemplateColumns:'1fr',gap:12}}>
              <div className="card">
                <strong>"Nexus replaced several ad hoc scripts and saved us time right away."</strong>
                <div style={{marginTop:8,color:'var(--muted)'}}>— Dev Team at ExampleCorp</div>
              </div>
            </div>
          </section>
        </div>

        <footer className="site-footer">
          <div style={{display:'flex',flexDirection:'column',gap:8}}>
            <div style={{display:'flex',justifyContent:'space-between',alignItems:'center'}}>
              <div>© {new Date().getFullYear()} Nexus — <a href="https://github.com/ayushedith/nexus" target="_blank" rel="noreferrer">ayushedith/nexus</a></div>
              <div style={{color:'var(--muted)'}}>Local first • Git native • Privacy first</div>
            </div>
            <div style={{color:'var(--muted)',fontSize:13}}>Try the CLI: <span style={{fontFamily:'var(--font-mono)'}}>nexus run examples/collections/demo.yaml</span></div>
          </div>
        </footer>
      </div>
    </div>
  )
}
