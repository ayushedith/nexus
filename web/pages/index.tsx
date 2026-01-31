import React from 'react'
import Head from 'next/head'
import Image from 'next/image'
import Link from 'next/link'
import '../styles/landing.css'

export default function Landing() {
  return (
    <div className="landing-root">
      <header className="hero">
        <div className="hero-inner">
          <div className="hero-copy">
            <h1>Nexus — Terminal-first API toolkit</h1>
            <p className="tagline">Build, test, mock and collaborate on APIs — fast, Git-native, and privacy-first.</p>

            <div className="cta">
              <a className="btn primary" href="/collections">Open the app</a>
              <a className="btn ghost" href="https://github.com/ayushedith/nexus" target="_blank" rel="noreferrer">Star on GitHub</a>
            </div>
          </div>

          <div className="hero-image">
            <Image src="/assets/nexus.jpg" alt="Nexus" width={420} height={300} />
          </div>
        </div>
      </header>

      <Head>
        <title>Nexus — API Collections Runner</title>
        <meta name="description" content="Run, test and mock APIs locally — with collaboration and AI helpers." />
        <link rel="icon" href="/favicon.ico" />
      </Head>

      <section className="features">
        <div className="container">
          <h2>What it does</h2>
          <div className="grid">
            <div className="card">
              <h3>Terminal-first TUI</h3>
              <p>Interactive Bubbletea-based UI for building and running collections with keyboard-friendly controls.</p>
            </div>

            <div className="card">
              <h3>Git-native storage</h3>
              <p>Store collections as files in your repo, with commits and history for reproducible tests.</p>
            </div>

            <div className="card">
              <h3>Mock & Load</h3>
              <p>Run a mock server for integration tests and basic load testing out of the box.</p>
            </div>

            <div className="card">
              <h3>AI helpers</h3>
              <p>Generate request bodies and test assertions using pluggable AI adapters.</p>
            </div>
          </div>
        </div>
      </section>

      <footer className="site-footer">
        <div className="container">
          <div>Made with ❤️ — <a href="https://github.com/ayushedith/nexus" target="_blank" rel="noreferrer">ayushedith/nexus</a></div>
        </div>
      </footer>
    </div>
  )
}
