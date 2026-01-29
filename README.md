# NEXUS-API

![Nexus](assets/nexus.jpg)

**Ultra-fast, terminal-first API development & testing platform with real-time collaboration**

Git-native | Load Testing | AI-Powered | Zero Config | 100% Open Source

## Features

### 🚀 Blazing Fast
- Sub-5ms request execution overhead
- HTTP/2 and HTTP/3 (QUIC) support
- Connection pooling and reuse
- Zero-copy buffer handling

### 🎨 Beautiful Terminal UI
- Bubbletea-powered TUI with Vim keybindings
- Split-pane interface (sidebar, request builder, response viewer)
- Fuzzy search for requests and collections
- Syntax highlighting for JSON, XML, GraphQL
- Real-time response streaming

### 🔄 Git-Native Storage
- Collections stored as plain YAML/JSON
- Full git integration (commit, push, pull, branch, merge)
- Branch-based testing
- Automatic conflict resolution
- Commit history visualization

### ⚡ Load Testing Built-in
- Virtual users with ramp-up/ramp-down
- Real-time metrics: RPS, latency percentiles, errors
- Distributed load generation
- Response validation during load
- P50/P95/P99 latency tracking

### 👥 Real-Time Collaboration
- WebSocket-based multi-user editing
- Live cursor tracking
- Presence indicators
- In-app chat per collection
- Team workspaces

### 🤖 AI-Powered Features
- Generate request bodies from schema
- Auto-generate tests from OpenAPI specs
- Suggest optimizations (caching, compression)
- Convert natural language to requests
- Ollama integration (local, private)

### 🎭 Mock Server
- Create mock endpoints from examples
- Dynamic responses with templates
- Request matching rules
- Response delays for latency testing
- OpenAPI-based mock generation

### 🔧 Developer Experience
- Environment variables with encryption
- Request chaining (use response in next request)
- Pre-request scripts
- Assertions and tests
- Import from Postman, Insomnia, OpenAPI
- Export to cURL, code snippets

## Installation

```bash
go install github.com/nexusapi/nexus/cmd/nexus@latest
```

Or build from source:

```bash
git clone https://github.com/nexusapi/nexus
cd nexus
go build -o nexus ./cmd/nexus
```

## Quick Start

### 1. Create a collection

```yaml
# api.yaml
name: My API
baseUrl: https://api.example.com
environment:
  dev:
    # Nexus — terminal-first API toolkit

    Nexus is a lightweight, developer-focused toolkit for building, testing, mocking, and load-testing HTTP APIs from the terminal. It's designed to be fast, Git-native, and friendly for both single developers and teams who want a CLI-first workflow.

    Key ideas:
    - Fast request execution and low overhead
    - Collections stored alongside your code (Git-native)
    - Terminal UI for interactive exploration and quick iteration
    - Built-in mock server and simple load-testing
    - Optional AI helpers to generate request bodies and tests

    Repository: https://github.com/ayushedith/nexus

    Getting started
    -----------------

    Build from source:

    ```bash
    go build -o nexus ./cmd/nexus
    ```

    Run the TUI against a collection file:

    ```bash
    ./nexus tui examples/collections/sample.yaml
    ```

    Run a collection from the CLI (useful for CI):

    ```bash
    ./nexus run examples/collections/sample.yaml
    ```

    Start the mock server (default port 9999):

    ```bash
    ./nexus mock 9999
    ```

    Start the collaboration WebSocket server (default port 8080):

    ```bash
    ./nexus collab 8080
    ```

    AI features
    -----------

    Nexus includes adapters for AI-assisted workflows (generate request bodies, auto-generate tests, suggest optimizations). To use the hosted openAI adapter set the `OPENAI_API_KEY` environment variable, or pass `--api-key` to the `ai` subcommands.

    For example:

    ```bash
    export OPENAI_API_KEY="your-key"
    ./nexus ai generate-body schema.json
    ```

    Files and components
    ---------------------
    - `cmd/nexus` — CLI entrypoint and subcommands (`tui`, `run`, `load`, `mock`, `collab`, `ai`).
    - `pkg/collection` — collection parsing and runner (assertions, variable resolution).
    - `pkg/http` — HTTP client with connection pooling and HTTP/2 support.
    - `pkg/storage` — file-based collections and Git integration.
    - `pkg/mock` — in-process mock server for testing and local development.
    - `pkg/collab` — WebSocket-based collaboration server.
    - `pkg/ai` — AI client adapters (OpenAI, local LLMs).

    Quick tips
    ----------
    - Store collections in your repo and commit them — Nexus treats collections as first-class files.
    - Use `nexus run` in CI to validate APIs and fail the job when assertions fail.
    - Use the mock server to run integration tests against predictable responses.

    Contributing
    -------------
    Contributions are welcome — open issues or send a PR. See `CONTRIBUTING.md` for contribution guidelines.

    License
    --------
    MIT
```

### 2. Run the TUI    
### Mock Server

```bash
./nexus mock 9999
```
