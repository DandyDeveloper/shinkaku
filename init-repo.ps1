# Run this once from the nihongo-sensei directory to initialize the git repo.
# Open PowerShell, cd to this folder, then: .\init-repo.ps1

Set-Location $PSScriptRoot

# Remove the corrupt .git dir left by the scaffold process
if (Test-Path ".git") {
    Remove-Item -Recurse -Force ".git"
    Write-Host "Removed stale .git directory"
}

# Initialize fresh repo
git init -b main
git add -A
git commit -m "chore: initial scaffold

- Go backend: chi router, SQLite (pure-Go), SM-2 SRS, Ollama client
- SvelteKit frontend: dashboard, review, challenge, grammar browser
- Full SM-2 algorithm (sm2.go)
- Ollama streaming client with JSON extraction (ollama.go)
- SQLite schema with WAL mode and FK enforcement
- Vite proxy: /api/* -> localhost:8080"

Write-Host ""
Write-Host "Done! Run 'git log --oneline' to verify."
