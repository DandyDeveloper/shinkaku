# 進確 (ShinKaku)

A local Japanese grammar learning tool with spaced repetition (SM-2) and LLM-powered sentence grading via Ollama.

On first startup, the backend automatically seeds the SQLite database with a built-in starter set of grammar points and review cards, so a fresh install is immediately usable.

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

### Docker Compose

1. Copy [.env.example](/home/dandy/github.com/dandydeveloper/shinkaku/.env.example) to `.env` and fill in the required values.
2. Start the default CPU stack:
   ```bash
   docker compose up --build
   ```
3. The one-shot `ollama-init` service automatically pulls `OLLAMA_MODEL` from `.env` after Ollama is healthy.
   On first boot this can take a few minutes while the model downloads.

If the backend is being OOM-killed, raise `BACKEND_MEMORY_LIMIT` and keep `BACKEND_GO_MEMORY_LIMIT` slightly lower in `.env`. For example:

```bash
BACKEND_MEMORY_LIMIT=2g
BACKEND_GO_MEMORY_LIMIT=1536MiB
```

On Docker Desktop or WSL2, the container cannot use more memory than the VM itself has been given, so you may also need to raise Docker Desktop's global memory allocation.

### AMD GPU Setup

Use the AMD override file when the host has a ROCm-capable AMD GPU and Docker can access `/dev/kfd` and `/dev/dri`.

1. Confirm the device nodes exist:
   ```bash
   ls -l /dev/kfd /dev/dri/renderD128
   ```
2. Capture the host GIDs used by those devices and add them to `.env`:
   ```bash
   VIDEO_GID=$(getent group video | cut -d: -f3)
   RENDER_GID=$(getent group render | cut -d: -f3)
   ```
   Then set `VIDEO_GID` and `RENDER_GID` in `.env`.
3. If you are running rootless Docker, make sure your host user is also in the `video` and `render` groups before restarting Docker:
   ```bash
   sudo usermod -aG video,render "$USER"
   ```
4. Start the stack with the AMD override:
   ```bash
   docker compose -f docker-compose.yml -f docker-compose.amd.yml up --build
   ```
5. For the production compose file, use the same override:
   ```bash
   docker compose -f docker-compose.prod.yml -f docker-compose.amd.yml up --build -d
   ```

The AMD override uses numeric `group_add` entries so the container matches the host device permissions instead of assuming the image's `video` and `render` groups use the same IDs.

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
