# 日本語先生 (Nihongo Sensei)

A local Japanese grammar SRS tool with LLM-powered sentence grading.  
You study grammar points via spaced repetition (SM-2), then prove retention by constructing original sentences that an Ollama LLM grades in real time.

---

## Architecture

```
Browser
  └─► SvelteKit  :5173  (dev proxy)
        └─► Go API  :8080
               ├─► SQLite  (nihongo.db)
               └─► Ollama  :11434
```

All components run locally — no cloud dependencies.

---

## Google OIDC Setup

The API is protected by Google OIDC.  Access is restricted to a single whitelisted email.

### 1 — Create OAuth 2.0 credentials in Google Cloud Console

1. Open [console.cloud.google.com](https://console.cloud.google.com) → **APIs & Services → Credentials**.
2. Click **Create Credentials → OAuth client ID**.
3. Application type: **Web application**.
4. Authorised redirect URI: `http://localhost:8080/auth/callback`  
   (Change the host/port if you deploy elsewhere.)
5. Copy the **Client ID** and **Client Secret**.

### 2 — Required environment variables

| Variable               | Description                                                  |
|------------------------|--------------------------------------------------------------|
| `GOOGLE_CLIENT_ID`     | OAuth 2.0 client ID from Cloud Console                       |
| `GOOGLE_CLIENT_SECRET` | OAuth 2.0 client secret from Cloud Console                   |
| `ALLOWED_EMAIL`        | Only this Google account may log in (e.g. `you@gmail.com`)  |
| `SESSION_SECRET`       | Random 32+ byte string used to sign session cookies          |
| `BASE_URL`             | Public base URL of the API (e.g. `http://localhost:8080`)    |

Example `.env` (never commit this file):

```bash
GOOGLE_CLIENT_ID=123456789-abc.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-...
ALLOWED_EMAIL=you@gmail.com
SESSION_SECRET=change-me-to-a-random-32-byte-value
BASE_URL=http://localhost:8080
```

---

## Local LLM Setup

The challenge-grading endpoint sends sentences to Ollama.

```bash
# Install Ollama from https://ollama.ai then pull the default model:
ollama pull llama3.2
```

| Variable       | Default                  | Description                     |
|----------------|--------------------------|---------------------------------|
| `OLLAMA_URL`   | `http://localhost:11434` | Ollama API base URL             |
| `OLLAMA_MODEL` | `llama3.2`               | Model used for sentence grading |

Any Ollama model that follows instruction prompts works.  Larger models give better grammar feedback; smaller ones (e.g. `qwen2.5:3b`) are faster on CPU.

---

## Data Sources

| Source     | Description                                                | Import command                                                        |
|------------|------------------------------------------------------------|-----------------------------------------------------------------------|
| Hanabira   | Grammar points from Hanabira.org (fetched via HTTP)        | `go run ./cmd/import --source=hanabira`                               |
| Tatoeba    | Japanese example sentences from tatoeba.org                | `go run ./cmd/import --source=tatoeba --file=sentences.tsv --limit=5000` |
| JMdict     | Common vocabulary entries from the EDICT project           | `go run ./cmd/import --source=jmdict  --file=JMdict_e.gz`             |
| Custom     | Your own JSON file of grammar points                       | `go run ./cmd/import --source=custom  --file=my_grammar.json`         |

**Tatoeba** — download `sentences.tsv` from <https://tatoeba.org/en/downloads> (select Japanese).  
**JMdict** — download `JMdict_e.gz` from <https://www.edrdg.org/jmdict/edict_doc.html>.

### Custom JSON format

```json
[
  {
    "pattern":    "〜てもいい",
    "meaning":    "it is okay to ~; you may ~",
    "jlpt_level": "N5",
    "example_jp": "ここに座ってもいいですか？",
    "example_en": "May I sit here?",
    "notes":      "polite permission request",
    "tags":       "permission,grammar",
    "audio_url":  "",
    "external_id": ""
  }
]
```

All fields except `pattern` and `meaning` are optional.

---

## Getting Started

### Prerequisites

- **Go 1.22+**
- **Node 20+**
- **Ollama** running locally with `llama3.2` pulled
- A Google OAuth 2.0 client ID and secret (see [OIDC Setup](#google-oidc-setup))

### Steps

```bash
# 1. Clone
git clone https://github.com/your-org/nihongo-sensei.git
cd nihongo-sensei

# 2. Set environment variables (copy and fill in the template above)
export GOOGLE_CLIENT_ID=...
export GOOGLE_CLIENT_SECRET=...
export ALLOWED_EMAIL=...
export SESSION_SECRET=...
export BASE_URL=http://localhost:8080

# 3. Start the API server
cd backend
go run ./cmd/server
# → listening on :8080, SQLite created at ./nihongo.db

# 4. (Optional) Seed grammar data
go run ./cmd/import --source=hanabira

# 5. Start the frontend (separate terminal)
cd frontend
npm install
npm run dev
# → http://localhost:5173
```

---

## SRS — How It Works

Nihongo Sensei uses the **SM-2** algorithm (the same one behind Anki).

| Grade | Meaning          |
|-------|------------------|
| 0     | Total blackout   |
| 1     | Wrong but close  |
| 2     | Wrong, easy hint |
| 3     | Correct, hard    |
| 4     | Correct          |
| 5     | Perfect recall   |

- Grades **0–2** reset the card's repetition counter; the interval drops back to 1 day.
- Grades **3–5** increase the interval: roughly 1 day → 6 days → 15 days → ... growing by the card's *ease factor* (default 2.5, min 1.3).
- A grade of 3 nudges the ease factor slightly down; a grade of 5 nudges it up.

Cards whose `due_date ≤ now` appear in the review queue at `GET /api/review/queue`.

---

## API Reference

All `/api/*` routes require an authenticated session cookie (login via `/auth/login`).

| Method | Path                    | Auth | Description                                      |
|--------|-------------------------|------|--------------------------------------------------|
| GET    | `/health`               | No   | Liveness probe — returns `{"status":"ok"}`       |
| GET    | `/auth/login`           | No   | Redirect to Google OIDC login                    |
| GET    | `/auth/callback`        | No   | Google OIDC callback; sets session cookie        |
| GET    | `/auth/logout`          | No   | Clear session cookie                             |
| GET    | `/api/grammar`          | Yes  | List grammar points; filter with `?jlpt=N3`      |
| POST   | `/api/grammar`          | Yes  | Create a grammar point                           |
| GET    | `/api/grammar/{id}`     | Yes  | Get a single grammar point                       |
| GET    | `/api/review/queue`     | Yes  | Cards due for review today                       |
| POST   | `/api/review/{id}/grade`| Yes  | Submit SM-2 grade (0–5) for a card               |
| POST   | `/api/challenge/grade`  | Yes  | LLM grades a user-constructed sentence           |
| GET    | `/api/import/log`       | Yes  | Last 100 import runs with counts and status      |

---

## Security Notes

- **CORS**: The API sets `Access-Control-Allow-Origin: *`.  This is intentional for local development.  Restrict the origin in production.
- **No HTTPS locally**: Sessions use cookies without `Secure` flag in local mode.  Use a TLS-terminating reverse proxy (nginx, Caddy) in production and set `BASE_URL` accordingly.
- **`SESSION_SECRET` must be fixed in production**: If it changes, all existing sessions are invalidated.  Store it in a secrets manager; do not rotate it casually.
- **Single-user design**: `ALLOWED_EMAIL` allows exactly one Google account.  There is no multi-tenant RBAC.  Do not expose this server to the public internet without additional hardening.

---

## Project Structure

```
nihongo-sensei/
├── backend/
│   ├── cmd/
│   │   ├── server/       API server entry point
│   │   └── import/       Data importer CLI
│   └── internal/
│       ├── api/          HTTP handlers + router
│       ├── auth/         Google OIDC + session middleware
│       ├── db/           SQLite wrapper + schema
│       ├── importer/     Hanabira / Tatoeba / JMdict / Custom
│       ├── llm/          Ollama client
│       ├── models/       Shared Go structs
│       └── srs/          SM-2 algorithm
└── frontend/
    └── src/
        ├── routes/       SvelteKit pages
        └── lib/          API client + shared components
```
