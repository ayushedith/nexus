import React from 'react'
import Head from 'next/head'
import Link from 'next/link'

export default function Landing() {
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
          if (e.isIntersecting) e.target.classList.add('visible')
        })
      }, {threshold: 0.12})
      const nodes = revealRef.current?.querySelectorAll('.reveal') || []
      nodes.forEach(n=>obs.observe(n))
      return ()=>obs.disconnect()
    }, [])

    function copyCLI() {
      try {
        navigator.clipboard.writeText('nexus run examples/collections/demo.yaml')
        alert('Copied to clipboard')
      } catch {
        alert('Copy failed')
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
            <button className="btn ghost" onClick={()=>setFontCombo(f=>f==='combo1'?'combo2':'combo1')} aria-label="Toggle font combo">Switch font</button>
          </div>

          <section className="hero">
            <div className="hero-inner">
              <div className="hero-copy">
                <h1>Nexus — API tooling, reimagined</h1>
                <p className="tagline">A fast, Git-native toolkit to build, run and collaborate on API collections — locally and privately.</p>

                <div className="cta">
                  <a className="btn primary" href="/collections">Open the app</a>
                  <a className="btn ghost" href="https://github.com/ayushedith/nexus" target="_blank" rel="noreferrer">Star on GitHub</a>
                </div>

                <div style={{display:'flex',alignItems:'center',marginTop:14}}>
                  <div className="code-sample">$ nexus run examples/collections/demo.yaml</div>
                  <button className="copy-btn" onClick={copyCLI}>Copy</button>
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
              <h2 className="reveal">What Nexus gives you</h2>
              <div className="grid">
                <div className="card reveal">
                  <h3>Fast local runs</h3>
                  <p>Execute collections from files with rich results and built-in metrics and retries.</p>
                </div>
                <div className="card reveal">
                  <h3>Private mocks</h3>
                  <p>Run isolated mock servers for integration tests, CI, and local development.</p>
                </div>
                <div className="card reveal">
                  <h3>AI-assisted</h3>
                  <p>Generate request bodies, test assertions, and quick API drafts via OpenAI adapters.</p>
                </div>
              </div>
            </section>

            <section style={{padding:'30px 0'}} className="reveal">
              <h2>How it works</h2>
              <div className="grid" style={{gridTemplateColumns:'repeat(3,1fr)'}}>
                <div className="card">
                  <h3>1. Store Collections</h3>
                  <p>Keep HTTP requests as YAML files in your repo for reproducible runs and diffs.</p>
                </div>
                <div className="card">
                  <h3>2. Run Locally</h3>
                  <p>Execute from CLI or UI; view responses, timings, and assertion results instantly.</p>
                </div>
                <div className="card">
                  <h3>3. Collaborate</h3>
                  <p>Share realtime edits via WebSockets and attach mock endpoints for integration tests.</p>
                </div>
              </div>
            </section>

            <section style={{padding:'30px 0'}} className="reveal">
              <h2>What people say</h2>
              <div className="grid" style={{gridTemplateColumns:'1fr',gap:12}}>
                <div className="card">
                  <strong>"Nexus replaced several ad-hoc scripts — instant productivity win."</strong>
                  <div style={{marginTop:8,color:'var(--muted)'}}>— Dev Team at ExampleCorp</div>
                </div>
              </div>
            </section>
          </div>

          <footer className="site-footer">
            <div style={{display:'flex',flexDirection:'column',gap:8}}>
              <div style={{display:'flex',justifyContent:'space-between',alignItems:'center'}}>
                <div>© {new Date().getFullYear()} Nexus — <a href="https://github.com/ayushedith/nexus" target="_blank" rel="noreferrer">ayushedith/nexus</a></div>
                <div style={{color:'var(--muted)'}}>Local-first • Git-native • Privacy-first</div>
              </div>
              <div style={{color:'var(--muted)',fontSize:13}}>Try the CLI: <span style={{fontFamily:'var(--font-mono)'}}>nexus run examples/collections/demo.yaml</span></div>
            </div>
          </footer>
        </div>
      </div>
    )
  }
