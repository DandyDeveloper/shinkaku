# 進確 (ShinKaku)

A local Japanese grammar learning tool with spaced repetition (SM-2) and LLM-powered sentence grading via Ollama.

## Stack

- **Backend**: Go + chi router + SQLite (pure-Go, no CGO)
- **Frontend**: SvelteKit (static adapter, proxied to Go API)
- **SRS**: SM-2 algorithm
- **LLM grading**: Ollama (default model: `llama3.2`)

## Prerequisites

- Go 1.22+
- Node 20+
- [Ollama](https://ollama.ai) running locally (see setup below)

## Ollama Setup

1. **Install Ollama** from [ollama.ai](https://ollama.ai) or via the terminal:
   ```bash
   # macOS / Linux
   curl -fsSL https://ollama.ai/install.sh | sh
   ```

2. **Pull a model.** The default is `llama3.2`, but `qwen2.5:3b` is recommended for better Japanese language support and lower resource usage:
   ```bash
   ollama pull qwen2.5:3b
   ```

3. **Start the Ollama server** (runs on `http://localhost:11434` by default):
   ```bash
   ollama serve
   ```

4. **Configure ShinKaku** to use your chosen model via the `OLLAMA_MODEL` environment variable (see [Environment Variables](#environment-variables) below). If you pulled `qwen2.5:3b`:
   ```bash
   export OLLAMA_MODEL=qwen2.5:3b
   ```

> **Resource guide**: `qwen2.5:3b` (~2 GB), `llama3.2:3b` (~2 GB), `gemma3:4b` (~3 GB). Any of these run comfortably on a machine with 8 GB RAM and no GPU required.

## Quick Start

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Server starts on `http://localhost:8080`. SQLite DB created at `./nihongo.db`.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend dev server at `http://localhost:5173` — API calls proxy to `:8080`.

## Environment Variables

| Variable       | Default                     | Description              |
|----------------|-----------------------------|--------------------------|
| `PORT`         | `8080`                      | HTTP server port         |
| `DB_PATH`      | `./nihongo.db`              | SQLite database path     |
| `OLLAMA_URL`   | `http://localhost:11434`    | Ollama API base URL      |
| `OLLAMA_MODEL` | `llama3.2`                  | Model for grading        |

## API

| Method | Path                       | Description                        |
|--------|----------------------------|------------------------------------|
| GET    | `/api/grammar`             | List grammar points (filter `?jlpt=N3`) |
| POST   | `/api/grammar`             | Add grammar point                  |
| GET    | `/api/review/queue`        | Today's due SRS cards              |
| POST   | `/api/review/:id/grade`    | Submit grade (0–5), recalculate SM-2 |
| POST   | `/api/challenge/grade`     | LLM grades a sentence construction |

## Project Structure

```
shinkaku/
├── backend/         Go API server
│   ├── cmd/server/  Entry point
│   ├── internal/
│   │   ├── srs/     SM-2 algorithm
│   │   ├── models/  Shared structs
│   │   ├── db/      SQLite + schema
│   │   ├── api/     HTTP handlers
│   │   ├── llm/     Ollama client
│   │   └── importer/ Data importers
│   └── config/      Env config
└── frontend/        SvelteKit app
    └── src/
        ├── routes/  Pages
        └── lib/     API client + components
```
