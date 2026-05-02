<svelte:head><title>日本語先生 — Sign in</title></svelte:head>

<main>
  <div class="card">
    <h1>日本語先生</h1>
    <p class="tagline">Japanese Grammar SRS · Powered by Ollama</p>
    <a href="/auth/login" class="google-btn">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="20" height="20">
        <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/>
        <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/>
        <path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"/>
        <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.18 1.48-4.97 2.31-8.16 2.31-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/>
      </svg>
      Sign in with Google
    </a>
  </div>
</main>

<style>
  main {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    font-family: system-ui, -apple-system, sans-serif;
    background: #f7fafc;
  }
  .card {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 16px;
    padding: 3rem 2.5rem;
    text-align: center;
    max-width: 360px;
    width: 100%;
  }
  h1 { font-size: 2rem; color: #2d3748; margin-bottom: 0.25rem; }
  .tagline { color: #718096; font-size: 0.9rem; margin-bottom: 2rem; }
  .google-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.7rem 1.4rem;
    background: white;
    border: 1px solid #dadce0;
    border-radius: 8px;
    color: #3c4043;
    font-size: 0.95rem;
    font-weight: 500;
    text-decoration: none;
    transition: box-shadow 0.15s;
  }
  .google-btn:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.12); }
</style>
